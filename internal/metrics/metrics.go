package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	JobsSubmittedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_submitted_total",
		Help: "Total number of jobs submitted to the API.",
	})

	JobsCompletedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_completed_total",
		Help: "Total number of background jobs successfully completed.",
	})

	JobsFailedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_failed_total",
		Help: "Total number of jobs permanently marked failed.",
	})

	JobsRetriedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_retried_total",
		Help: "Total number of job execution retry attempts.",
	})

	JobProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "job_processing_duration_seconds",
		Help:    "Histogram of job execution duration in seconds.",
		Buckets: prometheus.DefBuckets,
	})

	ActiveWorkers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_workers",
		Help: "Number of worker goroutines currently executing jobs.",
	})

	QueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "queue_size",
		Help: "Number of jobs currently buffered in the in-memory queue.",
	})
)

func init() {
	prometheus.MustRegister(
		JobsSubmittedTotal,
		JobsCompletedTotal,
		JobsFailedTotal,
		JobsRetriedTotal,
		JobProcessingDuration,
		ActiveWorkers,
		QueueSize,
	)
}

// Handler returns the Prometheus HTTP metrics handler.
func Handler() http.Handler {
	return promhttp.Handler()
}
