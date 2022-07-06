package api

import (
	"context"
	"net/http"
	"time"

	"github.com/ricoberger/gotemplate/pkg/api/middleware/httplog"
	"github.com/ricoberger/gotemplate/pkg/api/middleware/httptracer"
	"github.com/ricoberger/gotemplate/pkg/api/middleware/metrics"
	"github.com/ricoberger/gotemplate/pkg/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// Server is the interface for our API server.
type Server interface {
	Start()
	Stop()
}

type server struct {
	server *http.Server
}

// Start starts the API server.
func (s *server) Start() {
	log.Info(nil, "API server started", zap.String("address", s.server.Addr))

	if err := s.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Error(nil, "API server died unexpected", zap.Error(err))
		}
	}
}

// Stop terminates the API server gracefully.
func (s *server) Stop() {
	log.Debug(nil, "Start shutdown of the API server")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Error(nil, "Graceful shutdown of the API server failed", zap.Error(err))
	}
}

// New return a new API server.
func New(address string) (Server, error) {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"DELETE", "GET", "HEAD", "OPTIONS", "POST", "PUT"},
		AllowedHeaders: []string{"*"},
	}))
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(httptracer.Handler("gotemplate"))
	router.Use(metrics.Metrics)
	router.Use(httplog.Logger)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Get("/ping", pingHandler)
	router.Get("/request", requestHandler)

	return &server{
		server: &http.Server{
			Addr:    address,
			Handler: router,
		},
	}, nil
}
