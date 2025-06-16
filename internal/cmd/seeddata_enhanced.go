package cmd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/internal/service"
	"github.com/spf13/cobra"
)

var (
	// Enhanced parameters
	topologyType               string // three-tier, spine-leaf, fat-tree, mixed
	fatTreeScale               float64
	spineLeafScale             float64
	threeTierScale             float64
	targetDevices              int
	dcLocation                 string
	locationDelimiter          string
	includeServers             bool
	enableAutoClassifyEnhanced bool
)

var seedDataEnhancedCmd = &cobra.Command{
	Use:   "seed-enhanced",
	Short: "Generate enhanced sample data with realistic topologies",
	Long: `Generate sample network topology data with multiple topology patterns:
- Three-Tier (legacy): Core -> Aggregation -> Access
- Spine-Leaf (modern): Spine <-> Leaf with ECMP
- Fat-Tree (latest): Core-Spine -> Agg-Spine -> Edge-Leaf
- Mixed: Combination of all topologies in a single DC`,
	Run: runSeedDataEnhanced,
}

func init() {
	seedDataEnhancedCmd.Flags().StringVarP(&topologyType, "topology", "t", "mixed", "Topology type (three-tier, spine-leaf, fat-tree, mixed)")
	seedDataEnhancedCmd.Flags().Float64Var(&fatTreeScale, "fat-tree-scale", 0.3, "Fat-Tree topology scale factor")
	seedDataEnhancedCmd.Flags().Float64Var(&spineLeafScale, "spine-leaf-scale", 0.4, "Spine-Leaf topology scale factor")
	seedDataEnhancedCmd.Flags().Float64Var(&threeTierScale, "three-tier-scale", 0.3, "Three-Tier topology scale factor")
	seedDataEnhancedCmd.Flags().IntVarP(&targetDevices, "target-devices", "n", 100, "Target number of network infrastructure devices (excludes servers)")
	seedDataEnhancedCmd.Flags().StringVar(&dcLocation, "dc-location", "", "Datacenter location suffix (e.g., 'TYO', 'NYC')")
	seedDataEnhancedCmd.Flags().StringVar(&locationDelimiter, "location-delimiter", ".", "Delimiter for location suffix")
	seedDataEnhancedCmd.Flags().BoolVar(&includeServers, "include-servers", true, "Include server devices")
	seedDataEnhancedCmd.Flags().BoolVarP(&clearFirst, "clear", "", false, "Clear existing data before seeding")
	seedDataEnhancedCmd.Flags().BoolVar(&enableAutoClassifyEnhanced, "enable-auto-classify", true, "Enable automatic device classification for enhanced seed data")
}

type topologyGenerator struct {
	deviceCounter int
	linkCounter   int
	dcSuffix      string
	delimiter     string
	now           time.Time
}

func newTopologyGenerator(dcLocation, delimiter string) *topologyGenerator {
	suffix := ""
	if dcLocation != "" {
		suffix = strings.ToUpper(dcLocation)
	}

	return &topologyGenerator{
		deviceCounter: 0,
		linkCounter:   0,
		dcSuffix:      suffix,
		delimiter:     delimiter,
		now:           time.Now(),
	}
}

func (g *topologyGenerator) generateDeviceID(prefix string) string {
	g.deviceCounter++
	// タイムスタンプの下4桁を含めてユニーク性を保証
	baseID := fmt.Sprintf("%s-%04d-%04d", prefix, int(g.now.Unix()%10000), g.deviceCounter)
	if g.dcSuffix != "" {
		return fmt.Sprintf("%s%s%s", baseID, g.delimiter, g.dcSuffix)
	}
	return baseID
}

func (g *topologyGenerator) generateLinkID() string {
	g.linkCounter++
	// タイムスタンプを含めてユニーク性を保証
	return fmt.Sprintf("link-%d-%06d", g.now.Unix(), g.linkCounter)
}

