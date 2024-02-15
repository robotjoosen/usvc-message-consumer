// Package rabbit is a convenience wrapper for RabbitMQ actions.
package rabbit

import (
	"log/slog"
	"os"

	"github.com/wagslane/go-rabbitmq"
)

type Rabbit struct{}

func NewConnection(url string) *rabbitmq.Conn {
	conn, err := rabbitmq.NewConn(url)
	if err != nil {
		slog.Error(err.Error())

		os.Exit(1)
	}

	return conn
}

func NewConsumer(conn *rabbitmq.Conn, exchange, routingKey, queueName string, handler rabbitmq.Handler) *rabbitmq.Consumer {
	c, err := rabbitmq.NewConsumer(conn, handler, queueName,
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
		rabbitmq.WithConsumerOptionsQueueAutoDelete,
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeName(exchange),
	)
	if err != nil {
		slog.Error(err.Error())

		os.Exit(1)
	}

	return c
}
