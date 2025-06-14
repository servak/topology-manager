import json
import random
import argparse

# グローバルなデバイスIDカウンター
device_id_counter = 0
# グローバルな拠点名サフィックス
global_dc_location_suffix = ""
# グローバルな区切り文字
global_dc_location_delimiter = "" # デフォルトは空文字列として、後で処理

def generate_device_id(prefix):
    global device_id_counter
    device_id_counter += 1
    # 区切り文字と拠点名サフィックスを追加
    # サフィックスが空の場合は区切り文字も追加しない
    if global_dc_location_suffix:
        return f"{prefix}-{device_id_counter:04d}{global_dc_location_delimiter}{global_dc_location_suffix}"
    else:
        return f"{prefix}-{device_id_counter:04d}"

def create_three_tier_topology(num_core, num_agg_per_core, num_access_per_agg, add_servers=True):
    """
    3層構造のトポロジーを生成する
    """
    devices = []
    connections = []
    
    # Core Layer
    core_switches = []
    for _ in range(num_core):
        core_id = generate_device_id("C")
        devices.append({"id": core_id, "type": "Core_Switch", "layer": "Core_3T"})
        core_switches.append(core_id)

    # Aggregation Layer
    agg_switches = []
    if core_switches: # Coreスイッチが存在する場合のみ集約層を生成
        for core_id in core_switches:
            for _ in range(num_agg_per_core):
                agg_id = generate_device_id("A")
                devices.append({"id": agg_id, "type": "Aggregation_Switch", "layer": "Aggregation_3T"})
                agg_switches.append(agg_id)
                connections.append({"from": agg_id, "to": core_id, "link_type": "L2_trunk_L3_routed"}) # Aggregation to Core

    # Access Layer
    access_switches = []
    if agg_switches: # 集約層スイッチが存在する場合のみアクセス層を生成
        for agg_id in agg_switches:
            for _ in range(num_access_per_agg):
                access_id = generate_device_id("AS")
                devices.append({"id": access_id, "type": "Access_Switch", "layer": "Access_3T"})
                access_switches.append(access_id)
                connections.append({"from": access_id, "to": agg_id, "link_type": "L2_trunk"}) # Access to Aggregation
                
                # エンドポイント（サーバーなど）への接続を表現
                if add_servers:
                    server_id = f"Server-{access_id}" # サーバーIDには通し番号の代わりにスイッチIDを使うことでユニーク性を確保
                    devices.append({"id": server_id, "type": "Server_Group", "layer": "Endpoint", "count": random.randint(10, 20)})
                    connections.append({"from": access_id, "to": server_id, "link_type": "L2_access"})

    return {"devices": devices, "connections": connections, "core_outputs": core_switches}

def create_spine_leaf_topology(num_spines, num_leaves_per_spine_base, add_servers=True):
    """
    Spine-Leafトポロジーを生成する
    """
    devices = []
    connections = []
    
    # Spine Layer
    spine_switches = []
    for _ in range(num_spines):
        spine_id = generate_device_id("SP")
        devices.append({"id": spine_id, "type": "Spine_Switch", "layer": "Spine_SL"})
        spine_switches.append(spine_id)

    # Leaf Layer
    leaf_switches = []
    if spine_switches: # Spineスイッチが存在する場合のみLeaf層を生成
        num_leaves_to_generate = int(num_leaves_per_spine_base * num_spines)
        if num_leaves_to_generate == 0 and num_spines > 0:
            num_leaves_to_generate = 1

        for _ in range(num_leaves_to_generate):
            leaf_id = generate_device_id("LF")
            devices.append({"id": leaf_id, "type": "Leaf_Switch", "layer": "Leaf_SL"})
            leaf_switches.append(leaf_id)
            
            # 各LeafはすべてのSpineに接続（ECMP）
            for spine_id in spine_switches:
                connections.append({"from": leaf_id, "to": spine_id, "link_type": "L3_routed_ECMP"})

            # エンドポイント（サーバーなど）への接続を表現
            if add_servers:
                server_id = f"Server-{leaf_id}"
                devices.append({"id": server_id, "type": "Server_Group", "layer": "Endpoint", "count": random.randint(20, 48)})
                connections.append({"from": leaf_id, "to": server_id, "link_type": "L2_access_VXLAN"})
            
    return {"devices": devices, "connections": connections, "spine_outputs": spine_switches}

