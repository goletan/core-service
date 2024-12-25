package main

import (
	"context"
	"github.com/goletan/kernel-service/internal/kernel"
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

	// Set up kernel-service and services-library
	newKernel, err := kernel.NewKernel(shutdownCtx)
	if err != nil || newKernel == nil {
		panic("Failed to create kernel-service")
	}

	// Initialize and start services-library
	initializeAndStartServices(shutdownCtx, newKernel)

	// Wait for shutdown signal
	newKernel.Observability.Logger.Info("Kernel Service is running...")
	<-shutdownCtx.Done()
	newKernel.Observability.Logger.Info("Kernel Service shutting down...")
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

// initializeAndStartServices initializes and starts all services-library via the Kernel object.
func initializeAndStartServices(ctx context.Context, kernel *kernel.Kernel) {
	kernel.Observability.Logger.Info("Services are initializing...")
	if err := kernel.Services.InitializeAll(ctx); err != nil {
		kernel.Observability.Logger.Fatal("Failed to initialize services-library", zap.Error(err))
	}

	kernel.Observability.Logger.Info("Services are starting...")
	if err := kernel.Services.StartAll(ctx); err != nil {
		kernel.Observability.Logger.Fatal("Failed to start services-library", zap.Error(err))
	}
}
