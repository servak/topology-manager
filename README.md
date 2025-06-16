# Network Topology Manager

ネットワーク機器の階層分類・管理を中心としたシンプルなネットワーク管理システム。

## 主要機能

- **階層分類管理**: ネットワーク機器の自動・手動分類
- **分類ルール管理**: 命名規則や属性ベースの自動分類ルール
- **階層トポロジー表示**: レイヤー構造に基づく可視化
- **REST API**: OpenAPI準拠の管理API（Huma v2）
- **サイドバーUI**: 直感的な管理画面
- **データ同期**: Prometheusからのメトリクス自動収集

## アーキテクチャ

```
Prometheus → Worker → Database → API Server → Web UI
```

### 技術構成

- **バックエンド**: Go + Huma v2 (OpenAPI準拠)
- **データベース**: SQLite（開発・テスト）/ PostgreSQL（プロダクション）
- **フロントエンド**: React（サイドバー形式UI）
- **CLI**: Cobra
- **テスト**: Go標準テスト + SQLite（インメモリ）

## データベース戦略

開発効率とテスト容易性を重視したハイブリッド構成：

- **開発環境**: SQLite（ファイルベース）
- **テスト環境**: SQLite（インメモリ）
- **プロダクション**: PostgreSQL

同一のインターフェースで両方をサポートし、設定で切り替え可能。

## クイックスタート

### 1. 開発環境（SQLite使用）

```bash
# リポジトリクローン
git clone <repository-url>
cd topology-manager

# 依存関係インストール
go mod download

# SQLiteでデータベース初期化
go run ./cmd/ migrate up --db-type sqlite --db-path ./dev.db

# サンプルデータ投入
go run ./cmd/ seed --count 20

# API サーバー起動（SQLite使用）
go run ./cmd/ api --db-type sqlite --db-path ./dev.db
```

### 2. プロダクション環境（PostgreSQL使用）

```bash
# PostgreSQL起動
docker run -d --name postgres \
  -p 5432:5432 \
  -e POSTGRES_DB=topology_manager \
  -e POSTGRES_USER=tm \
  -e POSTGRES_PASSWORD=tm_password \
  postgres:15-alpine

# 環境変数設定
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_USER=tm
export DB_PASSWORD=tm_password
export DB_NAME=topology_manager

# マイグレーション実行
go run ./cmd/ migrate up

# API サーバー起動
go run ./cmd/ api --port 8080
```

### 3. フロントエンド起動

```bash
cd web
npm install
npm run dev
```

### 4. アクセス

- **Web UI**: http://localhost:8080
- **API ドキュメント**: http://localhost:8080/docs
- **ヘルスチェック**: http://localhost:8080/api/v1/health

## CLI コマンド

```bash
# API サーバー起動
topology-manager api [--port 8080] [--db-type sqlite|postgres]

# データ収集ワーカー起動  
topology-manager worker [--interval 300]

# データベースマイグレーション
topology-manager migrate up [--db-type sqlite|postgres]
topology-manager migrate down

# サンプルデータ生成
topology-manager seed --count 20
topology-manager seed --count 50 --clear

# バージョン表示
topology-manager version
```

## 主要APIエンドポイント

全てのAPIは `/api/v1` パスで始まり、OpenAPI準拠です。

### デバイス分類管理

```bash
# 未分類デバイス一覧
curl "http://localhost:8080/api/v1/classification/unclassified"

# 手動分類
curl -X POST "http://localhost:8080/api/v1/classification/devices/{deviceId}/classify" \
  -H "Content-Type: application/json" \
  -d '{"layer": 2, "device_type": "distribution", "user_id": "admin"}'

# 分類削除
curl -X DELETE "http://localhost:8080/api/v1/classification/devices/{deviceId}"
```

### 分類ルール管理

```bash
# ルール一覧
curl "http://localhost:8080/api/v1/classification/rules"

# ルール作成
curl -X POST "http://localhost:8080/api/v1/classification/rules" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Core Switch Rule",
    "conditions": [{"field": "name", "operator": "starts_with", "value": "core-"}],
    "layer": 1,
    "device_type": "core"
  }'
```

### 階層トポロジー

```bash
# 階層表示用トポロジー取得
curl "http://localhost:8080/api/v1/topology/visual/{deviceId}?depth=3"

# デバイス検索
curl "http://localhost:8080/api/v1/devices/search?q=switch"
```

## 設定

### 環境変数

```bash
# データベース設定
export DB_TYPE=sqlite              # sqlite または postgres
export DB_PATH=./dev.db           # SQLite使用時のファイルパス
export DB_HOST=localhost          # PostgreSQL接続先
export DB_USER=tm                 # PostgreSQLユーザー
export DB_PASSWORD=tm_password    # PostgreSQLパスワード
export DB_NAME=topology_manager   # データベース名

# API設定
export PORT=8080                  # APIサーバーポート

# Prometheus設定
export PROMETHEUS_URL=http://localhost:9090
```

### 設定ファイル（tm.yaml）

```yaml
database:
  type: sqlite  # または postgres
  sqlite:
    path: "./dev.db"
  postgres:
    host: ${DB_HOST:localhost}
    port: 5432
    user: ${DB_USER:tm}
    password: ${DB_PASSWORD:tm_password}
    dbname: ${DB_NAME:topology_manager}

prometheus:
  url: "${PROMETHEUS_URL:http://localhost:9090}"
```

## 開発・テスト

### 前提条件

- Go 1.21+
- Node.js 18+
- SQLite（開発）
- PostgreSQL（本番運用時）

### テスト実行

```bash
# 単体テスト（SQLiteインメモリ使用）
go test ./...

# 統合テスト
go test ./... -tags=integration

# カバレッジ取得
go test -cover ./...

# ベンチマークテスト
go test -bench=. ./...
```

### 開発環境

```bash
# バックエンド（ホットリロード）
go run ./cmd/ api --db-type sqlite --db-path ./dev.db

# フロントエンド（開発サーバー）
cd web && npm run dev

# データベース初期化
rm -f ./dev.db && go run ./cmd/ migrate up --db-type sqlite --db-path ./dev.db
```

### テスト用データベース

テストコードでは自動的にSQLiteインメモリデータベースを使用：

```go
// テスト用のリポジトリセットアップ
repo, err := postgres.NewRepository(":memory:", "sqlite")
if err != nil {
    t.Fatal(err)
}
defer repo.Close()
```

## ディレクトリ構成

```
topology-manager/
├── cmd/                    # CLIエントリーポイント
├── internal/              # 内部パッケージ
│   ├── api/handler/       # HTTPハンドラー
│   ├── domain/            # ドメインエンティティ
│   ├── repository/        # データアクセス層
│   │   ├── postgres/      # PostgreSQL/SQLite実装
│   │   └── sqlite/        # SQLite専用最適化（今後）
│   ├── service/           # ビジネスロジック
│   └── config/            # 設定管理
├── web/                   # React フロントエンド
├── tests/                 # 統合テスト
└── migrations/            # データベースマイグレーション
```

## ライセンス

MIT