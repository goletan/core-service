package strategies

import (
	"context"
	logger "github.com/goletan/logger-library/pkg"
	services "github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
)

type Strategy interface {
	Orchestrate(ctx context.Context, endpoints []servicesTypes.ServiceEndpoint) error
}

func NewStrategy(logger *logger.ZapLogger, services *services.Services, strategyType string) (Strategy, error) {
	if strategyType == "" {
		logger.Warn("No strategy type provided; defaulting to serial.")
		return NewSerialStrategy(logger, services), nil
	}

	switch strategyType {
	case "serial":
		return NewSerialStrategy(logger, services), nil
	case "parallel":
		return NewParallelStrategy(logger, services), nil
	case "hybrid":
		return NewHybridStrategy(logger, services), nil
	default:
		logger.Warn("No strategy type provided; defaulting to serial.")
		return NewSerialStrategy(logger, services), nil
	}
}

// logServiceOrchestration logs consistent service orchestration messages.
func logServiceOrchestration(logger *logger.ZapLogger, mode string, endpoint servicesTypes.ServiceEndpoint, priority int) {
	logger.Info("Orchestrating service",
		zap.String("mode", mode),
		zap.String("service", endpoint.Name),
		zap.Int("priority", priority),
	)
}
