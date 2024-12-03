package main

import (
	"context"
	"core/internal/core"

	"github.com/goletan/observability"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obsConfig, err := core.LoadObservabilityConfig()
	if err != nil {
		panic(err)
	}

	obs, err := observability.NewObserver(obsConfig)
	if err != nil {
		panic(err)
	}

}
