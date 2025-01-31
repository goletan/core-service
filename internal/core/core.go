package core

import (
	"context"
	"github.com/goletan/core-service/internal/metrics"
	"github.com/goletan/core-service/internal/orchestration"
	"github.com/goletan/core-service/internal/types"
	"github.com/goletan/core-service/internal/watcher"
	events "github.com/goletan/events-service/pkg"
	observability "github.com/goletan/observability-library/pkg"
	servicesLib "github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
	"log"
)

type Core struct {
	Config         *types.CoreConfig
	Observability  *observability.Observability
	Metrics        *metrics.Metrics
	Orchestrator   *orchestration.Orchestrator
	EventsClient   *events.EventsClient
	ServiceWatcher *watcher.ServiceWatcher
}

// NewCore initializes the core with essential components.
func NewCore() (*Core, error) {
	obs, err := observability.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observability", err)
	}

	obs.Logger.Info("Core booting...")

	cfg, err := LoadCoreConfig(obs.Logger)
	if err != nil {
		return nil, err
	}
	obs.Logger.Info("Core configuration loaded")

	newServices, err := servicesLib.NewServices(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to initialize services", zap.Error(err))
		return nil, err
	}
	obs.Logger.Info("Services initialized")

	newEventsClient, err := events.NewEventsClient(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to initialize events client", zap.Error(err))
		return nil, err
	}
	obs.Logger.Info("Events Client initialized")

	orc, err := orchestration.NewOrchestrator(obs, cfg, newServices)
	if err != nil {
		obs.Logger.Fatal("Failed to initialize orchestrator", zap.Error(err))
		return nil, err
	}
	obs.Logger.Info("Orchestrator initialized")

	met := metrics.InitMetrics(obs)
	obs.Logger.Info("Metrics initialized")

	newServiceWatcher := watcher.NewServiceWatcher(obs.Logger, newServices)

	return &Core{
		Config:         cfg,
		Observability:  obs,
		Metrics:        met,
		Orchestrator:   orc,
		EventsClient:   newEventsClient,
		ServiceWatcher: newServiceWatcher,
	}, nil
}

// Start launches the core's core-service components and begins service discovery.
func (c *Core) Start(ctx context.Context) error {
	filter := &servicesTypes.Filter{Labels: c.Config.Orchestrator.Filter}

	c.Observability.Logger.Info("Starting initial service orchestration...")
	err := c.Orchestrator.Orchestrate(ctx, filter)
	if err != nil {
		c.Observability.Logger.Error("Failed to orchestrate services", zap.Error(err))
		return err
	}

	c.Observability.Logger.Info("Starting service discovery and event handling...")
	go c.ServiceWatcher.Start(ctx, filter)

	return nil
}

// Shutdown gracefully stops the core's components.
func (c *Core) Shutdown(ctx context.Context) error {
	c.Observability.Logger.Info("Shutting down Services...")
	if err := c.Orchestrator.Services.StopAll(ctx); err != nil {
		c.Observability.Logger.Error("Failed to stop services", zap.Error(err))
	}

	c.Observability.Logger.Info("All services shut down gracefully")
	return nil
}
