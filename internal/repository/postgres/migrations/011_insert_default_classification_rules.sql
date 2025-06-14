-- 011_insert_default_classification_rules.sql
-- デフォルトの分類ルールを追加

-- Border Router/Leaf関連のルール
INSERT INTO classification_rules (id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at) VALUES
(gen_random_uuid(), 'Border Router - Name Pattern', 'ボーダールーターの命名パターンによる分類', 'OR', 
 '[{"field":"name","operator":"contains","value":"border"},{"field":"name","operator":"contains","value":"edge"},{"field":"name","operator":"contains","value":"wan"}]'::jsonb,
 10, 'Border Router/Leaf', 90, true, 'system', NOW(), NOW()),

(gen_random_uuid(), 'Border Router - Hardware Type', 'ボーダールーターのハードウェアタイプによる分類', 'OR',
 '[{"field":"hardware","operator":"contains","value":"ASR"},{"field":"hardware","operator":"contains","value":"ISR"},{"field":"hardware","operator":"contains","value":"Border"}]'::jsonb,
 10, 'Border Router/Leaf', 85, true, 'system', NOW(), NOW());

-- Security Appliances関連のルール
INSERT INTO classification_rules (id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at) VALUES
(gen_random_uuid(), 'Security - Firewall Name Pattern', 'ファイアウォールの命名パターンによる分類', 'OR',
 '[{"field":"name","operator":"contains","value":"fw"},{"field":"name","operator":"contains","value":"firewall"},{"field":"name","operator":"contains","value":"security"}]'::jsonb,
 11, 'Security Appliances', 95, true, 'system', NOW(), NOW()),

(gen_random_uuid(), 'Security - Firewall Hardware', 'ファイアウォールのハードウェアタイプによる分類', 'OR',
 '[{"field":"hardware","operator":"contains","value":"ASA"},{"field":"hardware","operator":"contains","value":"Firewall"},{"field":"hardware","operator":"contains","value":"FortiGate"}]'::jsonb,
 11, 'Security Appliances', 90, true, 'system', NOW(), NOW());

-- Spine Switches関連のルール
INSERT INTO classification_rules (id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at) VALUES
(gen_random_uuid(), 'Spine Switch - Name Pattern', 'スパインスイッチの命名パターンによる分類', 'OR',
 '[{"field":"name","operator":"contains","value":"spine"},{"field":"name","operator":"contains","value":"core"},{"field":"name","operator":"contains","value":"spn"}]'::jsonb,
 32, 'Spine Switches (Spine-Leaf)', 80, true, 'system', NOW(), NOW()),

(gen_random_uuid(), 'Spine Switch - Hardware Type', 'スパインスイッチのハードウェアタイプによる分類', 'OR',
 '[{"field":"hardware","operator":"contains","value":"Nexus 9"},{"field":"hardware","operator":"contains","value":"Arista 72"},{"field":"hardware","operator":"contains","value":"QFX10"}]'::jsonb,
 32, 'Spine Switches (Spine-Leaf)', 75, true, 'system', NOW(), NOW());

-- Leaf Switches関連のルール
INSERT INTO classification_rules (id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at) VALUES
(gen_random_uuid(), 'Leaf Switch - Name Pattern', 'リーフスイッチの命名パターンによる分類', 'OR',
 '[{"field":"name","operator":"contains","value":"leaf"},{"field":"name","operator":"contains","value":"tor"},{"field":"name","operator":"contains","value":"access"}]'::jsonb,
 41, 'Leaf Switches (Spine-Leaf)', 70, true, 'system', NOW(), NOW()),

(gen_random_uuid(), 'Leaf Switch - Hardware Type', 'リーフスイッチのハードウェアタイプによる分類', 'OR',
 '[{"field":"hardware","operator":"contains","value":"Nexus 93"},{"field":"hardware","operator":"contains","value":"Arista 73"},{"field":"hardware","operator":"contains","value":"EX43"}]'::jsonb,
 41, 'Leaf Switches (Spine-Leaf)', 65, true, 'system', NOW(), NOW());

-- Server関連のルール
INSERT INTO classification_rules (id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at) VALUES
(gen_random_uuid(), 'Server - Name Pattern', 'サーバーの命名パターンによる分類', 'OR',
 '[{"field":"name","operator":"contains","value":"srv"},{"field":"name","operator":"contains","value":"server"},{"field":"name","operator":"contains","value":"host"}]'::jsonb,
 50, 'Servers', 60, true, 'system', NOW(), NOW()),

(gen_random_uuid(), 'Server - Hardware Type', 'サーバーのハードウェアタイプによる分類', 'OR',
 '[{"field":"hardware","operator":"contains","value":"Dell"},{"field":"hardware","operator":"contains","value":"HP"},{"field":"hardware","operator":"contains","value":"Cisco UCS"}]'::jsonb,
 50, 'Servers', 55, true, 'system', NOW(), NOW());

-- Storage関連のルール
INSERT INTO classification_rules (id, name, description, logic_operator, conditions, layer, device_type, priority, is_active, created_by, created_at, updated_at) VALUES
(gen_random_uuid(), 'Storage - Name Pattern', 'ストレージの命名パターンによる分類', 'OR',
 '[{"field":"name","operator":"contains","value":"san"},{"field":"name","operator":"contains","value":"nas"},{"field":"name","operator":"contains","value":"storage"}]'::jsonb,
 51, 'Storage Devices', 65, true, 'system', NOW(), NOW()),

(gen_random_uuid(), 'Storage - Hardware Type', 'ストレージのハードウェアタイプによる分類', 'OR',
 '[{"field":"hardware","operator":"contains","value":"NetApp"},{"field":"hardware","operator":"contains","value":"EMC"},{"field":"hardware","operator":"contains","value":"Pure Storage"}]'::jsonb,
 51, 'Storage Devices', 60, true, 'system', NOW(), NOW());