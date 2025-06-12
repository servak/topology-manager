package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/repository/postgres"
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
	// PostgreSQL DSN を環境変数から取得
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://topology:topology@localhost/topology_manager?sslmode=disable"
		if verbose {
			log.Printf("Using default database URL (set DATABASE_URL to override)")
		}
	}

	// PostgreSQLリポジトリの初期化
	repo, err := postgres.NewPostgresRepository(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
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

func clearData(ctx context.Context, repo *postgres.PostgresRepository) error {
	// リンクを先に削除（外部キー制約のため）
	if _, err := repo.DB().ExecContext(ctx, "DELETE FROM links"); err != nil {
		return fmt.Errorf("failed to clear links: %w", err)
	}
	
	if _, err := repo.DB().ExecContext(ctx, "DELETE FROM devices"); err != nil {
		return fmt.Errorf("failed to clear devices: %w", err)
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
			
			// IPアドレス生成 (各オクテットが255を超えないように調整)
			subnet := (i / 254) + 1
			host := (i % 254) + 1
			ipAddress := fmt.Sprintf("10.%d.%d.%d", deviceType.layer, subnet, host)
			
			device := topology.Device{
				ID:        deviceID,
				Name:      deviceID,
				Type:      deviceType.typeName,
				Hardware:  deviceType.hardware,
				Instance:  fmt.Sprintf("dc1.%s", deviceType.prefix),
				IPAddress: ipAddress,
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

	// Core ↔ Distribution
	for _, coreDevice := range coreDevices {
		for i, distDevice := range distDevices {
			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   coreDevice.ID,
				TargetID:   distDevice.ID,
				SourcePort: fmt.Sprintf("Ethernet%d", i+1),
				TargetPort: "Ethernet49",
				Weight:     1.0,
				Status:     "up",
				Metadata: map[string]string{
					"link_type": "core-distribution",
					"speed":     "100G",
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			links = append(links, link)
			linkIndex++
		}
	}

	// Distribution ↔ Access
	for i, distDevice := range distDevices {
		startIdx := i * (len(accessDevices) / len(distDevices))
		endIdx := (i + 1) * (len(accessDevices) / len(distDevices))
		if i == len(distDevices)-1 {
			endIdx = len(accessDevices)
		}

		for j := startIdx; j < endIdx && j < len(accessDevices); j++ {
			accessDevice := accessDevices[j]
			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   distDevice.ID,
				TargetID:   accessDevice.ID,
				SourcePort: fmt.Sprintf("Ethernet%d", j-startIdx+1),
				TargetPort: "Ethernet49",
				Weight:     2.0,
				Status:     "up",
				Metadata: map[string]string{
					"link_type": "distribution-access",
					"speed":     "10G",
				},
				LastSeen:  now,
				CreatedAt: now,
				UpdatedAt: now,
			}
			links = append(links, link)
			linkIndex++
		}
	}

	// Access ↔ Server
	for i, accessDevice := range accessDevices {
		startIdx := i * (len(serverDevices) / len(accessDevices))
		endIdx := (i + 1) * (len(serverDevices) / len(accessDevices))
		if i == len(accessDevices)-1 {
			endIdx = len(serverDevices)
		}

		for j := startIdx; j < endIdx && j < len(serverDevices); j++ {
			serverDevice := serverDevices[j]
			linkID := fmt.Sprintf("link-%03d", linkIndex)
			link := topology.Link{
				ID:         linkID,
				SourceID:   accessDevice.ID,
				TargetID:   serverDevice.ID,
				SourcePort: fmt.Sprintf("Ethernet%d", j-startIdx+1),
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