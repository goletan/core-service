package main

import (
	"context"
	"core/internal/core"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	// Set up signal handling for shutdown
	setupSignalHandler(shutdownCancel)

	// Set up core-service and services-library
	newCore, err := core.NewCore(shutdownCtx)
	if err != nil || newCore == nil {
		panic("Failed to create core-service")
	}

	// Initialize and start services-library
	initializeAndStartServices(shutdownCtx, newCore)

	// Wait for shutdown signal
	newCore.Observability.Logger.Info("Core Service is running...")
	<-shutdownCtx.Done()
	newCore.Observability.Logger.Info("Core Service shutting down...")
}

// setupSignalHandler configures OS signal handling for graceful shutdown.
func setupSignalHandler(cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		cancelFunc() // Trigger shutdown
	}()
}

// initializeAndStartServices initializes and starts all services-library via the Core object.
func initializeAndStartServices(ctx context.Context, core *core.Core) {
	core.Observability.Logger.Info("Services are initializing...")
	if err := core.Services.InitializeAll(ctx); err != nil {
		core.Observability.Logger.Fatal("Failed to initialize services-library", zap.Error(err))
	}

	core.Observability.Logger.Info("Services are starting...")
	if err := core.Services.StartAll(ctx); err != nil {
		core.Observability.Logger.Fatal("Failed to start services-library", zap.Error(err))
	}
}
