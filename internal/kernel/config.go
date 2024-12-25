package kernel

import (
	config "github.com/goletan/config-library/pkg"
	"github.com/goletan/kernel-service/internal/types"
	logger "github.com/goletan/logger-library/pkg"
	"go.uber.org/zap"
)

// LoadCoreConfig loads the kernel-service configuration and returns it as a pointer to CoreConfig.
// It returns an error if the configuration cannot be loaded.
func LoadCoreConfig(log *logger.ZapLogger) (*types.KernelConfig, error) {
	var cfg types.KernelConfig
	if err := config.LoadConfig("Kernel", &cfg, log); err != nil {
		log.Error("Failed to load kernel-service configuration", zap.Error(err))
		return nil, err
	}

	return &cfg, nil
}
