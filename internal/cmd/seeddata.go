package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/internal/service"
	"github.com/spf13/cobra"
)

var (
	deviceCount            int
	clearFirst             bool
	enableAutoClassifySeed bool
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
	seedDataCmd.Flags().BoolVar(&enableAutoClassifySeed, "enable-auto-classify", true, "Enable automatic device classification for seed data")
}

func runSeedData(cmd *cobra.Command, args []string) {
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	repo, err := repository.NewRepository(config.Database)
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
		if err := repo.Clear(); err != nil {
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

	// 自動分類の実行
	if enableAutoClassifySeed {
		log.Println("Applying auto-classification to seed devices...")

		// PostgreSQL specific implementation for classification repository
		classificationService := service.NewClassificationService(repo, repo)

		// Extract device IDs
		deviceIDs := make([]string, len(devices))
		for i, device := range devices {
			deviceIDs[i] = device.ID
		}

		// Apply classification rules
		classifications, err := classificationService.ApplyClassificationRules(ctx, deviceIDs)
		if err != nil {
			log.Printf("Auto-classification failed: %v", err)
		} else {
			if len(classifications) > 0 {
				log.Printf("Successfully auto-classified %d devices:", len(classifications))
				for _, c := range classifications {
					log.Printf("  - %s → Layer %d (%s)", c.DeviceID, c.Layer, c.DeviceType)
				}
			} else {
				log.Printf("No devices matched existing classification rules (this is normal for initial setup)")
				log.Printf("You can create classification rules in the web interface and then re-run with auto-classification")
			}
		}
	}

	log.Println("Sample data generation completed successfully")
}

func generateSampleData(count int) ([]topology.Device, []topology.Link) {
	now := time.Now()
	devices := make([]topology.Device, 0, count)
	links := make([]topology.Link, 0, count*2)

	// 実際のデータセンター階層に合わせたサンプルデータ
	// デフォルト分類ルールとマッチするように命名規則を調整
	deviceTypes := []struct {
		prefix   string
		typeName string
		layer    int
		hardware string
	}{
		// Border Layer
		{"border", "unknown", 99, "Cisco ASR 9000"},
		{"edge", "unknown", 99, "Arista 7280R"},
		// Spine Layer
		{"spine", "unknown", 99, "Arista 7280R"},
		{"core", "unknown", 99, "Nexus 9000"},
		// Leaf Layer
		{"leaf", "unknown", 99, "Arista 7320X"},
		{"tor", "unknown", 99, "Nexus 9300"},
		// Servers
		{"server", "unknown", 99, "Dell PowerEdge"},
		{"srv", "unknown", 99, "HP ProLiant"},
		// Storage
		{"storage", "unknown", 99, "NetApp FAS"},
		{"san", "unknown", 99, "Pure Storage"},
	}

	deviceIndex := 0
	linkIndex := 0

	// 各タイプのデバイスを均等に生成
	for _, deviceType := range deviceTypes {
		devicesPerType := count / len(deviceTypes)
		if devicesPerType == 0 {
			devicesPerType = 1
		}

		if deviceIndex >= count {
			break
		}

		for i := 0; i < devicesPerType && deviceIndex < count; i++ {
			deviceID := fmt.Sprintf("%s-%03d", deviceType.prefix, i+1)

			device := topology.Device{
				ID:           deviceID,
				Type:         deviceType.typeName,
				Hardware:     deviceType.hardware,
				LayerID:      nil, // will be set by classification
				DeviceType:   "",  // will be set by classification
				ClassifiedBy: "",  // empty string will be handled as NULL in database
				Metadata: map[string]string{
					"datacenter": "dc1",
					"rack":       fmt.Sprintf("rack-%d", (i/10)+1),
					"role":       deviceType.typeName,
					"generated":  "seed",
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
		if device.LayerID != nil && *device.LayerID == layer {
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
