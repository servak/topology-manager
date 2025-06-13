package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/internal/repository/postgres"
	"github.com/spf13/cobra"
)

var (
	deviceCount int
	clearFirst  bool
)

var seedDataCmd = &cobra.Command{
	Use:   "seed",
	Short: "Generate sample data for testing",
	Long:  "Generate sample network topology data for testing and development",
	Run:   runSeedData,
}

func init() {
	seedDataCmd.Flags().IntVarP(&deviceCount, "count", "n", 10, "Number of devices to generate")
	seedDataCmd.Flags().BoolVarP(&clearFirst, "clear", "", false, "Clear existing data before seeding")
}

func runSeedData(cmd *cobra.Command, args []string) {
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	repo, err := repository.NewDatabase(&config.Database)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer repo.Close()

	if verbose {
		log.Println("Connected to PostgreSQL")
	}

	ctx := context.Background()

	// 既存データのクリア（オプション）
	if clearFirst {
		log.Println("Clearing existing data...")
		if err := clearData(ctx, repo); err != nil {
			log.Fatalf("Failed to clear data: %v", err)
		}
		log.Println("Existing data cleared")
	}

	// サンプルデータの生成
	log.Printf("Generating %d devices with links...", deviceCount)

	devices, links := generateSampleData(deviceCount)

	// デバイスの一括追加
	if err := repo.BulkAddDevices(ctx, devices); err != nil {
		log.Fatalf("Failed to add devices: %v", err)
	}
	log.Printf("Added %d devices", len(devices))

	// リンクの一括追加
	if err := repo.BulkAddLinks(ctx, links); err != nil {
		log.Fatalf("Failed to add links: %v", err)
	}
	log.Printf("Added %d links", len(links))

	log.Println("Sample data generation completed successfully")
}

func clearData(ctx context.Context, repo topology.Repository) error {
	// PostgreSQL専用の実装
	if pgRepo, ok := repo.(*postgres.PostgresRepository); ok {
		// リンクを先に削除（外部キー制約のため）
		if _, err := pgRepo.DB().ExecContext(ctx, "DELETE FROM links"); err != nil {
			return fmt.Errorf("failed to clear links: %w", err)
		}

		if _, err := pgRepo.DB().ExecContext(ctx, "DELETE FROM devices"); err != nil {
			return fmt.Errorf("failed to clear devices: %w", err)
		}
	} else {
		return fmt.Errorf("clear data is only supported for PostgreSQL repositories")
	}

	return nil
}

