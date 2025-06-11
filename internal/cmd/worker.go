package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/servak/topology-manager/internal/collector"
	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/storage"
)

var (
	workerInterval int
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the data collection worker",
	Long:  "Start the worker process that collects LLDP metrics from Prometheus",
	Run:   runWorker,
}

func init() {
	workerCmd.Flags().IntVarP(&workerInterval, "interval", "i", 300, "Collection interval in seconds")
}

func runWorker(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configFile := configPath
	if configFile == "" {
		configFile = config.GetDefaultConfigPath()
	}

	cfg, err := config.LoadConfig(configFile)
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

	prometheus := collector.NewPrometheusClient()
	
	if err := prometheus.Health(ctx); err != nil {
		log.Fatalf("Prometheus health check failed: %v", err)
	}

	if verbose {
		log.Println("Prometheus connection verified")
	}

	worker := collector.NewWorker(prometheus, redis, cfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	interval := time.Duration(workerInterval) * time.Second
	if verbose {
		log.Printf("Starting worker with %v interval", interval)
	}

	if err := worker.Run(ctx, interval); err != nil {
		if err != context.Canceled {
			log.Fatalf("Worker failed: %v", err)
		}
	}

	log.Println("Worker stopped")
}