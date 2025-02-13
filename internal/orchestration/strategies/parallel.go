package strategies

import (
	"context"
	"fmt"
	logger "github.com/goletan/logger-library/pkg"
	services "github.com/goletan/services-library/pkg"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
	"sync"
)

type ParallelStrategy struct {
	Logger   *logger.ZapLogger
	Services *services.Services
}

func NewParallelStrategy(logger *logger.ZapLogger, services *services.Services) *ParallelStrategy {
	return &ParallelStrategy{
		Logger:   logger,
		Services: services,
	}
}

func (p *ParallelStrategy) Orchestrate(ctx context.Context, endpoints []servicesTypes.ServiceEndpoint) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(endpoints))

	for _, endpoint := range endpoints {
		wg.Add(1)
		go func(endpoint servicesTypes.ServiceEndpoint) {
			defer wg.Done()
			service, err := p.Services.Register(endpoint)
			if err != nil {
				errCh <- err
				p.Logger.Error("Failed to register service", zap.String("name", endpoint.Name), zap.Error(err))
				return
			}

			err = service.Initialize()
			if err != nil {
				errCh <- err
				p.Logger.Error("Failed to initialize service", zap.String("name", service.Name()), zap.Error(err))
				return
			}

			err = service.Start(ctx)
			if err != nil {
				errCh <- err
				p.Logger.Error("Failed to start service", zap.String("name", service.Name()), zap.Error(err))
				return
			}
		}(endpoint)
	}
	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return fmt.Errorf("one or more services failed to orchestrate")
	}
	return nil
}
