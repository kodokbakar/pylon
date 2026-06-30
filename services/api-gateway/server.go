package apigateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/kodokbakar/pylon/internal/database"
	gatewayauth "github.com/kodokbakar/pylon/services/api-gateway/auth"

	authv1connect "github.com/kodokbakar/pylon/gen/pylon/auth/v1/authv1connect"
	chatv1connect "github.com/kodokbakar/pylon/gen/pylon/chat/v1/chatv1connect"
	presencev1connect "github.com/kodokbakar/pylon/gen/pylon/presence/v1/presencev1connect"
	roomv1connect "github.com/kodokbakar/pylon/gen/pylon/room/v1/roomv1connect"
	"github.com/kodokbakar/pylon/internal/config"
	internalmetrics "github.com/kodokbakar/pylon/internal/metrics"
	internalmiddleware "github.com/kodokbakar/pylon/internal/middleware"
	"github.com/kodokbakar/pylon/internal/observability"
	"github.com/kodokbakar/pylon/internal/response"
	internaltracing "github.com/kodokbakar/pylon/internal/tracing"
	gatewayclient "github.com/kodokbakar/pylon/services/api-gateway/client"
	gatewayhandler "github.com/kodokbakar/pylon/services/api-gateway/handler"
	gatewaymiddleware "github.com/kodokbakar/pylon/services/api-gateway/middleware"
)

type Server struct {
	cfg                    *config.Config
	clients                *gatewayclient.Clients
	mux                    *http.ServeMux
	httpServer             *http.Server
	authDB                 *pgxpool.Pool
	redisClient            *redis.Client
	authHandler            *gatewayhandler.AuthHandler
	authConnectHandler     *gatewayhandler.AuthConnectHandler
	roomHandler            *gatewayhandler.RoomHandler
	roomConnectHandler     *gatewayhandler.RoomConnectHandler
	presenceConnectHandler *gatewayhandler.PresenceConnectHandler
	messageHandler         *gatewayhandler.MessageHandler
	webSocketHandler       *gatewayhandler.WebSocketHandler
	authMiddleware         *gatewaymiddleware.AuthMiddleware
	rateLimiter            *internalmiddleware.RateLimiter
}

func New(cfg *config.Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	clients, err := gatewayclient.NewClients(cfg.Services)
	if err != nil {
		return nil, fmt.Errorf("create gateway clients: %w", err)
	}

	authMiddleware, err := gatewaymiddleware.NewAuthMiddleware(cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("create auth middleware: %w", err)
	}

	chatClient := chatv1connect.NewChatServiceClient(
		clients.Chat.HTTPClient,
		clients.Chat.BaseURL,
	)

	roomClient := roomv1connect.NewRoomServiceClient(
		clients.Room.HTTPClient,
		clients.Room.BaseURL,
	)

	presenceClient := presencev1connect.NewPresenceServiceClient(
		clients.Presence.StreamingHTTPClient,
		clients.Presence.BaseURL,
	)

	presenceConnectHandler, err := gatewayhandler.NewPresenceConnectHandler(presenceClient)
	if err != nil {
		return nil, fmt.Errorf("create presence connect handler: %w", err)
	}

	roomConnectHandler, err := gatewayhandler.NewRoomConnectHandler(roomClient)
	if err != nil {
		return nil, fmt.Errorf("create room connect handler: %w", err)
	}

	webSocketHandler, err := gatewayhandler.NewWebSocketHandler(
		cfg.WebSocket.MaxConnections,
		cfg.App.CORSOrigins,
		cfg.WebSocket.InsecureSkipVerify,
		chatClient,
	)
	if err != nil {
		return nil, fmt.Errorf("create websocket handler: %w", err)
	}

	authService, authDB, err := newAuthService(cfg)
	if err != nil {
		return nil, fmt.Errorf("create auth service: %w", err)
	}

	rateLimiter, redisClient, err := newRateLimiter(cfg)
	if err != nil {
		if authDB != nil {
			authDB.Close()
		}
		return nil, fmt.Errorf("create rate limiter: %w", err)
	}

	server := &Server{
		cfg:                    cfg,
		clients:                clients,
		mux:                    http.NewServeMux(),
		authDB:                 authDB,
		redisClient:            redisClient,
		authHandler:            gatewayhandler.NewAuthHandler(authService),
		authConnectHandler:     gatewayhandler.NewAuthConnectHandler(authService),
		roomHandler:            gatewayhandler.NewRoomHandler(roomClient),
		roomConnectHandler:     roomConnectHandler,
		presenceConnectHandler: presenceConnectHandler,
		messageHandler:         gatewayhandler.NewMessageHandler(chatClient),
		webSocketHandler:       webSocketHandler,
		authMiddleware:         authMiddleware,
		rateLimiter:            rateLimiter,
	}

	server.registerRoutes()

	server.httpServer = &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: server.Handler(),
	}

	return server, nil
}

