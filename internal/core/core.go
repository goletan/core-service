package core

import (
	"context"
	"core/internal/metrics"
	"core/internal/types"
	"github.com/goletan/observability/pkg"
	"github.com/goletan/services/pkg"
)

type Core struct {
	Config        *types.CoreConfig
	Observability *observability.Observability
	Services      *services.Services
}

// NewCore initializes the Core with essential components.
func NewCore(obs *observability.Observability) (*Core, error) {
	cfg, err := LoadCoreConfig(obs)
	if err != nil {
		return nil, err
	}

	metrics.InitMetrics(obs)

	newServices := services.NewServices(obs)

	return &Core{
		Config:        cfg,
		Observability: obs,
		Services:      newServices,
	}, nil
}

// Start launches the Core's core components.
func (c *Core) Start(ctx context.Context) error {
	c.Observability.Logger.Info("Starting Services...")
	return c.Services.StartAll(ctx)
}

// Shutdown gracefully stops the Core's components.
func (c *Core) Shutdown(ctx context.Context) error {
	c.Observability.Logger.Info("Shutting down Services...")
	return c.Services.StopAll(ctx)
}
