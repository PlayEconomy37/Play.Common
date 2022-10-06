package events

import (
	"fmt"

	"github.com/PlayEconomy37/Play.Common/configuration"
	amqp "github.com/rabbitmq/amqp091-go"
)

// NewRabbitMQConnection initializes new RabbitMQ connection
func NewRabbitMQConnection(cfg *configuration.Config) (*amqp.Connection, error) {
	connAddr := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	return amqp.Dial(connAddr)
}

// NewAzureServiceBusConnection initializes new Azure Service Bus connection
// func NewAzureServiceBusConnection(connectionString string) (*azservicebus.Client, error) {
// 	return azservicebus.NewClientFromConnectionString(connectionString, nil)
// }
