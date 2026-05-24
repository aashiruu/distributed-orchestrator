package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	JobsIngestedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orchestrator_jobs_ingested_total",
			Help: "The total number of jobs submitted via the API Ingestion gateway",
		},
		[]string{"job_name"},
	)

	JobExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "worker_job_execution_duration_seconds",
			Help:    "Time spent executing jobs inside the worker nodes pool",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"job_name", "status"},
	)

	LockFailuresCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_redis_lock_failures_total",
			Help: "The total number of processing loops rejected due to active execution locks",
		},
		[]string{"job_name"},
	)
)
