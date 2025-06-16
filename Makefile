.PHONY: build dev web-build backend-build test test-service test-api test-integration test-coverage

# デフォルトターゲット
all: build

##################################
## build
##################################
build: web-build backend-build

# フロントエンドのビルド
web-build:
	@make -C web build

# バックエンドのビルド
backend-build:
	@go build -o topology-manager ./cmd/

dev:
	@go run ./cmd/ api &
	@make -C web dev

##################################
## test
##################################
# 全テストの実行
test:
	@echo "Running all tests..."
	@go test -v ./internal/service/... ./internal/api/handler/...

# サービス層のテスト
test-service:
	@echo "Running service layer tests..."
	@go test -v ./internal/service/...

# API層のテスト
test-api:
	@echo "Running API handler tests..."
	@go test -v ./internal/api/handler/...

# 統合テスト (SQLiteを使用)
test-integration:
	@echo "Running integration tests..."
	@go test -v -tags=integration ./internal/...

# テストカバレッジ
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./internal/service/... ./internal/api/handler/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
