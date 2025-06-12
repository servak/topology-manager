package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/servak/topology-manager/internal/api"
	"github.com/servak/topology-manager/internal/repository/postgres"
)

func main() {
	// PostgreSQL DSN を環境変数から取得
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://topology:topology@localhost/topology_manager?sslmode=disable"
	}

	// PostgreSQLリポジトリの初期化
	repo, err := postgres.NewPostgresRepository(dsn)
	if err != nil {
		fmt.Printf("Failed to connect to PostgreSQL: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// APIサーバーの初期化
	server := api.NewServer(repo)

	// HTTPサーバーの設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: server.Handler(),
	}

	// サーバーの開始
	go func() {
		fmt.Printf("Starting server on port %s\n", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	// シグナルの待機
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down server...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Application shutdown error: %v\n", err)
	}

	fmt.Println("Server stopped")
}