func newAuthService(cfg *config.Config) (gatewayhandler.AuthService, *pgxpool.Pool, error) {
	if strings.TrimSpace(cfg.Database.URL) == "" {
		return gatewayauth.NewUnavailableService("database url is not configured"), nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("create postgres pool: %w", err)
	}

	repo, err := gatewayauth.NewRepository(db)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("create auth repository: %w", err)
	}

	authService, err := gatewayauth.NewService(repo, cfg.JWT)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("create auth domain service: %w", err)
	}

	return authService, db, nil
}

func newRateLimiter(cfg *config.Config) (*internalmiddleware.RateLimiter, *redis.Client, error) {
	if strings.TrimSpace(cfg.Redis.URL) == "" {
		return nil, nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := database.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		return nil, nil, fmt.Errorf("create redis client: %w", err)
	}

	return internalmiddleware.NewRateLimiter(client), client, nil
}

func (s *Server) Handler() http.Handler {
	handler := internalmetrics.HTTPMiddleware(s.mux)
	handler = internaltracing.HTTPMiddleware("api-gateway", handler)
	handler = internalmiddleware.Logger(handler)
	handler = internalmiddleware.Recovery(handler)
	handler = internalmiddleware.CORSWithOrigins(handler, s.cfg.App.CORSOrigins)

	return handler
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve api gateway: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.webSocketHandler.Shutdown()

	var shutdownErr error
	if err := s.httpServer.Shutdown(ctx); err != nil {
		shutdownErr = fmt.Errorf("shutdown api gateway server: %w", err)
	}

	if s.authDB != nil {
		s.authDB.Close()
	}

	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil && shutdownErr == nil {
			shutdownErr = fmt.Errorf("close redis client: %w", err)
		}
	}

	return shutdownErr
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.Handle("GET /metrics", observability.MetricsHandler())

	authConnectPath, authConnectHandler := authv1connect.NewAuthServiceHandler(s.authConnectHandler)
	s.mux.Handle(authConnectPath, authConnectHandler)

	roomConnectPath, roomConnectHandler := roomv1connect.NewRoomServiceHandler(s.roomConnectHandler)
	s.mux.Handle(roomConnectPath, s.authMiddleware.RequireAuth(roomConnectHandler))

	presenceConnectPath, presenceConnectHandler := presencev1connect.NewPresenceServiceHandler(s.presenceConnectHandler)
	s.mux.Handle(presenceConnectPath, s.authMiddleware.RequireAuth(presenceConnectHandler))

	s.mux.HandleFunc("POST /api/v1/auth/register", s.authHandler.Register)
	s.mux.HandleFunc("POST /api/v1/auth/login", s.authHandler.Login)
	s.mux.HandleFunc("POST /api/v1/auth/refresh", s.authHandler.Refresh)

	s.mux.Handle("GET /ws", s.authMiddleware.RequireAuth(http.HandlerFunc(s.webSocketHandler.Handle)))

	s.mux.Handle("GET /api/v1/rooms", s.authMiddleware.RequireAuth(http.HandlerFunc(s.roomHandler.ListRooms)))
	s.mux.Handle("POST /api/v1/rooms", s.authMiddleware.RequireAuth(http.HandlerFunc(s.roomHandler.CreateRoom)))
	s.mux.Handle("POST /api/v1/rooms/{id}/join", s.authMiddleware.RequireAuth(http.HandlerFunc(s.roomHandler.JoinRoom)))
	s.mux.Handle("POST /api/v1/rooms/{id}/leave", s.authMiddleware.RequireAuth(http.HandlerFunc(s.roomHandler.LeaveRoom)))
	s.mux.Handle("GET /api/v1/rooms/{id}/messages", s.authMiddleware.RequireAuth(http.HandlerFunc(s.messageHandler.ListMessages)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "api-gateway",
	})
}
