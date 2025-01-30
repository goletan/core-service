package strategies

import (
	"context"
	"github.com/coreos/go-semver/semver"
	logger "github.com/goletan/logger-library/pkg"
	services "github.com/goletan/services-library/pkg"
	"github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
	"strconv"
)

type HybridStrategy struct {
	Logger        *logger.ZapLogger
	Services      *services.Services
	PriorityQueue *PriorityQueueManager
	cache         map[string]int // Cached priority map to avoid re-computation
}

// NewHybridStrategy initializes the hybrid strategy but defers priority calculations to runtime.
func NewHybridStrategy(logger *logger.ZapLogger, services *services.Services) *HybridStrategy {
	return &HybridStrategy{
		Logger:        logger,
		Services:      services,
		PriorityQueue: NewPriorityQueueManager(nil), // Defer priority initialization
		cache:         make(map[string]int),
	}
}

// Orchestrate manages endpoints using their priority derived dynamically.
func (h *HybridStrategy) Orchestrate(ctx context.Context, endpoints []types.ServiceEndpoint) error {
	if len(endpoints) == 0 {
		h.Logger.Warn("No endpoints provided for orchestration")
		return nil
	}

	h.buildPriorityMap(endpoints)

	for _, endpoint := range endpoints {
		h.PriorityQueue.Push(endpoint)
	}

	for h.PriorityQueue.Len() > 0 {
		item := h.PriorityQueue.Pop()
		logServiceOrchestration(
			h.Logger, "hybrid",
			item.Endpoint, item.Priority,
		)

		service, err := h.Services.Register(item.Endpoint)
		if err != nil {
			return err
		}
		h.Logger.Info("Service found in services-library", zap.String("name", service.Name()))

		err = service.Initialize()
		if err != nil {
			return err
		}
		h.Logger.Info("Service initialized", zap.String("name", service.Name()))

		err = service.Start(ctx)
		if err != nil {
			return err
		}
		h.Logger.Info("Service started", zap.String("name", service.Name()))
	}
	return nil
}

// buildPriorityMap builds the priority map from the given endpoints and caches it.
func (h *HybridStrategy) buildPriorityMap(endpoints []types.ServiceEndpoint) {
	for _, endpoint := range endpoints {
		priority := h.extractPriorityFromEndpoint(endpoint)
		h.cache[endpoint.Name] = priority
	}

	h.PriorityQueue = NewPriorityQueueManager(h.cache)
}

// extractPriorityFromEndpoint derives a priority from the endpoint metadata.
func (h *HybridStrategy) extractPriorityFromEndpoint(endpoint types.ServiceEndpoint) int {
	if tagPriorityStr, exists := endpoint.Tags["priority"]; exists {
		if priority, err := strconv.Atoi(tagPriorityStr); err == nil {
			return priority // Valid priority from tag
		}
	}

	if endpoint.Version != "" {
		if priority := parseVersionToPriority(endpoint.Version); priority > 0 {
			return priority
		}
	}

	return DefaultServicePriority
}

// parseVersionToPriority maps version strings to priority levels.
func parseVersionToPriority(version string) int {
	parsedVersion := semver.New(version)
	if parsedVersion == nil {
		return DefaultServicePriority
	}

	return int(parsedVersion.Major*100 + parsedVersion.Minor*10 + parsedVersion.Patch)
}
