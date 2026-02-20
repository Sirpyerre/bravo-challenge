// Package metrics define las métricas Prometheus personalizadas de la aplicación.
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// EventsProcessed cuenta eventos procesados por los workers.
	EventsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "credit_events_processed_total",
			Help: "Total de eventos procesados por los workers.",
		},
		[]string{"worker", "status"},
	)

	// EventProcessingDuration mide la duración de procesamiento de eventos.
	EventProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "credit_event_processing_duration_seconds",
			Help:    "Duración del procesamiento de eventos por worker.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"worker", "status"},
	)

	// EventsErrors cuenta errores en el procesamiento de eventos.
	EventsErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "credit_events_errors_total",
			Help: "Total de errores en el procesamiento de eventos.",
		},
		[]string{"worker", "reason"},
	)

	// ApplicationsCreated cuenta solicitudes de crédito creadas.
	ApplicationsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "credit_applications_created_total",
			Help: "Total de solicitudes de crédito creadas.",
		},
		[]string{"country"},
	)

	// IdempotencyHits cuenta hits y misses del servicio de idempotencia.
	IdempotencyHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "credit_idempotency_hits_total",
			Help: "Total de consultas al servicio de idempotencia (hit/miss).",
		},
		[]string{"result"},
	)
)

func init() {
	prometheus.MustRegister(
		EventsProcessed,
		EventProcessingDuration,
		EventsErrors,
		ApplicationsCreated,
		IdempotencyHits,
	)
}
