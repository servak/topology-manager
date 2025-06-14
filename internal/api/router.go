package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/servak/topology-manager/internal/api/handler"
	apimiddleware "github.com/servak/topology-manager/internal/api/middleware"
	"github.com/servak/topology-manager/internal/domain/classification"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/service"
)

type Server struct {
	api                   huma.API
	router                chi.Router
	topologyService       *service.TopologyService
	visualizationService  *service.VisualizationService
	classificationService *service.ClassificationService
	topologyRepo          topology.Repository
	classificationRepo    classification.Repository
}

func NewServer(topologyRepo topology.Repository, classificationRepo classification.Repository) *Server {
	router := chi.NewRouter()

	// ミドルウェア
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(apimiddleware.Handler)

	// Huma API の設定
	config := huma.DefaultConfig("Network Topology Management API", "1.0.0")
	config.DocsPath = "/docs"
	config.Info.Description = "API for managing network topology and visualization"
	api := humachi.New(router, config)

	// サービス層の初期化
	topologyService := service.NewTopologyService(topologyRepo)
	visualizationService := service.NewVisualizationService(topologyRepo)
	classificationService := service.NewClassificationService(classificationRepo, topologyRepo)

	server := &Server{
		api:                   api,
		router:                router,
		topologyService:       topologyService,
		visualizationService:  visualizationService,
		classificationService: classificationService,
		topologyRepo:          topologyRepo,
		classificationRepo:    classificationRepo,
	}

	server.registerRoutes()

	return server
}

func (s *Server) registerRoutes() {
	// ハンドラーの初期化
	topologyHandler := handler.NewTopologyHandler(s.topologyService)
	visualizationHandler := handler.NewVisualizationHandler(s.visualizationService)
	classificationHandler := handler.NewClassificationHandler(s.classificationService)
	healthHandler := handler.NewHealthHandler(s.topologyRepo)

	// ルート登録
	topologyHandler.Register(s.api)
	visualizationHandler.Register(s.api)
	classificationHandler.RegisterRoutes(s.api)
	healthHandler.Register(s.api)

	// 静的ファイル配信（Web UI）
	s.router.Handle("/*", http.FileServer(http.Dir("./web/build/")))
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.topologyRepo.Close()
}