package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/servak/topology-manager/internal/api"
	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/storage"
	"github.com/servak/topology-manager/internal/topology"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configFile := configPath
	if configFile == "" {
		configFile = config.GetDefaultConfigPath()
	}

	_, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if verbose {
		log.Printf("Loaded config from: %s", configFile)
	}

	redis, err := storage.NewRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	if verbose {
		log.Println("Connected to Redis")
	}

	builder := topology.NewTopologyBuilder(redis)
	handler := api.NewTopologyHandler(builder, redis)
	server := api.NewServer(handler, apiPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	if err := server.Start(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("API server stopped")
}