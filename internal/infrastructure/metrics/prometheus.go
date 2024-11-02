package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ScheduleOpsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "schedule_operations_total",
			Help: "The total number of schedule operations",
		},
		[]string{"operation"},
	)

	ScheduleOpsDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "schedule_operation_duration_seconds",
			Help:    "Duration of schedule operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	ActiveSchedules = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_schedules",
			Help: "The current number of active schedules",
		},
	)
)
