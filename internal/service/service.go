package service

import (
	"context"
	"github.com/goletan/core-service/internal/core"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
)

type CoreService struct {
	ServiceName    string
	ServiceAddress string
	ServiceType    string
	Ports          []servicesTypes.ServicePort
	Tags           map[string]string
	core           *core.Core
}

func NewCoreService(endpoint servicesTypes.ServiceEndpoint) (servicesTypes.Service, error) {
	newCore, err := core.NewCore()
	if err != nil || newCore == nil {
		return nil, err
	}

	return &CoreService{
		ServiceName:    endpoint.Name,
		ServiceAddress: endpoint.Address,
		ServiceType:    "core",
		Ports:          endpoint.Ports,
		Tags:           endpoint.Tags,
		core:           newCore,
	}, nil
}

func (c *CoreService) Name() string {
	return c.ServiceName
}

func (c *CoreService) Address() string {
	return c.ServiceAddress
}

func (c *CoreService) Type() string {
	return c.ServiceType
}

func (c *CoreService) Metadata() map[string]string {
	return c.Tags
}

func (c *CoreService) Initialize() error {
	if c.core == nil {
		newCore, err := core.NewCore()
		if err != nil || newCore == nil {
			return err
		}
		c.core = newCore
	}
	c.core.Observability.Logger.Info("Core service initialized")

	return nil
}

func (c *CoreService) Start(ctx context.Context) error {
	c.core.Observability.Logger.Info("Starting core service...")
	err := c.core.Start(ctx)
	if err != nil {
		c.core.Observability.Logger.Fatal("Failed to start core service", zap.Error(err))
		return err
	}
	c.core.Observability.Logger.Info("Core service is running...")

	return nil
}

func (c *CoreService) Stop(ctx context.Context) error {
	c.core.Observability.Logger.Info("Core service shutting down...")
	err := c.core.Shutdown(ctx)
	if err != nil {
		c.core.Observability.Logger.Error("Failed to shutdown core-service", zap.Error(err))
		return err
	}

	c.core.Observability.Logger.Info("Core service shut down successfully")
	return nil
}
