package core

import (
	"core/internal/types"
	"github.com/goletan/config/pkg"
	obsCfg "github.com/goletan/observability/shared/config"
	"github.com/goletan/observability/shared/logger"
	"go.uber.org/zap"
)

// LoadCoreConfig loads the core configuration and returns it as a pointer to CoreConfig.
// It returns an error if the configuration cannot be loaded.
func LoadCoreConfig(log *logger.ZapLogger) (*types.CoreConfig, error) {
	var cfg types.CoreConfig
	if err := config.LoadConfig("Core", &cfg, nil); err != nil {
		log.Error("Failed to load core configuration", zap.Error(err))
		return nil, err
	}

	return &cfg, nil
}

// LoadObservabilityConfig loads the observability configuration settings from a predefined source.
// It returns the loaded ObservabilityConfig struct and an error if the configuration loading fails.
func LoadObservabilityConfig(log *logger.ZapLogger) (*obsCfg.ObservabilityConfig, error) {
	var cfg obsCfg.ObservabilityConfig
	if err := config.LoadConfig("Observability", &cfg, log); err != nil {
		log.Error("Failed to load observability configuration", zap.Error(err))
		return nil, err
	}

	return &cfg, nil
}
