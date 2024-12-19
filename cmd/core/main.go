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

	// Set up core and services
	newCore, err := core.NewCore(shutdownCtx)
	if err != nil || newCore == nil {
		panic("Failed to create core")
	}

	// Initialize and start services
	initializeAndStartServices(shutdownCtx, newCore)

	// Wait for shutdown signal
	newCore.Obs.Logger.Info("Core Service is running...")
	<-shutdownCtx.Done()
	newCore.Obs.Logger.Info("Core Service shutting down...")
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

// initializeAndStartServices initializes and starts all services via the Core object.
func initializeAndStartServices(ctx context.Context, core *core.Core) {
	core.Obs.Logger.Info("Services are initializing...")
	if err := core.Services.InitializeAll(ctx); err != nil {
		core.Obs.Logger.Fatal("Failed to initialize services", zap.Error(err))
	}

	core.Obs.Logger.Info("Services are starting...")
	if err := core.Services.StartAll(ctx); err != nil {
		core.Obs.Logger.Fatal("Failed to start services", zap.Error(err))
	}
}