func (g *topologyGenerator) createDevice(id, deviceType, hardware string, layer int) topology.Device {
	layerID := layer
	return topology.Device{
		ID:           id,
		Type:         deviceType,
		Hardware:     hardware,
		LayerID:      &layerID,
		DeviceType:   "", // will be set by classification
		ClassifiedBy: "", // will be set by classification
		Metadata: map[string]string{
			"datacenter": "dc1",
			"rack":       fmt.Sprintf("rack-%d", (g.deviceCounter/10)+1),
			"role":       deviceType,
		},
		LastSeen:  g.now,
		CreatedAt: g.now,
		UpdatedAt: g.now,
	}
}

func (g *topologyGenerator) createLink(sourceID, targetID, sourcePort, targetPort, linkType, speed string, weight float64) topology.Link {
	return topology.Link{
		ID:         g.generateLinkID(),
		SourceID:   sourceID,
		TargetID:   targetID,
		SourcePort: sourcePort,
		TargetPort: targetPort,
		Weight:     weight,
		Metadata: map[string]string{
			"link_type": linkType,
			"speed":     speed,
		},
		LastSeen:  g.now,
		CreatedAt: g.now,
		UpdatedAt: g.now,
	}
}

// Three-Tier Topology Generator
func (g *topologyGenerator) generateThreeTierTopology(numCore, numAggPerCore, numAccessPerAgg int, includeServers bool) ([]topology.Device, []topology.Link) {
	var devices []topology.Device
	var links []topology.Link

	// Core Layer
	var coreDevices []topology.Device
	for i := 0; i < numCore; i++ {
		device := g.createDevice(
			g.generateDeviceID("core"),
			"core",
			"Arista DCS-7280SR-48C6",
			42, // Aggregation layer for 3-tier
		)
		devices = append(devices, device)
		coreDevices = append(coreDevices, device)
	}

	// Aggregation Layer
	var aggDevices []topology.Device
	for _, coreDevice := range coreDevices {
		for i := 0; i < numAggPerCore; i++ {
			device := g.createDevice(
				g.generateDeviceID("agg"),
				"aggregation",
				"Juniper QFX5100-48S",
				42, // Aggregation layer for 3-tier
			)
			devices = append(devices, device)
			aggDevices = append(aggDevices, device)

			// Agg to Core link
			link := g.createLink(
				device.ID, coreDevice.ID,
				"et-0/0/47", fmt.Sprintf("Ethernet%d", i+1),
				"L3_routed", "100G", 1.0,
			)
			links = append(links, link)
		}
	}

	// Access Layer
	var accessDevices []topology.Device
	for _, aggDevice := range aggDevices {
		for i := 0; i < numAccessPerAgg; i++ {
			device := g.createDevice(
				g.generateDeviceID("access"),
				"access",
				"Cisco Catalyst 2960X-48TS",
				43, // Access layer for 3-tier
			)
			devices = append(devices, device)
			accessDevices = append(accessDevices, device)

			// Access to Agg link
			link := g.createLink(
				device.ID, aggDevice.ID,
				"GigabitEthernet0/49", fmt.Sprintf("et-0/0/%d", i+1),
				"L2_trunk", "10G", 2.0,
			)
			links = append(links, link)
		}
	}

	// Servers
	if includeServers {
		for _, accessDevice := range accessDevices {
			serverCount := rand.Intn(16) + 8 // 8-23 servers per access switch
			for i := 0; i < serverCount; i++ {
				device := g.createDevice(
					g.generateDeviceID("server"),
					"server",
					"Dell PowerEdge R640",
					50, // Server layer
				)
				devices = append(devices, device)

				// Server to Access link
				link := g.createLink(
					accessDevice.ID, device.ID,
					fmt.Sprintf("GigabitEthernet0/%d", i+1), "eth0",
					"L2_access", "1G", 3.0,
				)
				links = append(links, link)
			}
		}
	}

	return devices, links
}

