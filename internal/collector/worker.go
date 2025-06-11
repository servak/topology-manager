package collector

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/servak/topology-manager/internal/config"
	"github.com/servak/topology-manager/internal/storage"
)

type Worker struct {
	prometheus *PrometheusClient
	redis      *storage.RedisClient
	config     *config.Config
}

func NewWorker(prometheus *PrometheusClient, redis *storage.RedisClient, config *config.Config) *Worker {
	return &Worker{
		prometheus: prometheus,
		redis:      redis,
		config:     config,
	}
}

func (w *Worker) Run(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting worker with interval %v", interval)

	if err := w.collectOnce(ctx); err != nil {
		log.Printf("Initial collection failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker context cancelled, stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := w.collectOnce(ctx); err != nil {
				log.Printf("Collection failed: %v", err)
			}
		}
	}
}

func (w *Worker) collectOnce(ctx context.Context) error {
	log.Println("Starting LLDP metrics collection")

	metrics, err := w.prometheus.QueryLLDPMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to query LLDP metrics: %w", err)
	}

	log.Printf("Retrieved %d LLDP metrics", len(metrics))

	if err := w.redis.ClearTopology(ctx); err != nil {
		return fmt.Errorf("failed to clear existing topology: %w", err)
	}

	deviceMap := make(map[string]*storage.Device)
	links := make([]storage.Link, 0)

	for _, metric := range metrics {
		sourceDevice := metric.Instance
		targetDevice := metric.LLDPRemSysName

		if sourceDevice == "" || targetDevice == "" {
			continue
		}

		if _, exists := deviceMap[sourceDevice]; !exists {
			deviceType, layer, err := w.config.GetDeviceType(sourceDevice)
			if err != nil {
				log.Printf("Failed to get device type for %s: %v", sourceDevice, err)
				deviceType = "unknown"
				layer = w.config.Hierarchy.DeviceTypes["unknown"]
			}

			deviceMap[sourceDevice] = &storage.Device{
				Name:     sourceDevice,
				Type:     deviceType,
				Hardware: metric.Hardware,
				Status:   "active",
				Layer:    layer,
			}
		}

		if _, exists := deviceMap[targetDevice]; !exists {
			deviceType, layer, err := w.config.GetDeviceType(targetDevice)
			if err != nil {
				log.Printf("Failed to get device type for %s: %v", targetDevice, err)
				deviceType = "unknown"
				layer = w.config.Hierarchy.DeviceTypes["unknown"]
			}

			deviceMap[targetDevice] = &storage.Device{
				Name:     targetDevice,
				Type:     deviceType,
				Hardware: "",
				Status:   "active",
				Layer:    layer,
			}
		}

		link := storage.Link{
			Source:     sourceDevice,
			Target:     targetDevice,
			LocalPort:  metric.IfDescr,
			RemotePort: metric.LLDPRemPortId,
			Status:     "up",
		}
		links = append(links, link)

		if err := w.redis.AddNeighbor(ctx, sourceDevice, targetDevice); err != nil {
			log.Printf("Failed to add neighbor %s -> %s: %v", sourceDevice, targetDevice, err)
		}
	}

	for _, device := range deviceMap {
		if err := w.redis.SetDevice(ctx, *device); err != nil {
			log.Printf("Failed to store device %s: %v", device.Name, err)
		}

		if err := w.redis.AddDeviceToLayer(ctx, device.Layer, device.Name); err != nil {
			log.Printf("Failed to add device %s to layer %d: %v", device.Name, device.Layer, err)
		}
	}

	for _, link := range links {
		if err := w.redis.SetLink(ctx, link); err != nil {
			log.Printf("Failed to store link %s -> %s: %v", link.Source, link.Target, err)
		}
	}

	log.Printf("Stored %d devices and %d links", len(deviceMap), len(links))
	return nil
}