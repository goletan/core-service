package core

import (
	"context"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
)

// orchestrateServices performs initial discovery and setup of services-library.
func orchestrateServices(ctx context.Context, core *Core) {
	namespace := "goletan"
	core.Observability.Logger.Info("Performing initial service discovery...", zap.String("namespace", namespace))

	// Discover services-library
	endpoints, err := discoverServices(ctx, core, namespace)
	if err != nil {
		core.Observability.Logger.Error("Initial service discovery failed", zap.Error(err))
		return
	}

	// Initialize and start services-library
	for _, endpoint := range endpoints {
		initializeAndStartService(ctx, core, endpoint)
	}

	core.Observability.Logger.Info("Initial service discovery and orchestration completed")
}

// discoverServices performs a one-time discovery of available services-library in the namespace.
func discoverServices(ctx context.Context, core *Core, namespace string) ([]servicesTypes.ServiceEndpoint, error) {
	var endpoints []servicesTypes.ServiceEndpoint

	err := core.Resilience.ExecuteWithRetry(ctx, func() error {
		var err error
		endpoints, err = core.Services.Discover(ctx, namespace)
		return err
	})
	if err != nil {
		return nil, err
	}

	core.Observability.Logger.Info("Discovered initial services-library",
		zap.Int("count", len(endpoints)))
	return endpoints, nil
}

// initializeAndStartService handles the lifecycle setup for a single service.
func initializeAndStartService(ctx context.Context, core *Core, endpoint servicesTypes.ServiceEndpoint) {
	core.Observability.Logger.Info("Initializing service", zap.String("name", endpoint.Name))

	// Create a Service instance from the endpoint
	service, err := core.Services.CreateService(endpoint)
	if err != nil {
		core.Observability.Logger.Error("Failed to create service", zap.String("name", endpoint.Name), zap.Error(err))
		return
	}

	// Register the Service
	if err := core.Services.Register(service); err != nil {
		core.Observability.Logger.Error("Failed to register service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	// Resilient Initialization
	if err := core.Resilience.ExecuteWithRetry(ctx, func() error {
		return service.Initialize()
	}); err != nil {
		core.Observability.Logger.Error("Failed to initialize service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	// Resilient Start
	if err := core.Resilience.ExecuteWithRetry(ctx, func() error {
		return service.Start()
	}); err != nil {
		core.Observability.Logger.Error("Failed to start service", zap.String("name", service.Name()), zap.Error(err))
	}
}
