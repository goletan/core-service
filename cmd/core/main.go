package main

import (
	"context"
	"github.com/goletan/core-service/internal/core"
	eventsTypes "github.com/goletan/events-library/shared/types"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	initializeAndStartServices(shutdownCtx, newCore)

	event := &eventsTypes.Event{
		Type:    "test_event_types",
		Payload: "test_event_payload",
	}

	err = newCore.EventsClient.SendEvent(shutdownCtx, event)
	if err != nil {
		newCore.Observability.Logger.Error("Failed to send event", zap.Error(err))
		return
	}

	// Wait for shutdown signal
	newCore.Observability.Logger.Info("core Service is running...")
	<-shutdownCtx.Done()
	newCore.Observability.Logger.Info("core Service shutting down...")
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

// initializeAndStartServices initializes and starts all services-library via the core object.
func initializeAndStartServices(ctx context.Context, core *core.Core) {

	serviceDiscoveryTimeout := 5 * time.Second
	discoverCtx, discoverCancel := context.WithTimeout(ctx, serviceDiscoveryTimeout)
	defer discoverCancel()

	serviceEndpoints, err := core.Services.Discover(discoverCtx, "goletan_services_network")
	if err != nil {
		core.Observability.Logger.Warn("No services discovered on goletan_services_network", zap.Error(err))
	} else {
		for _, endpoint := range serviceEndpoints {
			core.Observability.Logger.Info("Discovered service",
				zap.String("name", endpoint.Name),
				zap.String("address", endpoint.Address))
		}
	}

	core.Observability.Logger.Info("Services are initializing...")
	if err := core.Services.InitializeAll(discoverCtx); err != nil {
		core.Observability.Logger.Fatal("Failed to initialize services-library", zap.Error(err))
	}

	core.Observability.Logger.Info("Services are starting...")
	if err := core.Services.StartAll(discoverCtx); err != nil {
		core.Observability.Logger.Fatal("Failed to start services-library", zap.Error(err))
	}
}
