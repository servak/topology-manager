package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/servak/topology-manager/internal/api"
	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/repository"
	"github.com/servak/topology-manager/pkg/logger"
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
	// ログシステムの初期化
	logLevel := "info"
	if verbose {
		logLevel = "debug"
	}
	appLogger := logger.New(logLevel)

	// PostgreSQL DSN を環境変数から取得
	config, err := config.LoadConfig(configPath)
	if err != nil {
		appLogger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	
	repo, err := repository.NewRepository(config.GetDatabaseConfig())
	if err != nil {
		appLogger.Error("Failed to create database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	appLogger.Info("Connected to PostgreSQL")

	// Repository includes both topology and classification interfaces
	// APIサーバーの初期化
	server := api.NewServer(repo, repo, appLogger)

	// HTTPサーバーの設定
	httpServer := &http.Server{
		Addr:    ":" + apiPort,
		Handler: server.Handler(),
	}

	// サーバーの開始
	go func() {
		appLogger.Info("Starting API server", "port", apiPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// シグナルの待機
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	appLogger.Info("Shutting down server...")

	// グレースフルシャットダウン
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		appLogger.Error("Server shutdown error", "error", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Application shutdown error", "error", err)
	}

	appLogger.Info("API server stopped")
}
