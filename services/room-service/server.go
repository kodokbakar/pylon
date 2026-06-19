package roomservice

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	roomv1connect "github.com/kodokbakar/pylon/gen/pylon/room/v1/roomv1connect"
	"github.com/kodokbakar/pylon/internal/config"
	"github.com/kodokbakar/pylon/internal/database"
	internalmetrics "github.com/kodokbakar/pylon/internal/metrics"
	"github.com/kodokbakar/pylon/internal/observability"
	roomhandler "github.com/kodokbakar/pylon/services/room-service/handler"
	roomrepository "github.com/kodokbakar/pylon/services/room-service/repository"
	roomdomain "github.com/kodokbakar/pylon/services/room-service/service"
)

type Server struct {
	cfg        *config.Config
	httpServer *http.Server
	db         *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	db, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	roomRepo, err := roomrepository.NewRoomRepository(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create room repository: %w", err)
	}

	memberRepo, err := roomrepository.NewMemberRepository(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create member repository: %w", err)
	}

	roomSvc, err := roomdomain.NewRoomService(roomRepo, memberRepo)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create room service: %w", err)
	}

	handler, err := roomhandler.NewRoomHandler(roomSvc)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create room handler: %w", err)
	}

	mux := http.NewServeMux()

	path, httpHandler := roomv1connect.NewRoomServiceHandler(handler)
	mux.Handle(path, internalmetrics.GRPCMiddleware("room-service", httpHandler))
	mux.Handle("GET /metrics", observability.MetricsHandler())
	mux.HandleFunc("GET /health", handleHealth)

	server := &Server{
		cfg: cfg,
		db:  db,
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
		return fmt.Errorf("listen and serve room service: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	var shutdownErr error

	if err := s.httpServer.Shutdown(ctx); err != nil {
		shutdownErr = fmt.Errorf("shutdown room service: %w", err)
	}

	if s.db != nil {
		s.db.Close()
	}

	return shutdownErr
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok","service":"room-service"}`)); err != nil {
		return
	}
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(r.Context()); err != nil {
		http.Error(w, `{"status":"error","service":"room-service","postgres":"unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok","service":"room-service","postgres":"ok"}`)); err != nil {
		return
	}
}
