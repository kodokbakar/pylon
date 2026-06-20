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
	roomservice "github.com/kodokbakar/pylon/services/room-service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	cfg.GRPC.Port = roomPort()

	shutdownTracer, err := internaltracing.InitTracer(context.Background(), "room-service", cfg.Tracing)
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

	server, err := roomservice.New(ctx, cfg)
	if err != nil {
		log.Fatalf("create room service server: %v", err)
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
			log.Fatalf("start room service server: %v", err)
		}
	case <-runCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("shutdown room service server: %v", err)
		}
	}
}

func roomPort() string {
	port := strings.TrimSpace(os.Getenv("ROOM_GRPC_PORT"))
	if port == "" {
		return "9003"
	}

	return port
}
