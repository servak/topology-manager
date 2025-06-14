package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/servak/topology-manager/internal/api"
	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/internal/repository/postgres"
	"github.com/spf13/cobra"
)

var (
	apiPort string
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Long:  "Start the REST API server for topology queries and device information",
	Run:   runAPI,
}

func init() {
	apiCmd.Flags().StringVarP(&apiPort, "port", "p", "8080", "API server port")
}

func runAPI(cmd *cobra.Command, args []string) {
	// PostgreSQL DSN を環境変数から取得
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	repo, err := repository.NewDatabase(&config.Database)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer repo.Close()

	if verbose {
		log.Println("Connected to PostgreSQL")
	}

	// PostgreSQL specific implementation
	pgRepo, ok := repo.(*postgres.PostgresRepository)
	if !ok {
		log.Fatalf("Classification repository requires PostgreSQL, got %T", repo)
	}
	
	classificationRepo := postgres.NewClassificationRepository(pgRepo.GetDB())
	// APIサーバーの初期化
	server := api.NewServer(repo, classificationRepo)

	// HTTPサーバーの設定
	httpServer := &http.Server{
		Addr:    ":" + apiPort,
		Handler: server.Handler(),
	}

	// サーバーの開始
	go func() {
		log.Printf("Starting API server on port %s", apiPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// シグナルの待機
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Application shutdown error: %v", err)
	}

	log.Println("API server stopped")
}