def create_fat_tree_topology(core_spines, agg_spines_per_core, edge_leaves_per_agg, add_servers=True):
    """
    Fat-Tree（Closを多層化したもの）トポロジーを生成する
    """
    devices = []
    connections = []

    # Core Spine Layer (L3)
    core_spine_switches = []
    for _ in range(core_spines):
        cs_id = generate_device_id("FTCS")
        devices.append({"id": cs_id, "type": "FatTree_Core_Spine", "layer": "FT_Core"})
        core_spine_switches.append(cs_id)

    # Aggregation Spine Layer (L2/L3)
    agg_spine_switches = []
    if core_spine_switches: # Core Spineが存在する場合のみAgg Spine層を生成
        for cs_id in core_spine_switches:
            for _ in range(agg_spines_per_core):
                as_id = generate_device_id("FTAS")
                devices.append({"id": as_id, "type": "FatTree_Agg_Spine", "layer": "FT_Aggregation"})
                agg_spine_switches.append(as_id)
                connections.append({"from": as_id, "to": cs_id, "link_type": "L3_routed_ECMP"}) # Agg_Spine to Core_Spine

    # Edge/Leaf Layer (L2)
    edge_leaf_switches = []
    if agg_spine_switches: # Aggregation Spineが存在する場合のみEdge/Leaf層を生成
        for as_id in agg_spine_switches:
            for _ in range(edge_leaves_per_agg):
                el_id = generate_device_id("FTEL")
                devices.append({"id": el_id, "type": "FatTree_Edge_Leaf", "layer": "FT_Edge"})
                edge_leaf_switches.append(el_id)
                connections.append({"from": el_id, "to": as_id, "link_type": "L3_routed_ECMP_L2_VLAN_overlay"}) # Edge_Leaf to Agg_Spine

                # エンドポイント（サーバーなど）への接続を表現
                if add_servers:
                    server_id = f"Server-{el_id}"
                    devices.append({"id": server_id, "type": "Server_Group", "layer": "Endpoint", "count": random.randint(30, 60)})
                    connections.append({"from": el_id, "to": server_id, "link_type": "L2_access_VXLAN"})
                
    return {"devices": devices, "connections": connections, "core_spine_outputs": core_spine_switches}

