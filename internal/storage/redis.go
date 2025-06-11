package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

type Device struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Hardware string `json:"hardware"`
	Status   string `json:"status"`
	Layer    int    `json:"layer"`
}

type Link struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	LocalPort  string `json:"local_port"`
	RemotePort string `json:"remote_port"`
	Status     string `json:"status"`
}

func NewRedisClient() (*RedisClient, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		var err error
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB value: %w", err)
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{client: rdb}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisClient) AddNeighbor(ctx context.Context, device, neighbor string) error {
	key := fmt.Sprintf("adj:%s", device)
	return r.client.SAdd(ctx, key, neighbor).Err()
}

func (r *RedisClient) GetNeighbors(ctx context.Context, device string) ([]string, error) {
	key := fmt.Sprintf("adj:%s", device)
	return r.client.SMembers(ctx, key).Result()
}

func (r *RedisClient) SetDevice(ctx context.Context, device Device) error {
	key := fmt.Sprintf("device:%s", device.Name)
	data, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %w", err)
	}
	return r.client.Set(ctx, key, data, 0).Err()
}

func (r *RedisClient) GetDevice(ctx context.Context, name string) (*Device, error) {
	key := fmt.Sprintf("device:%s", name)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var device Device
	if err := json.Unmarshal([]byte(data), &device); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device: %w", err)
	}
	return &device, nil
}

func (r *RedisClient) SetLink(ctx context.Context, link Link) error {
	key := fmt.Sprintf("link:%s:%s", link.Source, link.Target)
	data, err := json.Marshal(link)
	if err != nil {
		return fmt.Errorf("failed to marshal link: %w", err)
	}
	return r.client.Set(ctx, key, data, 0).Err()
}

func (r *RedisClient) GetLink(ctx context.Context, source, target string) (*Link, error) {
	key := fmt.Sprintf("link:%s:%s", source, target)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var link Link
	if err := json.Unmarshal([]byte(data), &link); err != nil {
		return nil, fmt.Errorf("failed to unmarshal link: %w", err)
	}
	return &link, nil
}

func (r *RedisClient) AddDeviceToLayer(ctx context.Context, layer int, device string) error {
	key := fmt.Sprintf("layer:%d", layer)
	return r.client.SAdd(ctx, key, device).Err()
}

func (r *RedisClient) GetDevicesInLayer(ctx context.Context, layer int) ([]string, error) {
	key := fmt.Sprintf("layer:%d", layer)
	return r.client.SMembers(ctx, key).Result()
}

func (r *RedisClient) GetAllDevices(ctx context.Context) ([]string, error) {
	keys, err := r.client.Keys(ctx, "device:*").Result()
	if err != nil {
		return nil, err
	}

	devices := make([]string, 0, len(keys))
	for _, key := range keys {
		if len(key) > 7 {
			devices = append(devices, key[7:])
		}
	}
	return devices, nil
}

func (r *RedisClient) ClearTopology(ctx context.Context) error {
	keys, err := r.client.Keys(ctx, "adj:*").Result()
	if err != nil {
		return err
	}
	
	deviceKeys, err := r.client.Keys(ctx, "device:*").Result()
	if err != nil {
		return err
	}
	keys = append(keys, deviceKeys...)
	
	linkKeys, err := r.client.Keys(ctx, "link:*").Result()
	if err != nil {
		return err
	}
	keys = append(keys, linkKeys...)
	
	layerKeys, err := r.client.Keys(ctx, "layer:*").Result()
	if err != nil {
		return err
	}
	keys = append(keys, layerKeys...)

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}