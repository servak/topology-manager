.PHONY: build dev web-build backend-build

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
