package presenceservice

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"

	presencev1connect "github.com/kodokbakar/pylon/gen/pylon/presence/v1/presencev1connect"
	"github.com/kodokbakar/pylon/internal/config"
	"github.com/kodokbakar/pylon/internal/database"
	"github.com/kodokbakar/pylon/internal/observability"
	presencehandler "github.com/kodokbakar/pylon/services/presence-service/handler"
	presencerepository "github.com/kodokbakar/pylon/services/presence-service/repository"
	presencedomain "github.com/kodokbakar/pylon/services/presence-service/service"
)

type Server struct {
	cfg         *config.Config
	httpServer  *http.Server
	redisClient *redis.Client
}

func New(ctx context.Context, cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	redisClient, err := database.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}

	repo, err := presencerepository.NewPresenceRepository(redisClient)
	if err != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("create presence repository: %w", err)
	}

	presenceSvc, err := presencedomain.NewPresenceService(repo)
	if err != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("create presence service: %w", err)
	}

	handler, err := presencehandler.NewPresenceHandler(presenceSvc)
	if err != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("create presence handler: %w", err)
	}

	mux := http.NewServeMux()

	path, httpHandler := presencev1connect.NewPresenceServiceHandler(handler)
	mux.Handle(path, httpHandler)
	mux.Handle("GET /metrics", observability.MetricsHandler())
	mux.HandleFunc("GET /health", handleHealth)

	server := &Server{
		cfg:         cfg,
		redisClient: redisClient,
	}

	mux.HandleFunc("GET /ready", server.handleReady)

	server.httpServer = &http.Server{
		Addr:    ":" + cfg.GRPC.Port,
		Handler: mux,
	}

	return server, nil
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve presence service: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	var shutdownErr error

	if err := s.httpServer.Shutdown(ctx); err != nil {
		shutdownErr = fmt.Errorf("shutdown presence service: %w", err)
	}

	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil && shutdownErr == nil {
			shutdownErr = fmt.Errorf("close redis client: %w", err)
		}
	}

	return shutdownErr
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok","service":"presence-service"}`)); err != nil {
		return
	}
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.redisClient.Ping(r.Context()).Err(); err != nil {
		http.Error(w, `{"status":"error","service":"presence-service","redis":"unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok","service":"presence-service","redis":"ok"}`)); err != nil {
		return
	}
}
