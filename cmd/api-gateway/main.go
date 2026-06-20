package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kodokbakar/pylon/internal/config"
	internaltracing "github.com/kodokbakar/pylon/internal/tracing"
	apigateway "github.com/kodokbakar/pylon/services/api-gateway"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	shutdownTracer, err := internaltracing.InitTracer(context.Background(), "api-gateway", cfg.Tracing)
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

	server, err := apigateway.New(cfg)
	if err != nil {
		log.Fatalf("create api gateway server: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("start api gateway server: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("shutdown api gateway server: %v", err)
		}
	}
}
