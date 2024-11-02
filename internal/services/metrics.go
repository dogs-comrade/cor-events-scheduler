package services

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SchedulerMetrics struct {
	scheduleCreations  prometheus.Counter
	scheduleUpdates    prometheus.Counter
	scheduleDeletions  prometheus.Counter
	scheduleRiskScores prometheus.Histogram
	techBreakDurations prometheus.Histogram
}

func NewSchedulerMetrics() *SchedulerMetrics {
	metrics := &SchedulerMetrics{
		scheduleCreations: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "schedule_creations_total",
			Help: "Total number of schedules created",
		}),
		scheduleUpdates: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "schedule_updates_total",
			Help: "Total number of schedule updates",
		}),
		scheduleDeletions: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "schedule_deletions_total",
			Help: "Total number of schedule deletions",
		}),
		scheduleRiskScores: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "schedule_risk_scores",
			Help:    "Distribution of schedule risk scores",
			Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9},
		}),
		techBreakDurations: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "tech_break_durations_minutes",
			Help:    "Distribution of technical break durations",
			Buckets: []float64{5, 10, 15, 20, 30, 45, 60, 90, 120},
		}),
	}

	prometheus.MustRegister(
		metrics.scheduleCreations,
		metrics.scheduleUpdates,
		metrics.scheduleDeletions,
		metrics.scheduleRiskScores,
		metrics.techBreakDurations,
	)

	return metrics
}
