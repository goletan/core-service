package metrics

import (
	"runtime"
	"time"

	observability "github.com/goletan/observability/pkg"
	"github.com/prometheus/client_golang/prometheus"
)

type CoreMetrics struct {
	obs *observability.Observability
}

// Application Metrics: Track application related metrics
var (
	AppErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "runtime",
			Name:      "error_count_total",
			Help:      "Total count of errors encountered by the application.",
		},
		[]string{"type", "service", "context"},
	)
	MemoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "goletan",
		Subsystem: "runtime",
		Name:      "memory_usage_bytes",
		Help:      "Current memory usage in bytes.",
	})
	GoroutinesCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "goletan",
		Subsystem: "runtime",
		Name:      "goroutines_count",
		Help:      "Number of goroutines currently running.",
	})
	scrubber = observability.NewScrubber()
)

func InitMetrics(obs *observability.Observability) *CoreMetrics {
	metrics := &CoreMetrics{obs: obs}
	metrics.Register()

	return metrics
}

func (cbm *CoreMetrics) Register() error {
	if err := prometheus.Register(AppErrorCount); err != nil {
		return err
	}

	if err := prometheus.Register(MemoryUsage); err != nil {
		return err
	}

	if err := prometheus.Register(GoroutinesCount); err != nil {
		return err
	}
	return nil
}

// IncrementErrorCount increments the error counter based on error type, service, and context.
func IncrementErrorCount(errorType, service, context string) {
	scrubbedErrorType := scrubber.Scrub(errorType)
	scrubbedService := scrubber.Scrub(service)
	scrubbedContext := scrubber.Scrub(context)
	AppErrorCount.WithLabelValues(scrubbedErrorType, scrubbedService, scrubbedContext).Inc()
}

// Collect current runtime metrics like memory usage and number of goroutines
func collectRuntimeMetrics(done chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			MemoryUsage.Set(float64(m.Alloc))
			GoroutinesCount.Set(float64(runtime.NumGoroutine()))

			time.Sleep(5 * time.Second)
		}
	}
}
