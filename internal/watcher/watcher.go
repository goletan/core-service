package watcher

import (
	"context"
	logger "github.com/goletan/logger-library/pkg"
	servicesLib "github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
)

// ServiceWatcher dynamically monitors service changes at runtime.
type ServiceWatcher struct {
	Logger   *logger.ZapLogger
	Services *servicesLib.Services
}

// NewServiceWatcher initializes the service watcher.
func NewServiceWatcher(logger *logger.ZapLogger, services *servicesLib.Services) *ServiceWatcher {
	return &ServiceWatcher{
		Logger:   logger,
		Services: services,
	}
}

// Start begins watching for service events (add, modify, delete).
func (sw *ServiceWatcher) Start(ctx context.Context, filter *servicesTypes.Filter) {
	if filter == nil {
		sw.Logger.Warn("No filter provided for service watcher")
		return
	}

	eventCh, err := sw.Services.Watch(ctx, filter)
	if err != nil {
		sw.Logger.Fatal("Failed to start service watcher", zap.Error(err))
		return
	}

	sw.Logger.Info("Service watcher started")
	for {
		select {
		case <-ctx.Done():
			sw.Logger.Info("Stopping service watcher...")
			return
		case event, ok := <-eventCh:
			if !ok {
				sw.Logger.Warn("Service watcher channel closed")
				return
			}

			sw.HandleEvent(event, ctx)
		}
	}
}

// HandleEvent processes service events dynamically.
func (sw *ServiceWatcher) HandleEvent(event servicesTypes.ServiceEvent, ctx context.Context) {
	switch event.Type {
	case "ADDED":
		sw.handleServiceAdded(event.Service, ctx)
	case "DELETED":
		sw.handleServiceDeleted(event.Service, ctx)
	case "MODIFIED":
		sw.handleServiceModified(event.Service, ctx)
	default:
		sw.Logger.Warn("Unknown service event type", zap.String("type", event.Type))
	}
}

// handleServiceAdded dynamically registers and starts new services.
func (sw *ServiceWatcher) handleServiceAdded(endpoint servicesTypes.ServiceEndpoint, ctx context.Context) {
	sw.Logger.Info("Adding service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))

	svc, err := sw.Services.Register(endpoint)
	if err != nil {
		sw.Logger.Error("Failed to register service", zap.String("name", endpoint.Name), zap.Error(err))
		return
	}

	if err = svc.Initialize(); err != nil {
		sw.Logger.Error("Failed to initialize service", zap.String("name", svc.Name()), zap.Error(err))
		return
	}

	if err := svc.Start(ctx); err != nil {
		sw.Logger.Error("Failed to start service", zap.String("name", svc.Name()), zap.Error(err))
	}
}

// handleServiceDeleted removes a service dynamically.
func (sw *ServiceWatcher) handleServiceDeleted(endpoint servicesTypes.ServiceEndpoint, ctx context.Context) {
	sw.Logger.Info("Removing service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))

	// TODO: Implement proper stopping and unregistering logic
}

// handleServiceModified handles service updates.
func (sw *ServiceWatcher) handleServiceModified(endpoint servicesTypes.ServiceEndpoint, ctx context.Context) {
	sw.Logger.Info("Modifying service", zap.String("name", endpoint.Name), zap.String("address", endpoint.Address))

	// TODO: Implement service update handling
}
