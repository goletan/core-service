package core

import (
	"context"
	"core/internal/types"
	observability "github.com/goletan/observability/pkg"
	services "github.com/goletan/services/pkg"
	"go.uber.org/zap"
	"log"
)

type Core struct {
	Config   *types.CoreConfig
	Services *services.Services
	Obs      *observability.Observability
}

// NewCore initializes the Core with essential components.
func NewCore(ctx context.Context) (*Core, error) {
	obs := initializeObservability()
	obs.Logger.Info("Core Initializing...")

	cfg, err := LoadCoreConfig(obs.Logger)
	if err != nil {
		return nil, err
	}

	newServices := services.NewServices(obs) // Pass context to Services

	return &Core{
		Config:   cfg,
		Obs:      obs,
		Services: newServices,
	}, nil
}

// Start launches the Core's core components.
func (c *Core) Start(ctx context.Context) error {
	c.Obs.Logger.Info("Starting Services...")
	return c.Services.StartAll(ctx)
}

// Shutdown gracefully stops the Core's components.
func (c *Core) Shutdown(ctx context.Context) error {
	c.Obs.Logger.Info("Shutting down Services...")

	if err := c.Services.StopAll(ctx); err != nil {
		c.Obs.Logger.Error("Error during shutdown", zap.Error(err))
		return err
	}

	c.Obs.Logger.Info("All Services shut down successfully.")
	return nil
}

// initializeObservability initializes the observability components (logger, metrics, tracing).
func initializeObservability() *observability.Observability {
	obs, err := observability.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observability", err)
	}
	return obs
}
