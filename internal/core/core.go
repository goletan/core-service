package core

import (
	"context"
	"github.com/goletan/core-service/internal/metrics"
	"github.com/goletan/core-service/internal/orchestration"
	"github.com/goletan/core-service/internal/types"
	events "github.com/goletan/events-service/pkg"
	observability "github.com/goletan/observability-library/pkg"
	servicesLib "github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
	"log"
)

type Core struct {
	Config        *types.CoreConfig
	Observability *observability.Observability
	Metrics       *metrics.Metrics
	Orchestrator  *orchestration.Orchestrator
	EventsClient  *events.EventsClient
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

	return &Core{
		Config:        cfg,
		Observability: obs,
		Metrics:       met,
		Orchestrator:  orc,
		EventsClient:  newEventsClient,
	}, nil
}

// Start launches the core's core-service components and begins service discovery.
func (c *Core) Start(ctx context.Context) error {
	c.Observability.Logger.Info("Starting initial service orchestration...")
	err := c.Orchestrator.Orchestrate(ctx)
	if err != nil {
		c.Observability.Logger.Error("Failed to orchestrate services", zap.Error(err))
		return err
	}

	c.Observability.Logger.Info("Starting service discovery and event handling...")
	go c.startServiceWatcher(ctx)

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

// startServiceWatcher listens for service events-service and dynamically updates the service registry.
func (c *Core) startServiceWatcher(ctx context.Context) {
	if c.Config.Orchestrator.Filter == nil {
		c.Observability.Logger.Warn("No filter provided for service watcher")
		return
	}

	filter := &servicesTypes.Filter{Labels: c.Config.Orchestrator.Filter}
	eventCh, err := c.Orchestrator.Services.Watch(ctx, filter)
	if err != nil {
		c.Observability.Logger.Fatal("Failed to start service watcher", zap.Error(err))
		return
	}

	c.Observability.Logger.Info("Service watcher started")
	for {
		select {
		case <-ctx.Done():
			c.Observability.Logger.Info("Stopping service watcher...")
			return
		case event, ok := <-eventCh:
			if !ok {
				c.Observability.Logger.Warn("Service watcher channel closed")
				return
			}

			switch event.Type {
			case "ADDED":
				c.handleServiceAdded(event.Service, ctx)
			case "DELETED":
				c.handleServiceDeleted(event.Service, ctx)
			case "MODIFIED":
				c.handleServiceModified(event.Service, ctx)
			}
		}
	}
}

// handleServiceAdded dynamically registers and initializes a new service.
func (c *Core) handleServiceAdded(endpoint servicesTypes.ServiceEndpoint, ctx context.Context) {
	c.Observability.Logger.Info("Adding service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))
	svc, err := c.Orchestrator.Services.Register(endpoint)
	if err != nil {
		c.Observability.Logger.Error("Failed to register service", zap.String("name", endpoint.Name), zap.Error(err))
		return
	}

	if err = svc.Initialize(); err != nil {
		c.Observability.Logger.Error("Failed to initialize service", zap.String("name", svc.Name()), zap.Error(err))
		return
	}

	if err := svc.Start(ctx); err != nil {
		c.Observability.Logger.Error("Failed to start service", zap.String("name", svc.Name()), zap.Error(err))
	}
}

// handleServiceDeleted dynamically removes a service from the registry.
func (c *Core) handleServiceDeleted(endpoint servicesTypes.ServiceEndpoint, ctx context.Context) {
	c.Observability.Logger.Info("Removing service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))
	// Implementation for stopping and unregistering services-library if needed
}

// handleServiceModified handles updates to an existing service.
func (c *Core) handleServiceModified(endpoint servicesTypes.ServiceEndpoint, ctx context.Context) {
	c.Observability.Logger.Info("Modifying service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))
	// Implementation for updating services-library dynamically
}
