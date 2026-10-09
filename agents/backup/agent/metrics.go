package agent

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricJobs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "backup_jobs_total",
		Help: "Backup jobs by source type and status.",
	}, []string{"source_type", "status"})

	metricDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "backup_job_duration_seconds",
		Help:    "Backup job duration.",
		Buckets: prometheus.DefBuckets,
	}, []string{"source_type"})

	metricBytesRead = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "backup_bytes_read_total",
		Help: "Bytes read from sources.",
	}, []string{"source_type"})

	metricBytesWritten = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "backup_bytes_written_total",
		Help: "Bytes written to MinIO.",
	}, []string{"source_type"})

	metricLastSuccess = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "backup_last_success_timestamp",
		Help: "Unix time of last successful backup per source.",
	}, []string{"source_handle"})

	metricInProgress = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "backup_job_in_progress",
		Help: "1 if a backup is running for the source.",
	}, []string{"source_handle"})

	metricRestores = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "backup_restore_total",
		Help: "Restore operations.",
	}, []string{"status"})

	metricPruneDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "backup_prune_objects_deleted",
		Help: "Objects deleted by retention prune.",
	})
)
