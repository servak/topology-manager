package api

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/servak/topology-manager/internal/api/handler"
	apimiddleware "github.com/servak/topology-manager/internal/api/middleware"
	"github.com/servak/topology-manager/internal/domain/classification"
	"github.com/servak/topology-manager/internal/domain/topology"
	"github.com/servak/topology-manager/internal/service"
	"github.com/servak/topology-manager/pkg/logger"
)

type Server struct {
	api                   huma.API
	router                chi.Router
	topologyService       *service.TopologyService
	visualizationService  *service.VisualizationService
	classificationService *service.ClassificationService
	topologyRepo          topology.Repository
	classificationRepo    classification.Repository
	logger                *logger.Logger
}

func NewServer(topologyRepo topology.Repository, classificationRepo classification.Repository, appLogger *logger.Logger) *Server {
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
		logger:                appLogger,
	}

	server.registerRoutes()

	return server
}

func (s *Server) registerRoutes() {
	// ハンドラーの初期化
	topologyHandler := handler.NewTopologyHandler(s.topologyService, s.logger)
	visualizationHandler := handler.NewVisualizationHandler(s.visualizationService, s.logger)
	classificationHandler := handler.NewClassificationHandler(s.classificationService, s.logger)
	healthHandler := handler.NewHealthHandler(s.topologyRepo, s.logger)

	// ルート登録
	topologyHandler.Register(s.api)
	visualizationHandler.Register(s.api)
	classificationHandler.RegisterRoutes(s.api)
	healthHandler.Register(s.api)

	// 静的ファイル配信（Web UI）- SPAルーティング対応
	s.setupSPARouting()
}

// setupSPARouting configures routing for Single Page Application
func (s *Server) setupSPARouting() {
	// 静的ファイルのディレクトリ
	staticDir := "./web/build"
	
	// アセットファイル（CSS, JS, images等）を直接配信
	s.router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(filepath.Join(staticDir, "assets")))))
	
	// Vite用の特別なファイル（存在する場合）
	s.router.HandleFunc("/vite.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "vite.svg"))
	})
	
	// APIルート以外のすべてのルートをSPAのindex.htmlにフォールバック
	s.router.NotFound(s.spaHandler(staticDir))
}

// spaHandler returns a handler that serves the SPA's index.html for non-API routes
func (s *Server) spaHandler(staticDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// API endpoints should not be handled by SPA
		if strings.HasPrefix(r.URL.Path, "/api/") || 
		   strings.HasPrefix(r.URL.Path, "/docs") ||
		   strings.HasPrefix(r.URL.Path, "/schemas") {
			http.NotFound(w, r)
			return
		}
		
		// 静的ファイルが存在するかチェック
		filePath := filepath.Join(staticDir, r.URL.Path)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			// ファイルが存在する場合は直接配信
			http.ServeFile(w, r, filePath)
			return
		}
		
		// SPA のルートの場合は index.html を配信
		indexPath := filepath.Join(staticDir, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			// index.html が存在しない場合
			s.logger.Error("index.html not found", "path", indexPath)
			http.NotFound(w, r)
			return
		}
		
		// Content-Type を設定
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, indexPath)
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.topologyRepo.Close()
}
