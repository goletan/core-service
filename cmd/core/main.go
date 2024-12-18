package main

import (
	"context"
	"core/internal/core"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/goletan/observability/pkg"
)

func main() {
	// Context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	// Set up signal handling for shutdown
	setupSignalHandler(shutdownCancel)

	// Initialize observability
	obs := initializeObservability()
	obs.Logger.Info("Core Initializing...")

	// Set up core and services
	newCore, err := core.NewCore(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to initialize core", zap.Error(err))
	}
	if newCore == nil {
		obs.Logger.Fatal("Failed to initialize core", zap.Error(err))
	}

	// Initialize and start services
	initializeAndStartServices(shutdownCtx, newCore, obs)
	
	// Wait for shutdown signal
	obs.Logger.Info("Core Service is running...")
	<-shutdownCtx.Done()
	obs.Logger.Info("Core Service shutting down...")
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

// initializeObservability initializes the observability components (logger, metrics, tracing).
func initializeObservability() *observability.Observability {
	obs, err := observability.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observability", err)
	}
	return obs
}

// initializeAndStartServices initializes and starts all services via the Core object.
func initializeAndStartServices(ctx context.Context, core *core.Core, obs *observability.Observability) {
	obs.Logger.Info("Services are initializing...")
	if err := core.Services.InitializeAll(ctx); err != nil {
		obs.Logger.Fatal("Failed to initialize services", zap.Error(err))
	}
	obs.Logger.Info("Services are starting...")
	if err := core.Services.StartAll(ctx); err != nil {
		obs.Logger.Fatal("Failed to start services", zap.Error(err))
	}
}
