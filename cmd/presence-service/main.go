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
	presenceservice "github.com/kodokbakar/pylon/services/presence-service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	cfg.GRPC.Port = presencePort()

	shutdownTracer, err := internaltracing.InitTracer(context.Background(), "presence-service", cfg.Tracing)
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

	server, err := presenceservice.New(ctx, cfg)
	if err != nil {
		log.Fatalf("create presence service server: %v", err)
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
			log.Fatalf("start presence service server: %v", err)
		}
	case <-runCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("shutdown presence service server: %v", err)
		}
	}
}

func presencePort() string {
	port := strings.TrimSpace(os.Getenv("PRESENCE_GRPC_PORT"))
	if port == "" {
		return "9002"
	}

	return port
}