func generateSampleData(count int) ([]topology.Device, []topology.Link) {
	now := time.Now()
	devices := make([]topology.Device, 0, count)
	links := make([]topology.Link, 0, count*2)

	// コア・ディストリビューション・アクセス階層のサンプルデータ
	deviceTypes := []struct {
		prefix   string
		typeName string
		layer    int
		hardware string
	}{
		{"core", "core", 1, "Arista 7280R"},
		{"dist", "distribution", 2, "Arista 7050X"},
		{"access", "access", 3, "Arista 7048T"},
		{"server", "server", 4, "Dell PowerEdge"},
	}

	deviceIndex := 0
	linkIndex := 0

	// 各階層のデバイスを生成
	for _, deviceType := range deviceTypes {
		var deviceCountForType int
		switch deviceType.layer {
		case 1: // core
			deviceCountForType = max(1, count/10)
		case 2: // distribution
			deviceCountForType = max(2, count/5)
		case 3: // access
			deviceCountForType = max(3, count/2)
		case 4: // server
			deviceCountForType = count - deviceIndex
		}

		if deviceIndex >= count {
			break
		}

		for i := 0; i < deviceCountForType && deviceIndex < count; i++ {
			deviceID := fmt.Sprintf("%s-%03d", deviceType.prefix, i+1)

			device := topology.Device{
				ID:       deviceID,
				Type:     deviceType.typeName,
				Hardware: deviceType.hardware,
				Instance: fmt.Sprintf("dc1.%s", deviceType.prefix),
				Location:  fmt.Sprintf("Rack-%d", (i/10)+1),
				Status:    "up",
				Layer:     deviceType.layer,
				Metadata: map[string]string{
					"datacenter": "dc1",
					"rack":       fmt.Sprintf("rack-%d", (i/10)+1),
					"role":       deviceType.typeName,
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			devices = append(devices, device)
			deviceIndex++
		}
	}

	// リンクの生成（階層間接続）
	coreDevices := filterDevicesByLayer(devices, 1)
	distDevices := filterDevicesByLayer(devices, 2)
	accessDevices := filterDevicesByLayer(devices, 3)
	serverDevices := filterDevicesByLayer(devices, 4)

	// Core ↔ Distribution (現実的な接続パターン)
	// 各Distributionは2つのCoreに冗長接続、各Coreは最大32ポート使用
	maxCoreConnections := 32
	coresPerDist := 2

	for i, distDevice := range distDevices {
		connectionsCount := 0

		// 各DistributionはプライマリとセカンダリのCoreに接続
		primaryCoreIndex := i % len(coreDevices)
		secondaryCoreIndex := (i + 1) % len(coreDevices)

		// プライマリCore接続
		if connectionsCount < coresPerDist && primaryCoreIndex < len(coreDevices) {
			coreDevice := coreDevices[primaryCoreIndex]
			corePortNum := (i % maxCoreConnections) + 1

			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   coreDevice.ID,
				TargetID:   distDevice.ID,
				SourcePort: fmt.Sprintf("Ethernet%d", corePortNum),
				TargetPort: "Ethernet49",
				Weight:     1.0,
				Status:     "up",
				Metadata: map[string]string{
					"link_type":  "core-distribution",
					"speed":      "100G",
					"redundancy": "primary",
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			links = append(links, link)
			linkIndex++
			connectionsCount++
		}

		// セカンダリCore接続（冗長化）
		if connectionsCount < coresPerDist && secondaryCoreIndex < len(coreDevices) && secondaryCoreIndex != primaryCoreIndex {
			coreDevice := coreDevices[secondaryCoreIndex]
			corePortNum := (i % maxCoreConnections) + 1

			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   coreDevice.ID,
				TargetID:   distDevice.ID,
				SourcePort: fmt.Sprintf("Ethernet%d", corePortNum),
				TargetPort: "Ethernet50",
				Weight:     1.0,
				Status:     "up",
				Metadata: map[string]string{
					"link_type":  "core-distribution",
					"speed":      "100G",
					"redundancy": "secondary",
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			links = append(links, link)
			linkIndex++
		}
	}

	// Distribution ↔ Access (現実的な接続パターン)
	// 各Accessは2つのDistributionに冗長接続、Distributionは最大24ポートをアクセス用に使用
	maxDistAccessPorts := 24
	distsPerAccess := 2

	for i, accessDevice := range accessDevices {
		connectionsCount := 0

		// 各AccessはプライマリとセカンダリのDistributionに接続
		primaryDistIndex := i % len(distDevices)
		secondaryDistIndex := (i + 1) % len(distDevices)

		// プライマリDistribution接続
		if connectionsCount < distsPerAccess && primaryDistIndex < len(distDevices) {
			distDevice := distDevices[primaryDistIndex]
			distPortNum := (i % maxDistAccessPorts) + 1

			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   distDevice.ID,
				TargetID:   accessDevice.ID,
				SourcePort: fmt.Sprintf("Ethernet%d", distPortNum),
				TargetPort: "Ethernet49",
				Weight:     2.0,
				Status:     "up",
				Metadata: map[string]string{
					"link_type":  "distribution-access",
					"speed":      "10G",
					"redundancy": "primary",
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			links = append(links, link)
			linkIndex++
			connectionsCount++
		}

		// セカンダリDistribution接続（冗長化）
		if connectionsCount < distsPerAccess && secondaryDistIndex < len(distDevices) && secondaryDistIndex != primaryDistIndex {
			distDevice := distDevices[secondaryDistIndex]
			distPortNum := (i % maxDistAccessPorts) + 1

			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   distDevice.ID,
				TargetID:   accessDevice.ID,
				SourcePort: fmt.Sprintf("Ethernet%d", distPortNum),
				TargetPort: "Ethernet50",
				Weight:     2.0,
				Status:     "up",
				Metadata: map[string]string{
					"link_type":  "distribution-access",
					"speed":      "10G",
					"redundancy": "secondary",
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			links = append(links, link)
			linkIndex++
		}
	}

	// Access ↔ Server (現実的な接続パターン)
	// 各Accessスイッチは最大24台のサーバーに接続
	maxAccessServerPorts := 24

	for i, accessDevice := range accessDevices {
		startIdx := i * maxAccessServerPorts
		endIdx := startIdx + maxAccessServerPorts
		if endIdx > len(serverDevices) {
			endIdx = len(serverDevices)
		}

		for j := startIdx; j < endIdx && j < len(serverDevices); j++ {
			serverDevice := serverDevices[j]
			accessPortNum := (j - startIdx) + 1

			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   accessDevice.ID,
				TargetID:   serverDevice.ID,
				SourcePort: fmt.Sprintf("Ethernet%d", accessPortNum),
				TargetPort: "eth0",
				Weight:     3.0,
				Status:     "up",
				Metadata: map[string]string{
					"link_type": "access-server",
					"speed":     "1G",
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			links = append(links, link)
			linkIndex++
		}
	}

	// 冗長リンクの追加（一部のデバイス間）
	if len(coreDevices) > 1 {
		for i := 0; i < len(coreDevices)-1; i++ {
			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   coreDevices[i].ID,
				TargetID:   coreDevices[i+1].ID,
				SourcePort: "Ethernet50",
				TargetPort: "Ethernet50",
				Weight:     1.0,
				Status:     "up",
				Metadata: map[string]string{
					"link_type": "core-core",
					"speed":     "100G",
					"redundant": "true",
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			links = append(links, link)
			linkIndex++
		}
	}

	return devices, links
}

func filterDevicesByLayer(devices []topology.Device, layer int) []topology.Device {
	var filtered []topology.Device
	for _, device := range devices {
		if device.Layer == layer {
			filtered = append(filtered, device)
		}
	}
	return filtered
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
