package api

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router http.Handler
	port   string
}

func NewServer(handler *TopologyHandler, port string) *Server {
	if port == "" {
		port = "8080"
	}

	router := chi.NewRouter()
	
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	config := huma.DefaultConfig("Topology Manager API", "1.0.0")
	config.Info.Description = "Network topology management and visualization API"
	
	api := humachi.New(router, config)
	
	RegisterRoutes(api, handler)

	webDir := os.Getenv("WEB_DIR")
	if webDir == "" {
		webDir = "./web/build"
	}
	
	if _, err := os.Stat(webDir); err == nil {
		fileServer := http.FileServer(http.Dir(webDir))
		router.Handle("/*", fileServer)
	}

	return &Server{
		router: router,
		port:   port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%s", s.port)
	server := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	fmt.Printf("Starting server on port %s\n", s.port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}