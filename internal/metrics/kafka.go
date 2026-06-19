package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	kafkaMessagesPublished = promauto.With(defaultRegistry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "pylon_kafka_messages_published_total",
			Help: "Total number of Kafka messages published.",
		},
		[]string{"topic"},
	)

	kafkaMessagesConsumed = promauto.With(defaultRegistry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "pylon_kafka_messages_consumed_total",
			Help: "Total number of Kafka messages consumed.",
		},
		[]string{"topic", "consumer_group"},
	)

	kafkaPublishDuration = promauto.With(defaultRegistry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pylon_kafka_publish_duration_seconds",
			Help:    "Duration of Kafka publish operations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)
)

func RecordKafkaMessagePublished(topic string) {
	kafkaMessagesPublished.WithLabelValues(normalizeLabel(topic, "unknown")).Inc()
}

func RecordKafkaMessageConsumed(topic, consumerGroup string) {
	kafkaMessagesConsumed.WithLabelValues(
		normalizeLabel(topic, "unknown"),
		normalizeLabel(consumerGroup, "unknown"),
	).Inc()
}

func ObserveKafkaPublish(topic string, duration time.Duration) {
	kafkaPublishDuration.WithLabelValues(normalizeLabel(topic, "unknown")).Observe(duration.Seconds())
}
