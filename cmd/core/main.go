package main

import (
	"context"
	"core/internal/core"
	"fmt"
	"github.com/goletan/observability/shared/errors"
	"github.com/goletan/observability/shared/logger"
	services "github.com/goletan/services/pkg"
	"github.com/goletan/services/shared/types"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/goletan/observability/pkg"
)

func main() {
	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		cancel() // Trigger shutdown
	}()

	zapLogger, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	// Load configuration
	obsConfig, err := core.LoadObservabilityConfig(zapLogger)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Observability
	obs, err := observability.NewObserver(obsConfig, zapLogger)
	if err != nil {
		log.Fatal(err)
	}

	obs.Logger.Info("Core Service initializing...")

	// Initialize Services.
	serviceManager := services.NewServices(obs)

	// Create a static service
	staticService := NewStaticService("TestService", []types.ServiceEndpoint{
		{
			Name:    "TestServiceEndpoint",
			Address: "localhost:8080",
			Ports: []types.ServicePort{
				{Name: "http", Port: 8080, Protocol: "TCP"},
			},
			Version: "v1.0",
			Tags:    []string{"testing", "static"},
		},
	})

	// Register services.
	err = serviceManager.Registry.Register(staticService)
	if err != nil {
		obs.Logger.Fatal("Failed to register services", zap.Error(err))
	}

	// Start services.
	if err := serviceManager.InitializeAll(ctx); err != nil {
		obs.Logger.Error("Failed to initialize services", zap.Error(err))
	}
	if err := serviceManager.StartAll(ctx); err != nil {
		obs.Logger.Fatal("Failed to start services", zap.Error(err))
	}

	obs.Logger.Info("Core Service is running...")
	<-ctx.Done() // Wait for shutdown signal

	obs.Logger.Info("Core Service shutting down...")
}

// StaticService is a simple implementation of the Service interface for testing.
type StaticService struct {
	name      string
	endpoints []types.ServiceEndpoint
}

// NewStaticService creates a new instance of StaticService.
func NewStaticService(name string, endpoints []types.ServiceEndpoint) *StaticService {
	return &StaticService{
		name:      name,
		endpoints: endpoints,
	}
}

// Name returns the name of the service.
func (s *StaticService) Name() string {
	return s.name
}

// Initialize prepares the service for startup.
func (s *StaticService) Initialize() error {
	fmt.Printf("[%s] Initializing...\n", s.name)
	// Simulate initialization (e.g., loading configuration)
	return nil
}

// Start starts the service.
func (s *StaticService) Start() error {
	fmt.Printf("[%s] Starting...\n", s.name)
	// Simulate service logic startup
	return nil
}

// Stop stops the service.
func (s *StaticService) Stop() error {
	fmt.Printf("[%s] Stopping...\n", s.name)
	// Simulate cleanup tasks
	return nil
}

// Discover simulates the discovery of service endpoints.
func (s *StaticService) Discover(log *logger.ZapLogger) ([]types.ServiceEndpoint, error) {
	if len(s.endpoints) == 0 {
		return nil, errors.NewError(log, "No endpoints found", 1001, nil)
	}

	return s.endpoints, nil
}
