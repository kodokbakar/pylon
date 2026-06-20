package chatservice

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	chatv1connect "github.com/kodokbakar/pylon/gen/pylon/chat/v1/chatv1connect"
	"github.com/kodokbakar/pylon/internal/config"
	"github.com/kodokbakar/pylon/internal/database"
	internalmetrics "github.com/kodokbakar/pylon/internal/metrics"
	"github.com/kodokbakar/pylon/internal/observability"
	internaltracing "github.com/kodokbakar/pylon/internal/tracing"
	chathandler "github.com/kodokbakar/pylon/services/chat-service/handler"
	chatkafka "github.com/kodokbakar/pylon/services/chat-service/kafka"
	chatrepository "github.com/kodokbakar/pylon/services/chat-service/repository"
	chatdomain "github.com/kodokbakar/pylon/services/chat-service/service"
)

type Server struct {
	cfg        *config.Config
	httpServer *http.Server
	db         *pgxpool.Pool
	producer   *chatkafka.Producer
}

func New(ctx context.Context, cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	db, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	producer, err := chatkafka.NewProducer(cfg.Kafka.Brokers, chatkafka.MessageEventsTopic, cfg.Kafka.ClientID)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	repo, err := chatrepository.NewMessageRepository(db)
	if err != nil {
		db.Close()
		_ = producer.Close()
		return nil, fmt.Errorf("create message repository: %w", err)
	}

	membershipRepo, err := chatrepository.NewMembershipRepository(db)
	if err != nil {
		db.Close()
		_ = producer.Close()
		return nil, fmt.Errorf("create membership repository: %w", err)
	}

	chatSvc, err := chatdomain.NewChatService(
		repo,
		producer,
		chatdomain.WithRoomMembershipChecker(membershipRepo),
	)
	if err != nil {
		db.Close()
		_ = producer.Close()
		return nil, fmt.Errorf("create chat service: %w", err)
	}

	handler, err := chathandler.NewChatHandler(chatSvc)
	if err != nil {
		db.Close()
		_ = producer.Close()
		return nil, fmt.Errorf("create chat handler: %w", err)
	}

	mux := http.NewServeMux()

	path, httpHandler := chatv1connect.NewChatServiceHandler(handler)
	mux.Handle(path, internalmetrics.GRPCMiddleware("chat-service", httpHandler))
	mux.Handle(path, internaltracing.HTTPMiddleware("chat-service", httpHandler))
	mux.Handle("GET /metrics", observability.MetricsHandler())
	mux.HandleFunc("GET /health", handleHealth)

	httpServer := &http.Server{
		Addr:    ":" + cfg.GRPC.Port,
		Handler: mux,
	}

	return &Server{
		cfg:        cfg,
		httpServer: httpServer,
		db:         db,
		producer:   producer,
	}, nil
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve chat service: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	var shutdownErr error

	if err := s.httpServer.Shutdown(ctx); err != nil {
		shutdownErr = fmt.Errorf("shutdown chat service: %w", err)
	}

	if s.producer != nil {
		if err := s.producer.Close(); err != nil && shutdownErr == nil {
			shutdownErr = fmt.Errorf("close kafka producer: %w", err)
		}
	}

	if s.db != nil {
		s.db.Close()
	}

	return shutdownErr
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok","service":"chat-service"}`)); err != nil {
		return
	}
}