def create_datacenter_topology(fat_tree_scale=1.0, spine_leaf_scale=1.0, three_tier_scale=1.0, total_target_devices=1000, dc_location=None, dc_location_delimiter="."):
    global device_id_counter # カウンターをリセット
    device_id_counter = 0
    
    global global_dc_location_suffix # グローバル変数に拠点サフィックスを設定
    global global_dc_location_delimiter # グローバル変数に区切り文字を設定

    if dc_location:
        global_dc_location_suffix = f"{dc_location.upper()}" # 大文字で統一
        global_dc_location_delimiter = dc_location_delimiter
    else:
        global_dc_location_suffix = ""
        global_dc_location_delimiter = "" # サフィックスがない場合はデリミタも使用しない

    all_devices = []
    all_connections = []
    
    # Core / Interconnect Layer (メインのSpineとなる部分)
    core_interconnect_switches = []
    num_core_interconnect = max(1, int(total_target_devices * 0.01)) 
    for _ in range(num_core_interconnect):
        core_id = generate_device_id("DCCore")
        all_devices.append({"id": core_id, "type": "DC_Core_Interconnect_Switch", "layer": "DC_Interconnect"})
        core_interconnect_switches.append(core_id)

    # Border Leaf Layer (外部接続、Core/Interconnectに接続)
    border_leaf_switches = []
    num_border_leaf = max(1, int(total_target_devices * 0.03)) 
    for _ in range(num_border_leaf):
        bl_id = generate_device_id("BL")
        all_devices.append({"id": bl_id, "type": "Border_Leaf_Switch", "layer": "Border"})
        border_leaf_switches.append(bl_id)
        
        if core_interconnect_switches:
            all_connections.append({"from": bl_id, "to": random.choice(core_interconnect_switches), "link_type": "L3_routed"})
        
        # 外部ネットワークのIDにも拠点名サフィックスを付与 (オプション)
        # 外部接続先はDCに依存しないため、デリミタとサフィックスを付与するかは要件次第ですが、ここでは統一
        ext_suffix = f"{global_dc_location_delimiter}{global_dc_location_suffix}" if global_dc_location_suffix else ""
        all_connections.append({"from": bl_id, "to": f"External_ISP_Router{ext_suffix}", "link_type": "BGP_peering"})
        all_connections.append({"from": bl_id, "to": f"WAN_Router_or_SDWAN_Hub{ext_suffix}", "link_type": "WAN_connection"})
        all_connections.append({"from": bl_id, "to": f"Security_Appliance{ext_suffix}", "link_type": "Service_Chain"})

    # ---- 各Podのパラメータをスケールに基づいて設定 ----
    # 基準となる台数設定 (合計約1000台時の目安)
    
    # Fat-Tree Pod (最新)
    print(f"Generating Fat-Tree Pod with scale: {fat_tree_scale}...")
    ft_pod = create_fat_tree_topology(
        core_spines=max(1, int(4 * fat_tree_scale)),
        agg_spines_per_core=max(1, int(5 * fat_tree_scale)),
        edge_leaves_per_agg=max(1, int(20 * fat_tree_scale))
    )
    all_devices.extend(ft_pod["devices"])
    all_connections.extend(ft_pod["connections"])
    if ft_pod["core_spine_outputs"] and core_interconnect_switches:
        for ft_core_spine_id in ft_pod["core_spine_outputs"]:
            all_connections.append({"from": ft_core_spine_id, "to": random.choice(core_interconnect_switches), "link_type": "L3_routed"})

    # Spine-Leaf Pod (既存)
    print(f"Generating Spine-Leaf Pod with scale: {spine_leaf_scale}...")
    sl_pod = create_spine_leaf_topology(
        num_spines=max(1, int(8 * spine_leaf_scale)),
        num_leaves_per_spine_base=max(1, int(40 * spine_leaf_scale))
    )
    all_devices.extend(sl_pod["devices"])
    all_connections.extend(sl_pod["connections"])
    if sl_pod["spine_outputs"] and core_interconnect_switches:
        for sl_spine_id in sl_pod["spine_outputs"]:
            all_connections.append({"from": sl_spine_id, "to": random.choice(core_interconnect_switches), "link_type": "L3_routed"})
        
    # Three-Tier Pod (レガシー)
    print(f"Generating Three-Tier Pod with scale: {three_tier_scale}...")
    tt_pod = create_three_tier_topology(
        num_core=max(1, int(2 * three_tier_scale)),
        num_agg_per_core=max(1, int(10 * three_tier_scale)),
        num_access_per_agg=max(1, int(5 * three_tier_scale))
    )
    all_devices.extend(tt_pod["devices"])
    all_connections.extend(tt_pod["connections"])
    if tt_pod["core_outputs"] and core_interconnect_switches:
        for tt_core_id in tt_pod["core_outputs"]:
            all_connections.append({"from": tt_core_id, "to": random.choice(core_interconnect_switches), "link_type": "L3_routed_legacy"})


    # 生成されたデバイス数を表示
    network_devices = [d for d in all_devices if d["type"] not in ["Server", "Server_Group"]]
    server_devices = [d for d in all_devices if d["type"] in ["Server", "Server_Group"]]
    total_server_count_val = sum(d.get('count', 1) for d in server_devices) # Server_Groupの場合はcountを使用

    print(f"\n--- Topology Generation Summary ---")
    print(f"Total network devices generated: {len(network_devices)}")
    print(f"Total server groups generated: {len(server_devices)}")
    print(f"Estimated total servers (individual): {total_server_count_val}")
    print(f"Grand total devices (network + server groups): {len(all_devices)}")
    print(f"Target network devices: {total_target_devices}")


    topology = {
        "datacenter_name": f"Mixed_Topology_DC{global_dc_location_delimiter}{global_dc_location_suffix}" if global_dc_location_suffix else "Mixed_Topology_DC_Default",
        "description": "Data Center with mixed topologies: Fat-Tree (latest), Spine-Leaf (existing), and Three-Tier (legacy). Includes randomly generated servers.",
        "parameters": {
            "total_target_network_devices": total_target_devices,
            "fat_tree_scale": fat_tree_scale,
            "spine_leaf_scale": spine_leaf_scale,
            "three_tier_scale": three_tier_scale,
            "datacenter_location_suffix": global_dc_location_suffix if dc_location else "N/A",
            "location_id_delimiter": global_dc_location_delimiter if dc_location else "N/A"
        },
        "summary": {
            "total_network_devices_generated": len(network_devices),
            "total_server_groups_generated": len(server_devices),
            "estimated_total_individual_servers": total_server_count_val
        },
        "layers_and_pods": {
            "DC_Interconnect_Layer": {
                "description": "Core layer connecting different pods and to Border Leaf switches.",
                "devices": [d for d in network_devices if d["layer"] == "DC_Interconnect"]
            },
            "Border_Layer": {
                "description": "Connects to external networks (ISP, WAN) and security services.",
                "devices": [d for d in network_devices if d["layer"] == "Border"]
            },
            "Fat_Tree_Pod": {
                "description": "Latest deployment, high East-West traffic optimization. Fat-Tree Core/Agg/Edge.",
                "devices": [d for d in network_devices if d["layer"].startswith("FT_")]
            },
            "Spine_Leaf_Pod": {
                "description": "Existing deployment, scalable for virtualized environments. Spine/Leaf.",
                "devices": [d for d in network_devices if d["layer"].endswith("_SL")]
            },
            "Three_Tier_Pod": {
                "description": "Legacy deployment, traditional Core-Aggregation-Access structure. Core/Agg/Access.",
                "devices": [d for d in network_devices if d["layer"].endswith("_3T")]
            },
            "Endpoint_Layer": {
                "description": "Servers and other end-user devices connected to access/leaf switches.",
                "devices": server_devices
            }
        },
        "connections": all_connections,
        "notes": [
            "Device IDs are unique per generation and include a location suffix if specified, separated by the chosen delimiter.",
            "Connection types indicate typical link usage (e.g., L3_routed, L2_trunk, VXLAN overlay).",
            "Specific port numbers, VLANs, IP addresses, and routing protocols are abstracted for simplicity.",
            "Random elements introduce variation in connected server group counts and specific inter-layer connections, and overall scaling of components.",
            "Scales of 0 for a topology will still generate a minimal set of devices (e.g., 1 per core/spine/aggregation layer) to ensure valid connections and avoid errors."
        ]
    }
    
    return topology

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate a mixed data center network topology (JSON).")
    parser.add_argument("--fat_tree_scale", type=float, default=1.0,
                        help="Scaling factor for the Fat-Tree pod components. Default: 1.0")
    parser.add_argument("--spine_leaf_scale", type=float, default=1.0,
                        help="Scaling factor for the Spine-Leaf pod components. Default: 1.0")
    parser.add_argument("--three_tier_scale", type=float, default=1.0,
                        help="Scaling factor for the Three-Tier pod components. Default: 1.0")
    parser.add_argument("--total_target_devices", type=int, default=1000,
                        help="Target total number of network devices (switches/routers). Server count is additional. Default: 1000")
    parser.add_argument("--output_file", type=str, default="mixed_dc_topology.json",
                        help="Output JSON file name. Default: mixed_dc_topology.json")
    parser.add_argument("--dc_location", type=str, default=None,
                        help="Optional location suffix for device IDs (e.g., 'A' for DC-A).")
    parser.add_argument("--dc_location_delimiter", type=str, default=".",
                        help="Delimiter character for the location suffix (e.g., '.' for 'DeviceID.Location'). Default: '.'")

    args = parser.parse_args()

    topology_data = create_datacenter_topology(
        fat_tree_scale=args.fat_tree_scale,
        spine_leaf_scale=args.spine_leaf_scale,
        three_tier_scale=args.three_tier_scale,
        total_target_devices=args.total_target_devices,
        dc_location=args.dc_location,
        dc_location_delimiter=args.dc_location_delimiter
    )
    
    with open(args.output_file, "w") as f:
        json.dump(topology_data, f, indent=2)
    print(f"\nJSON topology saved to {args.output_file}")