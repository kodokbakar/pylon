package apigateway

import (
	"context"
	"fmt"
	"net/http"

	chatv1connect "github.com/kodokbakar/pylon/gen/pylon/chat/v1/chatv1connect"
	roomv1connect "github.com/kodokbakar/pylon/gen/pylon/room/v1/roomv1connect"
	"github.com/kodokbakar/pylon/internal/config"
	internalmiddleware "github.com/kodokbakar/pylon/internal/middleware"
	"github.com/kodokbakar/pylon/internal/observability"
	"github.com/kodokbakar/pylon/internal/response"
	gatewayclient "github.com/kodokbakar/pylon/services/api-gateway/client"
	gatewayhandler "github.com/kodokbakar/pylon/services/api-gateway/handler"
	gatewaymiddleware "github.com/kodokbakar/pylon/services/api-gateway/middleware"
)

type Server struct {
	cfg              *config.Config
	clients          *gatewayclient.Clients
	mux              *http.ServeMux
	httpServer       *http.Server
	authHandler      *gatewayhandler.AuthHandler
	roomHandler      *gatewayhandler.RoomHandler
	messageHandler   *gatewayhandler.MessageHandler
	webSocketHandler *gatewayhandler.WebSocketHandler
	authMiddleware   *gatewaymiddleware.AuthMiddleware
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

	webSocketHandler, err := gatewayhandler.NewWebSocketHandler(
		cfg.WebSocket.MaxConnections,
		cfg.App.CORSOrigins,
		cfg.WebSocket.InsecureSkipVerify,
		chatClient,
	)
	if err != nil {
		return nil, fmt.Errorf("create websocket handler: %w", err)
	}

	server := &Server{
		cfg:              cfg,
		clients:          clients,
		mux:              http.NewServeMux(),
		authHandler:      gatewayhandler.NewAuthHandler(),
		roomHandler:      gatewayhandler.NewRoomHandler(roomClient),
		messageHandler:   gatewayhandler.NewMessageHandler(chatClient),
		webSocketHandler: webSocketHandler,
		authMiddleware:   authMiddleware,
	}

	server.registerRoutes()

	server.httpServer = &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: server.Handler(),
	}

	return server, nil
}

func (s *Server) Handler() http.Handler {
	return internalmiddleware.Recovery(
		internalmiddleware.Logger(
			internalmiddleware.CORSWithOrigins(s.mux, s.cfg.App.CORSOrigins),
		),
	)
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve api gateway: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.webSocketHandler.Shutdown()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown api gateway server: %w", err)
	}

	return nil
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.Handle("GET /metrics", observability.MetricsHandler())

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
