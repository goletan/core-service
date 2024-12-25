package kernel

import (
	"context"
	"github.com/goletan/kernel-service/internal/types"
	observability "github.com/goletan/observability-library/pkg"
	resilience "github.com/goletan/resilience-library/pkg"
	resTypes "github.com/goletan/resilience-library/shared/types"
	services "github.com/goletan/services-library/pkg"
	serTypes "github.com/goletan/services-library/shared/types"
	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
	"log"
)

type Kernel struct {
	Config        *types.KernelConfig
	Observability *observability.Observability
	Resilience    *resilience.DefaultResilienceService
	Services      *services.Services
}

// NewKernel initializes the Kernel with essential components.
func NewKernel(ctx context.Context) (*Kernel, error) {
	obs := initializeObservability()
	obs.Logger.Info("Kernel Initializing...")

	cfg, err := LoadCoreConfig(obs.Logger)
	if err != nil {
		return nil, err
	}

	res := resilience.NewResilienceService(
		"kernel-service",
		obs,
		func(err error) bool { return true }, // Retry on all errors
		&resTypes.CircuitBreakerCallbacks{
			OnOpen: func(name string, from, to gobreaker.State) {
				obs.Logger.Warn("Circuit breaker opened", zap.String("name", name))
			},
			OnClose: func(name string, from, to gobreaker.State) {
				obs.Logger.Info("Circuit breaker closed", zap.String("name", name))
			},
		},
	)

	newServices, err := services.NewServices(obs)
	if err != nil {
		obs.Logger.Error("Failed to initialize services-library", zap.Error(err))
		return nil, err
	}

	return &Kernel{
		Config:        cfg,
		Observability: obs,
		Resilience:    res,
		Services:      newServices,
	}, nil
}

// Start launches the Kernel's kernel-service components and begins service discovery.
func (k *Kernel) Start(ctx context.Context) error {
	k.Observability.Logger.Info("Starting initial service orchestration...")
	orchestrateServices(ctx, k)

	k.Observability.Logger.Info("Starting service discovery and event handling...")
	go k.startServiceWatcher(ctx)

	return nil
}

// Shutdown gracefully stops the Kernel's components.
func (k *Kernel) Shutdown(ctx context.Context) error {
	k.Observability.Logger.Info("Shutting down Services...")
	if err := k.Services.StopAll(ctx); err != nil {
		k.Observability.Logger.Error("Failed to stop services-library", zap.Error(err))
	}

	k.Observability.Logger.Info("Shutting down Resilience...")
	if err := k.Resilience.Shutdown(&ctx); err != nil {
		k.Observability.Logger.Error("Failed to shut down resilience-library", zap.Error(err))
		return err
	}

	k.Observability.Logger.Info("Kernel shut down successfully")
	return nil
}

// startServiceWatcher listens for service events-service and dynamically updates the service registry.
func (k *Kernel) startServiceWatcher(ctx context.Context) {
	eventCh, err := k.Services.Watch(ctx, "default-namespace")
	if err != nil {
		k.Observability.Logger.Fatal("Failed to start service watcher", zap.Error(err))
		return
	}

	for {
		select {
		case <-ctx.Done():
			k.Observability.Logger.Info("Stopping service watcher...")
			return
		case event, ok := <-eventCh:
			if !ok {
				k.Observability.Logger.Warn("Service watcher channel closed")
				return
			}

			switch event.Type {
			case "ADDED":
				k.handleServiceAdded(event.Service)
			case "DELETED":
				k.handleServiceDeleted(event.Service)
			case "MODIFIED":
				k.handleServiceModified(event.Service)
			}
		}
	}
}

// handleServiceAdded dynamically registers and initializes a new service.
func (k *Kernel) handleServiceAdded(endpoint serTypes.ServiceEndpoint) {
	k.Observability.Logger.Info("Adding service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))
	service, err := k.Services.CreateService(endpoint)
	if err != nil {
		k.Observability.Logger.Error("Failed to create service", zap.String("name", endpoint.Name), zap.Error(err))
		return
	}

	if err := k.Services.Register(service); err != nil {
		k.Observability.Logger.Error("Failed to register service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	if err := service.Initialize(); err != nil {
		k.Observability.Logger.Error("Failed to initialize service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	if err := service.Start(); err != nil {
		k.Observability.Logger.Error("Failed to start service", zap.String("name", service.Name()), zap.Error(err))
	}
}

// handleServiceDeleted dynamically removes a service from the registry.
func (k *Kernel) handleServiceDeleted(endpoint serTypes.ServiceEndpoint) {
	k.Observability.Logger.Info("Removing service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))
	// Implementation for stopping and unregistering services-library if needed
}

// handleServiceModified handles updates to an existing service.
func (k *Kernel) handleServiceModified(endpoint serTypes.ServiceEndpoint) {
	k.Observability.Logger.Info("Modifying service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))
	// Implementation for updating services-library dynamically
}

// initializeObservability initializes the observability-library components (logger-library, metrics, tracing).
func initializeObservability() *observability.Observability {
	obs, err := observability.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observability-library", err)
	}
	return obs
}
