package core

import (
	config "github.com/goletan/config-library/pkg"
	"github.com/goletan/core-service/internal/types"
	logger "github.com/goletan/logger-library/pkg"
	"go.uber.org/zap"
)

// LoadCoreConfig loads the core-service configuration and returns it as a pointer to CoreConfig.
// It returns an error if the configuration cannot be loaded.
func LoadCoreConfig(log *logger.ZapLogger) (*types.CoreConfig, error) {
	var cfg types.CoreConfig
	if err := config.LoadConfig("Core", &cfg, log); err != nil {
		log.Error("Failed to load core-service configuration", zap.Error(err))
		return nil, err
	}

	return &cfg, nil
}
