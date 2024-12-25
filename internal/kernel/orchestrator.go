package kernel

import (
	"context"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
)

// orchestrateServices performs initial discovery and setup of services-library.
func orchestrateServices(ctx context.Context, kernel *Kernel) {
	namespace := "goletan"
	kernel.Observability.Logger.Info("Performing initial service discovery...", zap.String("namespace", namespace))

	// Discover services-library
	endpoints, err := discoverServices(ctx, kernel, namespace)
	if err != nil {
		kernel.Observability.Logger.Error("Initial service discovery failed", zap.Error(err))
		return
	}

	// Initialize and start services-library
	for _, endpoint := range endpoints {
		initializeAndStartService(ctx, kernel, endpoint)
	}

	kernel.Observability.Logger.Info("Initial service discovery and orchestration completed")
}

// discoverServices performs a one-time discovery of available services-library in the namespace.
func discoverServices(ctx context.Context, kernel *Kernel, namespace string) ([]servicesTypes.ServiceEndpoint, error) {
	var endpoints []servicesTypes.ServiceEndpoint

	err := kernel.Resilience.ExecuteWithRetry(ctx, func() error {
		var err error
		endpoints, err = kernel.Services.Discover(ctx, namespace)
		return err
	})
	if err != nil {
		return nil, err
	}

	kernel.Observability.Logger.Info("Discovered initial services-library",
		zap.Int("count", len(endpoints)))
	return endpoints, nil
}

// initializeAndStartService handles the lifecycle setup for a single service.
func initializeAndStartService(ctx context.Context, kernel *Kernel, endpoint servicesTypes.ServiceEndpoint) {
	kernel.Observability.Logger.Info("Initializing service", zap.String("name", endpoint.Name))

	// Create a Service instance from the endpoint
	service, err := kernel.Services.CreateService(endpoint)
	if err != nil {
		kernel.Observability.Logger.Error("Failed to create service", zap.String("name", endpoint.Name), zap.Error(err))
		return
	}

	// Register the Service
	if err := kernel.Services.Register(service); err != nil {
		kernel.Observability.Logger.Error("Failed to register service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	// Resilient Initialization
	if err := kernel.Resilience.ExecuteWithRetry(ctx, func() error {
		return service.Initialize()
	}); err != nil {
		kernel.Observability.Logger.Error("Failed to initialize service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	// Resilient Start
	if err := kernel.Resilience.ExecuteWithRetry(ctx, func() error {
		return service.Start()
	}); err != nil {
		kernel.Observability.Logger.Error("Failed to start service", zap.String("name", service.Name()), zap.Error(err))
	}
}
