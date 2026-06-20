package notificationservice

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	notificationv1connect "github.com/kodokbakar/pylon/gen/pylon/notification/v1/notificationv1connect"
	roomv1connect "github.com/kodokbakar/pylon/gen/pylon/room/v1/roomv1connect"
	"github.com/kodokbakar/pylon/internal/config"
	"github.com/kodokbakar/pylon/internal/database"
	internalmetrics "github.com/kodokbakar/pylon/internal/metrics"
	"github.com/kodokbakar/pylon/internal/observability"
	internaltracing "github.com/kodokbakar/pylon/internal/tracing"
	notificationhandler "github.com/kodokbakar/pylon/services/notification-service/handler"
	notificationkafka "github.com/kodokbakar/pylon/services/notification-service/kafka"
	notificationrepository "github.com/kodokbakar/pylon/services/notification-service/repository"
	notificationdomain "github.com/kodokbakar/pylon/services/notification-service/service"
)

type Server struct {
	cfg            *config.Config
	httpServer     *http.Server
	db             *pgxpool.Pool
	consumer       *notificationkafka.Consumer
	consumerCancel context.CancelFunc
}

func New(ctx context.Context, cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	db, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	repo, err := notificationrepository.NewNotificationRepository(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create notification repository: %w", err)
	}

	notificationSvc, err := notificationdomain.NewNotificationService(repo, nil)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create notification service: %w", err)
	}

	handler, err := notificationhandler.NewNotificationHandler(notificationSvc)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create notification handler: %w", err)
	}

	roomClient := roomv1connect.NewRoomServiceClient(&http.Client{
		Timeout:   10 * time.Second,
		Transport: internaltracing.HTTPTransport(http.DefaultTransport),
	}, cfg.Services.RoomURL)

	consumer, err := notificationkafka.NewConsumer(
		cfg.Kafka.Brokers,
		notificationkafka.MessageEventsTopic,
		notificationkafka.NotificationConsumerGroupID,
		roomClient,
		notificationSvc,
	)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create notification kafka consumer: %w", err)
	}

	mux := http.NewServeMux()

	path, httpHandler := notificationv1connect.NewNotificationServiceHandler(handler)
	wrappedHandler := internalmetrics.GRPCMiddleware("notification-service", httpHandler)
	wrappedHandler = internaltracing.HTTPMiddleware("notification-service", wrappedHandler)
	mux.Handle(path, wrappedHandler)
	mux.Handle("GET /metrics", observability.MetricsHandler())
	mux.HandleFunc("GET /health", handleHealth)

	server := &Server{
		cfg:      cfg,
		db:       db,
		consumer: consumer,
	}

	mux.HandleFunc("GET /ready", server.handleReady)

	server.httpServer = &http.Server{
		Addr:    ":" + cfg.GRPC.Port,
		Handler: mux,
	}

	return server, nil
}

func (s *Server) Start() error {
	consumerCtx, cancel := context.WithCancel(context.Background())
	s.consumerCancel = cancel

	go func() {
		if err := s.consumer.Start(consumerCtx); err != nil {
			log.Printf("notification kafka consumer stopped: %v", err)
		}
	}()

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve notification service: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	var shutdownErr error

	if s.consumerCancel != nil {
		s.consumerCancel()
	}

	if s.consumer != nil {
		if err := s.consumer.Close(); err != nil {
			shutdownErr = fmt.Errorf("close kafka consumer: %w", err)
		}
	}

	if err := s.httpServer.Shutdown(ctx); err != nil && shutdownErr == nil {
		shutdownErr = fmt.Errorf("shutdown notification service: %w", err)
	}

	if s.db != nil {
		s.db.Close()
	}

	return shutdownErr
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok","service":"notification-service"}`)); err != nil {
		return
	}
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(r.Context()); err != nil {
		http.Error(w, `{"status":"error","service":"notification-service","postgres":"unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status":"ok","service":"notification-service","postgres":"ok"}`)); err != nil {
		return
	}
}
