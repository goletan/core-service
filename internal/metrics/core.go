package metrics

import (
	"go.uber.org/zap"
	"runtime"
	"time"

	"github.com/goletan/observability/internal/logger"
	"github.com/goletan/security/shared/scrubber"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct{}

var (

	// AppErrorCount is a Prometheus CounterVec that tracks the total count of errors encountered by the application.
	// It uses labels—type, service, and context—to differentiate error occurrences.
	AppErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "runtime",
			Name:      "error_count_total",
			Help:      "Total count of errors encountered by the application.",
		},
		[]string{"type", "service", "context"},
	)

	// MemoryUsage is a Prometheus Gauge that tracks the current memory usage in bytes.
	MemoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "goletan",
		Subsystem: "runtime",
		Name:      "memory_usage_bytes",
		Help:      "Current memory usage in bytes.",
	})

	// GoroutinesCount is a Prometheus Gauge that tracks the number of goroutines currently running.
	GoroutinesCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "goletan",
		Subsystem: "runtime",
		Name:      "goroutines_count",
		Help:      "Number of goroutines currently running.",
	})

	// scrub is an instance of Scrubber initialized with default patterns for sanitizing sensitive information.
	scrub = scrubber.NewScrubber()
)

func InitMetrics(log *logger.ZapLogger) *Metrics {
	metrics := &Metrics{}
	err := metrics.Register()

	if err != nil {
		log.Error("Cannot register metrics", zap.Error(err))
	}

	return metrics
}

func (cbm *Metrics) Register() error {
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
	scrubbedErrorType := scrub.Scrub(errorType)
	scrubbedService := scrub.Scrub(service)
	scrubbedContext := scrub.Scrub(context)
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
