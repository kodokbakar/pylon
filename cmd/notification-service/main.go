package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kodokbakar/pylon/internal/config"
	internaltracing "github.com/kodokbakar/pylon/internal/tracing"
	notificationservice "github.com/kodokbakar/pylon/services/notification-service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	cfg.GRPC.Port = notificationPort()

	shutdownTracer, err := internaltracing.InitTracer(context.Background(), "notification-service", cfg.Tracing)
	if err != nil {
		log.Fatalf("init tracer: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := shutdownTracer(shutdownCtx); err != nil {
			log.Printf("shutdown tracer: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server, err := notificationservice.New(ctx, cfg)
	if err != nil {
		log.Fatalf("create notification service server: %v", err)
	}

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("start notification service server: %v", err)
		}
	case <-runCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("shutdown notification service server: %v", err)
		}
	}
}

func notificationPort() string {
	port := strings.TrimSpace(os.Getenv("NOTIFICATION_GRPC_PORT"))
	if port == "" {
		return "9004"
	}

	return port
}
