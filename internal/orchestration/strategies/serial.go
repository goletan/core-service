package strategies

import (
	"context"
	logger "github.com/goletan/logger-library/pkg"
	services "github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
)

type SerialStrategy struct {
	Logger   *logger.ZapLogger
	Services *services.Services
}

func NewSerialStrategy(logger *logger.ZapLogger, services *services.Services) *SerialStrategy {
	return &SerialStrategy{
		Logger:   logger,
		Services: services,
	}
}

func (s *SerialStrategy) Orchestrate(ctx context.Context, endpoints []servicesTypes.ServiceEndpoint) error {
	for _, endpoint := range endpoints {
		s.Logger.Info("Orchestrating service in serial mode", zap.String("name", endpoint.Name))

		service, err := s.Services.Register(endpoint)
		if err != nil {
			return err
		}
		s.Logger.Info("Service registered", zap.String("name", endpoint.Name))

		err = service.Initialize()
		if err != nil {
			return err
		}
		s.Logger.Info("Service initialized", zap.String("name", endpoint.Name))

		err = service.Start(ctx)
		if err != nil {
			return err
		}
		s.Logger.Info("Service started", zap.String("name", endpoint.Name))

	}

	return nil
}
