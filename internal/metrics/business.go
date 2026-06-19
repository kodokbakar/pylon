package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	activeConnections = promauto.With(defaultRegistry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pylon_websocket_connections_active",
			Help: "Number of active WebSocket connections.",
		},
		[]string{"room_type"},
	)

	messagesSent = promauto.With(defaultRegistry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "pylon_messages_sent_total",
			Help: "Total number of messages sent.",
		},
		[]string{"room_type", "message_type"},
	)

	roomsCreated = promauto.With(defaultRegistry).NewCounter(
		prometheus.CounterOpts{
			Name: "pylon_rooms_created_total",
			Help: "Total number of rooms created.",
		},
	)

	usersOnline = promauto.With(defaultRegistry).NewGauge(
		prometheus.GaugeOpts{
			Name: "pylon_users_online",
			Help: "Number of currently online users.",
		},
	)
)

func IncActiveConnections(roomType string) {
	activeConnections.WithLabelValues(normalizeLabel(roomType, "unknown")).Inc()
}

func DecActiveConnections(roomType string) {
	activeConnections.WithLabelValues(normalizeLabel(roomType, "unknown")).Dec()
}

func RecordMessageSent(roomType, messageType string) {
	messagesSent.WithLabelValues(
		normalizeLabel(roomType, "unknown"),
		normalizeLabel(messageType, "unknown"),
	).Inc()
}

func RecordRoomCreated() {
	roomsCreated.Inc()
}

func IncUsersOnline() {
	usersOnline.Inc()
}

func DecUsersOnline() {
	usersOnline.Dec()
}

func SetUsersOnline(value float64) {
	usersOnline.Set(value)
}
