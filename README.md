# Network Topology Manager

PrometheusからSNMP/LLDP情報を収集し、ネットワーク機器の階層トポロジーを可視化するシステム。

## 機能

- PrometheusからLLDPメトリクス自動収集
- デバイス階層分類（設定ファイルベース）
- インタラクティブなトポロジー可視化
- REST API提供
- 単一バイナリでCLI実行

## アーキテクチャ

```
Prometheus → Worker → Redis → API → React UI
```

## クイックスタート

### 1. 依存関係の起動

```bash
# Redis起動
docker run -d --name redis -p 6379:6379 redis:7-alpine

# または Docker Compose使用
cd deployments
docker-compose up -d redis
```

### 2. アプリケーションのビルド

```bash
# Goアプリケーション
go build -o topology-manager ./cmd/main.go

# フロントエンド（オプション）
cd web
npm install
npm run build
```

### 3. 設定

環境変数設定:
```bash
export REDIS_ADDR=localhost:6379
export PROMETHEUS_URL=http://localhost:9090
```

### 4. 実行

```bash
# データ収集ワーカー起動
./topology-manager worker --interval 300

# API サーバー起動
./topology-manager api --port 8080
```

### 5. アクセス

- Web UI: http://localhost:8080
- API: http://localhost:8080/topology?hostname=device.example

## CLI コマンド

```bash
# API サーバー起動
topology-manager api --port 8080

# データ収集ワーカー起動  
topology-manager worker --interval 300

# バージョン表示
topology-manager version

# ヘルプ
topology-manager --help
```

## API エンドポイント

### GET /topology
トポロジー取得

**Parameters:**
- `hostname` (required): 起点デバイス名
- `depth` (default: 3): 探索深度

**Example:**
```bash
curl "http://localhost:8080/topology?hostname=s4.colo&depth=3"
```

### GET /device/{name}
デバイス詳細情報

```bash
curl "http://localhost:8080/device/s4.colo"
```

### GET /health
ヘルスチェック

```bash
curl "http://localhost:8080/health"
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

- `REDIS_ADDR`: Redis接続先 (default: localhost:6379)
- `REDIS_PASSWORD`: Redisパスワード
- `REDIS_DB`: RedisDB番号 (default: 0)
- `PROMETHEUS_URL`: Prometheus URL (default: http://localhost:9090)
- `TOPOLOGY_CONFIG_PATH`: 設定ファイルパス
- `WEB_DIR`: Webアセットディレクトリ

## 開発

### 前提条件

- Go 1.21+
- Node.js 18+
- Redis
- Prometheus (LLDP metrics)

### 開発環境起動

```bash
# バックエンド
go run ./cmd/main.go api

# フロントエンド
cd web
npm run dev
```

### Docker Compose

```bash
cd deployments
docker-compose up --build
```

## トラブルシューティング

### Prometheus接続エラー
- PROMETHEUS_URL環境変数を確認
- Prometheusの起動状態を確認
- LLDPメトリクス(`lldpRemSysName`)の存在を確認

### Redis接続エラー
- REDIS_ADDR環境変数を確認
- Redisの起動状態を確認

### 空のトポロジー
- Prometheusにlldpメトリクスが存在するか確認
- デバイス名が階層設定とマッチするか確認
- ワーカーがデータ収集できているか確認

## ライセンス

MIT