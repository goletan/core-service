package orchestration

import (
	"context"
	"github.com/goletan/core-service/internal/health"
	"github.com/goletan/core-service/internal/orchestration/strategies"
	"github.com/goletan/core-service/internal/types"
	"github.com/goletan/core-service/internal/watcher"
	logger "github.com/goletan/logger-library/pkg"
	observability "github.com/goletan/observability-library/pkg"
	servicesLib "github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
	"time"
)

type Orchestrator struct {
	Logger         *logger.ZapLogger
	Config         *types.CoreConfig
	Services       *servicesLib.Services
	Strategy       strategies.Strategy
	ServiceWatcher *watcher.ServiceWatcher
	HealthMonitor  *health.Monitor
}

func NewOrchestrator(obs *observability.Observability, cfg *types.CoreConfig) (*Orchestrator, error) {
	newServices, err := servicesLib.NewServices(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to initialize services", zap.Error(err))
		return nil, err
	}
	obs.Logger.Info("Services initialized")

	strategy, err := strategies.NewStrategy(
		obs.Logger,
		newServices,
		cfg.Orchestrator.Strategy,
	)
	if err != nil {
		return nil, err
	}
	obs.Logger.Info("Orchestration strategy initialized")

	newServiceWatcher := watcher.NewServiceWatcher(obs.Logger, newServices)
	healthMonitor := health.NewMonitor(obs.Logger, newServices, 10*time.Second)

	return &Orchestrator{
		Logger:         obs.Logger,
		Config:         cfg,
		Services:       newServices,
		Strategy:       strategy,
		ServiceWatcher: newServiceWatcher,
		HealthMonitor:  healthMonitor,
	}, nil
}

func (o *Orchestrator) Orchestrate(ctx context.Context) error {
	filter := &servicesTypes.Filter{
		Labels: o.Config.Discovery.Filter.Labels,
		Tags:   o.Config.Discovery.Filter.Tags,
	}

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

	o.Logger.Info("Starting service discovery and event handling...")
	go o.ServiceWatcher.Start(ctx, filter)

	o.Logger.Info("Starting Health Monitor...")
	go o.HealthMonitor.Start(ctx)

	return nil
}
