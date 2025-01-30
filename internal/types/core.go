package types

import (
	events "github.com/goletan/events-service/pkg"
	observability "github.com/goletan/observability-library/pkg"
	resilience "github.com/goletan/resilience-library/pkg"
	services "github.com/goletan/services-library/pkg"
)

type Core struct {
	Config        *CoreConfig
	Observability *observability.Observability
	Resilience    *resilience.DefaultResilienceService
	Services      *services.Services
	EventsClient  *events.EventsClient
}