// Spine-Leaf Topology Generator
func (g *topologyGenerator) generateSpineLeafTopology(numSpines, numLeavesPerSpine int, includeServers bool) ([]topology.Device, []topology.Link) {
	var devices []topology.Device
	var links []topology.Link

	// Spine Layer
	var spineDevices []topology.Device
	for i := 0; i < numSpines; i++ {
		device := g.createDevice(
			g.generateDeviceID("spine"),
			"spine",
			"Mellanox SN3700C",
			32, // Spine layer for Spine-Leaf
		)
		devices = append(devices, device)
		spineDevices = append(spineDevices, device)
	}

	// Leaf Layer
	var leafDevices []topology.Device
	numLeaves := numLeavesPerSpine * numSpines
	if numLeaves == 0 && numSpines > 0 {
		numLeaves = 1
	}

	for i := 0; i < numLeaves; i++ {
		device := g.createDevice(
			g.generateDeviceID("leaf"),
			"leaf",
			"Mellanox SN2700",
			41, // Leaf layer for Spine-Leaf
		)
		devices = append(devices, device)
		leafDevices = append(leafDevices, device)

		// Each leaf connects to all spines (ECMP)
		for j, spineDevice := range spineDevices {
			link := g.createLink(
				device.ID, spineDevice.ID,
				fmt.Sprintf("swp%d", j+1), fmt.Sprintf("swp%d", i+1),
				"L3_routed_ECMP", "100G", 1.0,
			)
			links = append(links, link)
		}
	}

	// Servers
	if includeServers {
		for _, leafDevice := range leafDevices {
			serverCount := rand.Intn(28) + 20 // 20-47 servers per leaf
			for i := 0; i < serverCount; i++ {
				device := g.createDevice(
					g.generateDeviceID("server"),
					"server",
					"HPE ProLiant DL380",
					50, // Server layer
				)
				devices = append(devices, device)

				// Server to Leaf link
				link := g.createLink(
					leafDevice.ID, device.ID,
					fmt.Sprintf("swp%d", i+10), "ens1f0",
					"L2_access_VXLAN", "25G", 2.0,
				)
				links = append(links, link)
			}
		}
	}

	return devices, links
}

// Fat-Tree Topology Generator
func (g *topologyGenerator) generateFatTreeTopology(coreSpines, aggSpinesPerCore, edgeLeavesPerAgg int, includeServers bool) ([]topology.Device, []topology.Link) {
	var devices []topology.Device
	var links []topology.Link

	// Core Spine Layer
	var coreSpineDevices []topology.Device
	for i := 0; i < coreSpines; i++ {
		device := g.createDevice(
			g.generateDeviceID("cs"),
			"core_spine",
			"Broadcom Tomahawk 4",
			30, // Core Spine layer for Fat-Tree
		)
		devices = append(devices, device)
		coreSpineDevices = append(coreSpineDevices, device)
	}

	// Aggregation Spine Layer
	var aggSpineDevices []topology.Device
	for _, coreSpineDevice := range coreSpineDevices {
		for i := 0; i < aggSpinesPerCore; i++ {
			device := g.createDevice(
				g.generateDeviceID("as"),
				"agg_spine",
				"Broadcom Trident 4",
				31, // Aggregation Spine layer for Fat-Tree
			)
			devices = append(devices, device)
			aggSpineDevices = append(aggSpineDevices, device)

			// Agg Spine to Core Spine link
			link := g.createLink(
				device.ID, coreSpineDevice.ID,
				fmt.Sprintf("Ethernet%d", i*4+1), fmt.Sprintf("Ethernet%d", len(aggSpineDevices)),
				"L3_routed_ECMP", "400G", 1.0,
			)
			links = append(links, link)
		}
	}

	// Edge Leaf Layer
	var edgeLeafDevices []topology.Device
	for _, aggSpineDevice := range aggSpineDevices {
		for i := 0; i < edgeLeavesPerAgg; i++ {
			device := g.createDevice(
				g.generateDeviceID("el"),
				"edge_leaf",
				"Broadcom Trident 3",
				40, // Edge/Leaf layer for Fat-Tree
			)
			devices = append(devices, device)
			edgeLeafDevices = append(edgeLeafDevices, device)

			// Edge Leaf to Agg Spine link
			link := g.createLink(
				device.ID, aggSpineDevice.ID,
				fmt.Sprintf("Ethernet%d", i*2+1), fmt.Sprintf("Ethernet%d", len(edgeLeafDevices)),
				"L3_routed_ECMP_L2_VLAN_overlay", "200G", 1.0,
			)
			links = append(links, link)
		}
	}

	// Servers
	if includeServers {
		for _, edgeLeafDevice := range edgeLeafDevices {
			serverCount := rand.Intn(31) + 30 // 30-60 servers per edge leaf
			for i := 0; i < serverCount; i++ {
				device := g.createDevice(
					g.generateDeviceID("server"),
					"server",
					"Supermicro SYS-2029U-TN24R4T",
					50, // Server layer
				)
				devices = append(devices, device)

				// Server to Edge Leaf link
				link := g.createLink(
					edgeLeafDevice.ID, device.ID,
					fmt.Sprintf("Ethernet%d", i+10), "ens2f0",
					"L2_access_VXLAN", "100G", 2.0,
				)
				links = append(links, link)
			}
		}
	}

	return devices, links
}

