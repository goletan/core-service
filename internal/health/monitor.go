package health

import (
	"context"
	logger "github.com/goletan/logger-library/pkg"
	"github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
	"time"
)

// Monitor keeps services running by detecting failures.
type Monitor struct {
	Logger   *logger.ZapLogger
	Services *services.Services
	Interval time.Duration
}

// NewMonitor initializes the monitor.
func NewMonitor(logger *logger.ZapLogger, services *services.Services, interval time.Duration) *Monitor {
	return &Monitor{
		Logger:   logger,
		Services: services,
		Interval: interval,
	}
}

// Start begins periodic health checks and auto-healing.
func (m *Monitor) Start(ctx context.Context) {
	m.Logger.Info("Health Monitor started", zap.Duration("interval", m.Interval))
	ticker := time.NewTicker(m.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.Logger.Info("Stopping Health Monitor...")
			return
		case <-ticker.C:
			m.checkAndRecoverServices(ctx)
		}
	}
}

// checkAndRecoverServices checks service health and applies recovery if needed.
func (m *Monitor) checkAndRecoverServices(ctx context.Context) {
	m.Logger.Info("Performing health checks on services...")
	for _, service := range m.Services.List() {
		status := m.checkServiceHealth(ctx, service)
		m.Logger.Info("Service health check result", zap.String("name", service.Name()), zap.String("status", status))

		switch status {
		case "HEALTHY":
			m.Logger.Info("Service is healthy", zap.String("name", service.Name()))
		case "DEGRADED":
			m.Logger.Warn("Service is degraded, considering restart", zap.String("name", service.Name()))
			m.attemptRestart(ctx, service)
		case "FAILED":
			m.Logger.Error("Service is failing, restarting...", zap.String("name", service.Name()))
			m.forceRestart(ctx, service)
		}
	}
}

// checkServiceHealth runs service-specific health checks.
func (m *Monitor) checkServiceHealth(ctx context.Context, service servicesTypes.Service) string {
	// Placeholder: In the future, check latency, resource usage, custom health endpoints, etc.
	if err := service.Initialize(); err != nil {
		return "FAILED"
	}

	return "HEALTHY" // This should be enhanced with real health signals.
}

// attemptRestart tries a soft restart if the service is degraded.
func (m *Monitor) attemptRestart(ctx context.Context, service servicesTypes.Service) {
	if err := service.Stop(ctx); err != nil {
		m.Logger.Error("Failed to stop service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	if err := service.Start(ctx); err != nil {
		m.Logger.Error("Failed to restart service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	m.Logger.Info("Service restarted successfully", zap.String("name", service.Name()))
}

// forceRestart forcefully restarts the service if it's completely unresponsive.
func (m *Monitor) forceRestart(ctx context.Context, service servicesTypes.Service) {
	m.Logger.Warn("Force restarting service", zap.String("name", service.Name()))

	if err := service.Stop(ctx); err != nil {
		m.Logger.Error("Failed to stop failing service", zap.String("name", service.Name()), zap.Error(err))
	}

	// Unregister service before re-registering
	if err := m.Services.Unregister(service.Name()); err != nil {
		m.Logger.Error("Failed to unregister service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	endpoint := servicesTypes.ServiceEndpoint{
		Name:    service.Name(),
		Address: service.Address(),
	}

	// Register the service again
	newService, err := m.Services.Register(endpoint)
	if err != nil {
		m.Logger.Error("Failed to re-register service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	if err := newService.Start(ctx); err != nil {
		m.Logger.Error("Failed to restart service", zap.String("name", service.Name()), zap.Error(err))
		return
	}

	m.Logger.Info("Service restarted successfully", zap.String("name", service.Name()))
}
