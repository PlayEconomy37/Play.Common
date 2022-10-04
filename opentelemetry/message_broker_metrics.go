package opentelemetry

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MessageBrokerMetrics is a struct that holds some prometheus metrics
// regarding our message broker
type MessageBrokerMetrics struct {
	IncomingMessagesCounter *prometheus.CounterVec
	SuccessMessagesCounter  *prometheus.CounterVec
	ErrorMessagesCounter    *prometheus.CounterVec
}

// CreateMessageBrokerMetrics creates counters used to keep
// track of message broker metrics in our application
func CreateMessageBrokerMetrics(appName string) *MessageBrokerMetrics {
	incomingMessagesCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_incoming_messages_total", appName),
		Help: "The total number of incoming messages",
	}, []string{"queue"})

	successMessagesCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_success_incoming_messages_total", appName),
		Help: "The total number of success incoming success messages",
	}, []string{"queue"})

	errorMessagesCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_error_incoming_message_total", appName),
		Help: "The total number of error incoming success messages",
	}, []string{"queue"})

	return &MessageBrokerMetrics{
		IncomingMessagesCounter: incomingMessagesCounter,
		SuccessMessagesCounter:  successMessagesCounter,
		ErrorMessagesCounter:    errorMessagesCounter,
	}
}