// Mixed Topology Generator
func (g *topologyGenerator) generateMixedTopology(fatTreeScale, spineLeafScale, threeTierScale float64, targetNetworkDevices int, includeServers bool) ([]topology.Device, []topology.Link) {
	var allDevices []topology.Device
	var allLinks []topology.Link

	// DC Core Interconnect Layer（ネットワークインフラデバイスの一部として計算）
	numCoreInterconnect := maxInt(1, int(float64(targetNetworkDevices)*0.005)) // 0.5%
	var coreInterconnectDevices []topology.Device
	for i := 0; i < numCoreInterconnect; i++ {
		device := g.createDevice(
			g.generateDeviceID("dccore"),
			"dc_core_interconnect",
			"Cisco NCS-5500",
			20, // DC Core Interconnect layer
		)
		allDevices = append(allDevices, device)
		coreInterconnectDevices = append(coreInterconnectDevices, device)
	}

	// Border Leaf Layer（ネットワークインフラデバイスの一部として計算）
	numBorderLeaf := maxInt(1, int(float64(targetNetworkDevices)*0.01)) // 1%
	var borderLeafDevices []topology.Device
	for i := 0; i < numBorderLeaf; i++ {
		device := g.createDevice(
			g.generateDeviceID("bl"),
			"border_leaf",
			"Arista 7280R3",
			10, // Border Router/Leaf layer
		)
		allDevices = append(allDevices, device)
		borderLeafDevices = append(borderLeafDevices, device)

		// Border Leaf to Core Interconnect
		if len(coreInterconnectDevices) > 0 {
			coreDevice := coreInterconnectDevices[rand.Intn(len(coreInterconnectDevices))]
			link := g.createLink(
				device.ID, coreDevice.ID,
				"Ethernet49", fmt.Sprintf("Ethernet%d", i+1),
				"L3_routed", "100G", 1.0,
			)
			allLinks = append(allLinks, link)
		}
	}

	// Fat-Tree Pod（残りのネットワークデバイス数を計算して配分）
	remainingDevices := targetNetworkDevices - len(allDevices)
	if fatTreeScale > 0 && remainingDevices > 0 {
		ftTargetDevices := int(float64(remainingDevices) * fatTreeScale)
		ftDevices, ftLinks := g.generateFatTreeTopology(
			maxInt(1, ftTargetDevices/200), // core spines: ~0.5%
			maxInt(1, ftTargetDevices/40),  // agg spines: ~2.5%
			maxInt(1, ftTargetDevices/4),   // edge leaves: ~25%
			includeServers,
		)
		allDevices = append(allDevices, ftDevices...)
		allLinks = append(allLinks, ftLinks...)

		// Connect Fat-Tree core spines to DC core
		for _, device := range ftDevices {
			if device.Type == "core_spine" && len(coreInterconnectDevices) > 0 {
				coreDevice := coreInterconnectDevices[rand.Intn(len(coreInterconnectDevices))]
				link := g.createLink(
					device.ID, coreDevice.ID,
					"Ethernet129", fmt.Sprintf("Ethernet%d", rand.Intn(32)+1),
					"L3_routed", "400G", 1.0,
				)
				allLinks = append(allLinks, link)
			}
		}
	}

	// Spine-Leaf Pod（残りのネットワークデバイス数を再計算して配分）
	remainingDevices = targetNetworkDevices - len(allDevices)
	if spineLeafScale > 0 && remainingDevices > 0 {
		slTargetDevices := int(float64(remainingDevices) * spineLeafScale / (spineLeafScale + threeTierScale))
		slDevices, slLinks := g.generateSpineLeafTopology(
			maxInt(1, slTargetDevices/50), // spines: ~2%
			maxInt(1, slTargetDevices/2),  // leaves: ~50%
			includeServers,
		)
		allDevices = append(allDevices, slDevices...)
		allLinks = append(allLinks, slLinks...)

		// Connect Spine-Leaf spines to DC core
		for _, device := range slDevices {
			if device.Type == "spine" && len(coreInterconnectDevices) > 0 {
				coreDevice := coreInterconnectDevices[rand.Intn(len(coreInterconnectDevices))]
				link := g.createLink(
					device.ID, coreDevice.ID,
					"swp32", fmt.Sprintf("Ethernet%d", rand.Intn(32)+33),
					"L3_routed", "100G", 1.0,
				)
				allLinks = append(allLinks, link)
			}
		}
	}

	// Three-Tier Pod（残りのネットワークデバイス数を再計算して配分）
	remainingDevices = targetNetworkDevices - len(allDevices)
	if threeTierScale > 0 && remainingDevices > 0 {
		ttTargetDevices := remainingDevices // 残り全てをThree-Tierに配分
		ttDevices, ttLinks := g.generateThreeTierTopology(
			maxInt(1, ttTargetDevices/100), // core: ~1%
			maxInt(1, ttTargetDevices/20),  // agg: ~5%
			maxInt(1, ttTargetDevices/5),   // access: ~20%
			includeServers,
		)
		allDevices = append(allDevices, ttDevices...)
		allLinks = append(allLinks, ttLinks...)

		// Connect Three-Tier cores to DC core
		for _, device := range ttDevices {
			if device.Type == "core" && len(coreInterconnectDevices) > 0 {
				coreDevice := coreInterconnectDevices[rand.Intn(len(coreInterconnectDevices))]
				link := g.createLink(
					device.ID, coreDevice.ID,
					"Ethernet49", fmt.Sprintf("Ethernet%d", rand.Intn(32)+65),
					"L3_routed_legacy", "100G", 1.0,
				)
				allLinks = append(allLinks, link)
			}
		}
	}

	return allDevices, allLinks
}

