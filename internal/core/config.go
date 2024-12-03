package core

import (
	config "github.com/goletan/config/pkg"
	observability "github.com/goletan/observability/shared/config"
)

func LoadObservabilityConfig() (*observability.ObservabilityConfig, error) {
	var cfg observability.ObservabilityConfig
	if err := config.LoadConfig("Observability", &cfg, nil); err != nil {
		return nil, err
	}
	return &cfg, nil
}
