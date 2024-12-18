package core

import (
	"context"
	"core/internal/types"
	observability "github.com/goletan/observability/pkg"
	services "github.com/goletan/services/pkg"
	"log"
)

type Core struct {
	Config   *types.CoreConfig
	Services *services.Services
	Obs      *observability.Observability
}

// NewCore initializes the Core with essential components.
func NewCore() (*Core, error) {
	obs := initializeObservability()
	obs.Logger.Info("Core Initializing...")

	cfg, err := LoadCoreConfig(obs.Logger)
	if err != nil {
		return nil, err
	}

	newServices := services.NewServices(obs)

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
	return c.Services.StopAll(ctx)
}

// initializeObservability initializes the observability components (logger, metrics, tracing).
func initializeObservability() *observability.Observability {
	obs, err := observability.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observability", err)
	}
	return obs
}