func runSeedDataEnhanced(cmd *cobra.Command, args []string) {
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	repo, err := repository.NewRepository(config.GetDatabaseConfig())
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer repo.Close()

	if verbose {
		log.Println("Connected to PostgreSQL")
	}

	ctx := context.Background()

	// Clear existing data if requested
	if clearFirst {
		log.Println("Clearing existing data...")
		if err := repo.Clear(); err != nil {
			log.Fatalf("Failed to clear data: %v", err)
		}
		log.Println("Existing data cleared")
	}

	// Initialize generator
	generator := newTopologyGenerator(dcLocation, locationDelimiter)

	var devices []topology.Device
	var links []topology.Link

	// Generate topology based on type
	log.Printf("Generating %s topology with %d target network infrastructure devices...", topologyType, targetDevices)

	switch topologyType {
	case "three-tier":
		// ネットワークインフラデバイスの構成比を調整
		numCore := maxInt(1, targetDevices/100)       // core: ~1%
		numAggPerCore := maxInt(1, targetDevices/20)  // agg: ~5%
		numAccessPerAgg := maxInt(1, targetDevices/5) // access: ~20%
		devices, links = generator.generateThreeTierTopology(
			numCore, numAggPerCore, numAccessPerAgg,
			includeServers,
		)
	case "spine-leaf":
		// ネットワークインフラデバイスの構成比を調整
		numSpines := maxInt(1, targetDevices/50)        // spines: ~2%
		numLeavesPerSpine := maxInt(1, targetDevices/2) // leaves: ~50%
		devices, links = generator.generateSpineLeafTopology(
			numSpines, numLeavesPerSpine,
			includeServers,
		)
	case "fat-tree":
		// ネットワークインフラデバイスの構成比を調整
		coreSpines := maxInt(1, targetDevices/200)      // core spines: ~0.5%
		aggSpinesPerCore := maxInt(1, targetDevices/40) // agg spines: ~2.5%
		edgeLeavesPerAgg := maxInt(1, targetDevices/4)  // edge leaves: ~25%
		devices, links = generator.generateFatTreeTopology(
			coreSpines, aggSpinesPerCore, edgeLeavesPerAgg,
			includeServers,
		)
	case "mixed":
		devices, links = generator.generateMixedTopology(
			fatTreeScale, spineLeafScale, threeTierScale,
			targetDevices, includeServers,
		)
	default:
		log.Fatalf("Unknown topology type: %s", topologyType)
	}

	// Add devices to database
	if err := repo.BulkAddDevices(ctx, devices); err != nil {
		log.Fatalf("Failed to add devices: %v", err)
	}
	log.Printf("Added %d devices", len(devices))

	// Add links to database
	if err := repo.BulkAddLinks(ctx, links); err != nil {
		log.Fatalf("Failed to add links: %v", err)
	}
	log.Printf("Added %d links", len(links))

	// Statistics
	networkDevices := filterDevicesByNonServerTypes(devices)
	serverDevices := filterDevicesByType(devices, "server")

	// 自動分類の実行
	if enableAutoClassifyEnhanced {
		log.Println("Applying auto-classification to enhanced seed devices...")

		// Repository includes both topology and classification interfaces
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

				// Group by device type for better readability
				typeCount := make(map[string]int)
				for _, c := range classifications {
					typeCount[c.DeviceType]++
					if len(classifications) <= 20 { // Only show details for small datasets
						log.Printf("  - %s → Layer %d (%s)", c.DeviceID, c.Layer, c.DeviceType)
					}
				}

				// Show summary for large datasets
				if len(classifications) > 20 {
					log.Printf("Classification summary by type:")
					for deviceType, count := range typeCount {
						log.Printf("  - %s: %d devices", deviceType, count)
					}
				}
			} else {
				log.Printf("No devices matched existing classification rules (this is normal for initial setup)")
				log.Printf("You can create classification rules in the web interface and then re-run with auto-classification")
			}
		}
	}

	log.Printf("\n=== Topology Generation Summary ===")
	log.Printf("Topology type: %s", topologyType)
	log.Printf("Network devices: %d", len(networkDevices))
	log.Printf("Server devices: %d", len(serverDevices))
	log.Printf("Total devices: %d", len(devices))
	log.Printf("Total links: %d", len(links))
	if dcLocation != "" {
		log.Printf("Datacenter location: %s", dcLocation)
	}

	log.Println("Enhanced sample data generation completed successfully")
}

func filterDevicesByType(devices []topology.Device, deviceType string) []topology.Device {
	var filtered []topology.Device
	for _, device := range devices {
		if device.Type == deviceType {
			filtered = append(filtered, device)
		}
	}
	return filtered
}

func filterDevicesByNonServerTypes(devices []topology.Device) []topology.Device {
	var filtered []topology.Device
	for _, device := range devices {
		if device.Type != "server" {
			filtered = append(filtered, device)
		}
	}
	return filtered
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
