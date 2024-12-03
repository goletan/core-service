package core

import (
	"context"
	"github.com/goletan/observability"
	"github.com/goletan/security/config"
	services "github.com/goletan/services/pkg"

	"go.uber.org/zap"
)

type Core struct {
	Config        *CoreConfig
	Observability *observability.Observability
	Services      *services.Service
}

// NewCore initializes the Core with essential components.
func NewCore(obs *observability.Observability) (*Core, error) {
	cfg, err := LoadCoreConfig(logger)
	if err != nil {
		return nil, err
	}

	srv, err := services.NewServices(obs)
	if err != nil {
		return nil, err
	}

	return &Core{
		Config:        cfg,
		Observability: obs,
		Services:      srv,
	}, nil
}

// Start launches the Core's core components.
func (c *Core) Start(ctx context.Context) error {
	c.Observability.Logger.Info("Starting Core...")
	return c.Services.Start(ctx)
}

// Shutdown gracefully stops the Core's components.
func (c *Core) Shutdown(ctx context.Context) error {
	c.Observability.Info("Shutting down Core...")
	return c.Services.Stop(ctx)
}
