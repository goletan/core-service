package orchestration

import (
	"context"
	"github.com/goletan/core-service/internal/orchestration/strategies"
	"github.com/goletan/core-service/internal/types"
	logger "github.com/goletan/logger-library/pkg"
	observability "github.com/goletan/observability-library/pkg"
	servicesLib "github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
)

type Orchestrator struct {
	Logger   *logger.ZapLogger
	Config   *types.CoreConfig
	Services *servicesLib.Services
	Strategy strategies.Strategy
}

func NewOrchestrator(obs *observability.Observability, cfg *types.CoreConfig, services *servicesLib.Services) (*Orchestrator, error) {
	// Extract strategy from config
	strategy, err := strategies.NewStrategy(
		obs.Logger,
		services,
		cfg.Orchestrator.Strategy,
	)
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		Logger:   obs.Logger,
		Config:   cfg,
		Services: services,
		Strategy: strategy,
	}, nil
}

func (o *Orchestrator) Orchestrate(ctx context.Context, filter *servicesTypes.Filter) error {
	endpoints, err := o.Services.Discover(ctx, filter)
	if err != nil {
		o.Logger.Error("Failed to discover services", zap.Error(err))
		return err
	}

	if err := o.Strategy.Orchestrate(ctx, endpoints); err != nil {
		o.Logger.Error("Orchestration failed", zap.Error(err))
		return err
	}

	o.Logger.Info("Orchestration completed successfully")
	return nil
}
