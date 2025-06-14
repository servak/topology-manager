package handler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/repository/inmemory"
	"github.com/servak/topology-manager/internal/service"
)

func TestClassificationServiceBasic(t *testing.T) {
	// テスト用リポジトリの初期化
	classificationRepo := inmemory.NewClassificationRepository()
	topologyRepo := inmemory.NewTopologyRepository()

	// テスト用デバイスを追加
	testDevice := topology.Device{
		ID:       "fw1.edge",
		Type:     "firewall",
		Hardware: "Palo Alto",
		Instance: "192.168.1.1",
	}

	err := topologyRepo.AddDevice(context.Background(), testDevice)
	require.NoError(t, err)

	// サービス初期化
	classificationService := service.NewClassificationService(classificationRepo, topologyRepo)

	// デバイス分類テスト
	err = classificationService.ClassifyDevice(context.Background(), "fw1.edge", 1, "firewall", "admin")
	assert.NoError(t, err)

	// 分類結果の確認
	classification, err := classificationService.GetDeviceClassification(context.Background(), "fw1.edge")
	assert.NoError(t, err)
	assert.NotNil(t, classification)
	assert.Equal(t, "fw1.edge", classification.DeviceID)
	assert.Equal(t, 1, classification.Layer)
	assert.Equal(t, "firewall", classification.DeviceType)
	assert.True(t, classification.IsManual)
}

func TestUnclassifiedDevicesList(t *testing.T) {
	// テスト用リポジトリの初期化
	classificationRepo := inmemory.NewClassificationRepository()
	topologyRepo := inmemory.NewTopologyRepository()

	// テスト用デバイスを追加
	testDevices := []topology.Device{
		{ID: "fw1.edge", Type: "firewall", Hardware: "Palo Alto"},
		{ID: "sw1.core", Type: "switch", Hardware: "Arista 7280"},
		{ID: "srv1.web", Type: "server", Hardware: "Dell R740"},
	}

	for _, device := range testDevices {
		err := topologyRepo.AddDevice(context.Background(), device)
		require.NoError(t, err)
	}

	// サービス初期化
	classificationService := service.NewClassificationService(classificationRepo, topologyRepo)

	// 未分類デバイスの確認
	unclassified, err := classificationService.ListUnclassifiedDevices(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 3, len(unclassified))

	// 1つのデバイスを分類
	err = classificationService.ClassifyDevice(context.Background(), "fw1.edge", 1, "firewall", "admin")
	assert.NoError(t, err)

	// 未分類デバイスの数が減ったことを確認
	unclassified, err = classificationService.ListUnclassifiedDevices(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(unclassified))
}

func TestClassificationRules(t *testing.T) {
	// テスト用リポジトリの初期化
	classificationRepo := inmemory.NewClassificationRepository()
	topologyRepo := inmemory.NewTopologyRepository()

	// テスト用デバイスを追加
	testDevice := topology.Device{
		ID:       "fw1.edge",
		Type:     "firewall",
		Hardware: "Palo Alto",
	}

	err := topologyRepo.AddDevice(context.Background(), testDevice)
	require.NoError(t, err)

	// サービス初期化
	classificationService := service.NewClassificationService(classificationRepo, topologyRepo)

	// ルール作成とテスト
	rules, err := classificationService.ListClassificationRules(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(rules))

	t.Log("Classification service basic tests passed!")
}