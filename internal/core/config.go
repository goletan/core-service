package core

import (
	"core/internal/types"
	"github.com/goletan/config/pkg"
	logger "github.com/goletan/logger/pkg"
	"go.uber.org/zap"
)

// LoadCoreConfig loads the core configuration and returns it as a pointer to CoreConfig.
// It returns an error if the configuration cannot be loaded.
func LoadCoreConfig(log *logger.ZapLogger) (*types.CoreConfig, error) {
	var cfg types.CoreConfig
	if err := config.LoadConfig("Core", &cfg, log); err != nil {
		log.Error("Failed to load core configuration", zap.Error(err))
		return nil, err
	}

	return &cfg, nil
}
