package main

import (
	"context"
	"github.com/goletan/core-service/internal/core"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	// configures OS signal handling for graceful shutdown.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		shutdownCancel()
	}()

	// Set up core-service and services-library
	newCore, err := core.NewCore()
	if err != nil || newCore == nil {
		log.Fatal("Failed to create core service", err)
		return
	}

	err = newCore.Start(shutdownCtx)
	if err != nil {
		newCore.Observability.Logger.Fatal("Failed to start core-service", zap.Error(err))
		return
	}

	// Wait for shutdown signal
	newCore.Observability.Logger.Info("core Service is running...")
	<-shutdownCtx.Done()
	newCore.Observability.Logger.Info("core Service shutting down...")
	err = newCore.Shutdown(shutdownCtx)
	if err != nil {
		newCore.Observability.Logger.Error("Failed to shutdown core-service", zap.Error(err))
		return
	}
}
