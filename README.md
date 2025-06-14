# Network Topology Manager

PrometheusからSNMP/LLDP情報を収集し、ネットワーク機器の階層トポロジーを可視化するシステム。

## 機能

- PrometheusからLLDPメトリクス自動収集
- デバイス階層分類
- インタラクティブなトポロジー可視化
- OpenAPI準拠のREST API（Huma v2）
- PostgreSQLによる永続化
- 単一バイナリでCLI実行

## アーキテクチャ

```
Prometheus ← Worker → PostgreSQL → API (Huma) → React UI
```

### 技術構成

- **バックエンド**: Go + Huma v2 (OpenAPI準拠)
- **データベース**: PostgreSQL
- **フロントエンド**: React + Cytoscape.js
- **CLI**: Cobra
- **データソース**: Prometheus (LLDP metrics)

## クイックスタート

### 1. 依存関係の起動

```bash
# PostgreSQL起動
docker run -d --name postgres -p 5432:5432 -e POSTGRES_DB=topology_manager -e POSTGRES_USER=topology -e POSTGRES_PASSWORD=topology postgres:15-alpine
```

### 2. データベースセットアップ

```bash
# マイグレーション実行
go run ./cmd/ migrate up

# または CLI使用
./topology-manager migrate up
```

### 3. アプリケーションのビルド

```bash
# Goアプリケーション
go build -o topology-manager ./cmd/

# フロントエンド（オプション）
cd web
npm install
npm run build
```

### 4. 設定

環境変数設定:
```bash
export DATABASE_URL="postgres://topology:topology@localhost/topology_manager?sslmode=disable"
export PROMETHEUS_URL=http://localhost:9090
```

### 5. 実行

```bash
# データ収集ワーカー起動
./topology-manager worker --interval 300

# API サーバー起動
./topology-manager api --port 8080
```

### 6. アクセス

- Web UI: http://localhost:8080
- API ドキュメント: http://localhost:8080/docs
- トポロジー取得: http://localhost:8080/api/topology/device.example?depth=3

## CLI コマンド

```bash
# API サーバー起動
topology-manager api --port 8080

# データ収集ワーカー起動  
topology-manager worker --interval 300

# データベースマイグレーション
topology-manager migrate up
topology-manager migrate down

# サンプルデータ生成
topology-manager seed --count 20
topology-manager seed --count 50 --clear

# バージョン表示
topology-manager version

# ヘルプ
topology-manager --help
```

## API エンドポイント

全てのAPIエンドポイントは `/api` パスで始まり、OpenAPI準拠の自動生成ドキュメントが `/docs` で確認できます。

### GET /api/topology/{deviceId}
トポロジー取得

**Parameters:**
- `deviceId` (path, required): 起点デバイスID
- `depth` (query, default: 3): 探索深度

**Example:**
```bash
curl "http://localhost:8080/api/topology/s4.colo?depth=3"
```

### GET /api/devices
デバイス一覧取得（ページング対応）

```bash
# 基本的な一覧取得
curl "http://localhost:8080/api/devices"

# ページング指定
curl "http://localhost:8080/api/devices?page=2&page_size=50"

# フィルタリング + ソート
curl "http://localhost:8080/api/devices?type=switch&order_by=layer&sort_dir=desc"

# 複合条件
curl "http://localhost:8080/api/devices?hardware=Arista&page=1&page_size=10&order_by=name"
```

### GET /api/devices/{deviceId}
デバイス詳細情報

```bash
curl "http://localhost:8080/api/devices/s4.colo"
```

### GET /api/devices/{deviceId}/neighbors
デバイスと隣接機器情報

```bash
curl "http://localhost:8080/api/devices/s4.colo/neighbors"
```

### POST /api/devices
デバイス追加

```bash
curl -X POST "http://localhost:8080/api/devices" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "device-001",
    "name": "example-switch",
    "type": "switch",
    "hardware": "Arista 7280",
    "layer": 4
  }'
```

### POST /api/links
リンク追加

```bash
curl -X POST "http://localhost:8080/api/links" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "link-001",
    "source_id": "device-001",
    "target_id": "device-002",
    "source_port": "Ethernet1",
    "target_port": "Ethernet1"
  }'
```

### GET /api/health
ヘルスチェック

```bash
curl "http://localhost:8080/api/health"
```

## 設定

### 階層分類 (config/hierarchy.yaml)

```yaml
hierarchy:
  device_types:
    core: 3
    distribution: 4
    access: 5
    server: 6

  naming_rules:
    - pattern: "^core.*"
      type: "core"
    - pattern: "^access.*"
      type: "access"

  manual_overrides:
    "special-device": "core"
```

### 環境変数

- `DATABASE_URL`: PostgreSQL接続先 (default: postgres://topology:topology@localhost/topology_manager?sslmode=disable)
- `PORT`: APIサーバーポート (default: 8080)
- `PROMETHEUS_URL`: Prometheus URL (default: http://localhost:9090)
- `TOPOLOGY_CONFIG_PATH`: 設定ファイルパス
- `WEB_DIR`: Webアセットディレクトリ

## 開発

### 前提条件

- Go 1.21+
- Node.js 18+
- PostgreSQL
- Prometheus (LLDP metrics)

### 開発環境起動

```bash
# バックエンド
go run ./cmd/main.go api

# フロントエンド
cd web
npm run dev
```

## ライセンス

MIT
