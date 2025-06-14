-- 010_insert_default_hierarchy_layers.sql
-- デフォルトの階層レイヤー定義を追加

-- ボーダー・レイヤー (Border Layer) - 10番台
INSERT INTO hierarchy_layers (id, name, description, order_index, color, created_at, updated_at) VALUES
(10, 'Border Router/Leaf', 'データセンターと外部ネットワークの境界。外部ASとのBGPピアリング、WAN接続の終端を行う物理的なルーターまたはBorder Leafスイッチ', 1, '#e74c3c', NOW(), NOW()),
(11, 'Security Appliances', 'ファイアウォール、IDS/IPS、DDoS防御など、外部トラフィックの検査・制御を行うセキュリティ機器', 2, '#c0392b', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    order_index = EXCLUDED.order_index,
    color = EXCLUDED.color,
    updated_at = EXCLUDED.updated_at;

-- データセンター・インターコネクト・レイヤー (DC Interconnect Layer) - 20番台  
INSERT INTO hierarchy_layers (id, name, description, order_index, color, created_at, updated_at) VALUES
(20, 'DC Core Interconnect Switches', 'データセンター内の異なるトポロジーセグメント間を接続。各ポッドの最上位スイッチからの接続を集約し、Border Layerへのルート提供を行う', 3, '#8e44ad', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    order_index = EXCLUDED.order_index,
    color = EXCLUDED.color,
    updated_at = EXCLUDED.updated_at;

-- ファブリック・レイヤー (Fabric Layer: スパイン) - 30番台
INSERT INTO hierarchy_layers (id, name, description, order_index, color, created_at, updated_at) VALUES
(30, 'Fat-Tree Core Spine', 'Fat-Tree構造における最上位のスパイン。大規模なEast-Westトラフィックの基盤となる高性能L3スイッチ', 4, '#2980b9', NOW(), NOW()),
(31, 'Fat-Tree Aggregation Spine', 'Fat-Tree構造における中間層のスパイン。コアスパインとエッジ/リーフスイッチ間の接続を提供', 5, '#3498db', NOW(), NOW()),
(32, 'Spine Switches (Spine-Leaf)', 'Spine-Leaf構造におけるスパイン。すべてのリーフスイッチと接続され、L3ルーティングを提供', 6, '#5dade2', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    order_index = EXCLUDED.order_index,
    color = EXCLUDED.color,
    updated_at = EXCLUDED.updated_at;

-- アクセス・レイヤー (Access Layer: リーフ / 従来の集約・アクセス) - 40番台
INSERT INTO hierarchy_layers (id, name, description, order_index, color, created_at, updated_at) VALUES
(40, 'Fat-Tree Edge/Leaf', 'Fat-Tree構造の最下層のスイッチ。サーバーを直接収容し、L2ネットワークの提供とL3へのルーティングエッジを担う', 7, '#27ae60', NOW(), NOW()),
(41, 'Leaf Switches (Spine-Leaf)', 'Spine-Leaf構造の最下層のスイッチ。サーバーを直接収容し、すべてのSpineスイッチに接続', 8, '#2ecc71', NOW(), NOW()),
(42, 'Aggregation Switches (3-tier)', '従来の3層構造における集約層。複数のアクセススイッチを収容し、コアスイッチへ接続', 9, '#58d68d', NOW(), NOW()),
(43, 'Access Switches (3-tier)', '従来の3層構造におけるアクセス層。エンドポイントを直接収容するエッジスイッチ', 10, '#85c1e9', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    order_index = EXCLUDED.order_index,
    color = EXCLUDED.color,
    updated_at = EXCLUDED.updated_at;

-- エンドポイント・レイヤー (Endpoint Layer) - 50番台
INSERT INTO hierarchy_layers (id, name, description, order_index, color, created_at, updated_at) VALUES
(50, 'Servers', '物理サーバー、仮想マシン、コンテナホストなど、実際のワークロードを実行するデバイス', 11, '#f39c12', NOW(), NOW()),
(51, 'Storage Devices', 'SAN、NASなど、データストレージを提供する専用機器', 12, '#e67e22', NOW(), NOW()),
(52, 'Other Appliances', 'ロードバランサー、サービスメッシュ、APIゲートウェイなど、特定のネットワークサービスを提供する機器', 13, '#d35400', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    order_index = EXCLUDED.order_index,
    color = EXCLUDED.color,
    updated_at = EXCLUDED.updated_at;