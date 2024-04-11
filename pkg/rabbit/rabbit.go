// Package rabbit is a convenience wrapper for RabbitMQ actions.
package rabbit

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/wagslane/go-rabbitmq"
)

func NewConnection(url string) (*rabbitmq.Conn, error) {
	conn, err := rabbitmq.NewConn(url)
	if err != nil {
		slog.Error("failed to connect to rabbitMQ",
			slog.String("error", err.Error()),
		)

		return nil, err
	}

	return conn, nil
}

func NewConsumer(conn *rabbitmq.Conn, exchange, routingKey, queueName string) (*rabbitmq.Consumer, error) {
	c, err := rabbitmq.NewConsumer(conn, queueName,
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
		rabbitmq.WithConsumerOptionsQueueAutoDelete,
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeName(exchange),
	)
	if err != nil {
		slog.Error("failed to create consumer",
			slog.String("error", err.Error()),
			slog.String("queue_name", queueName),
			slog.String("exchange", exchange),
			slog.String("routing_key", routingKey),
		)

		return nil, err
	}

	slog.Info("consumer created",
		slog.String("queue_name", queueName),
		slog.String("exchange", exchange),
		slog.String("routing_key", routingKey),
	)

	return c, nil
}

func RunConsumer(conn *rabbitmq.Conn, exchange, routingKey, queueName string, handler rabbitmq.Handler, ctx context.Context) error {
	c, err := NewConsumer(conn, exchange, routingKey, queueName)
	if err != nil {
		return err
	}

	defer func(c *rabbitmq.Consumer) {
		slog.Info("consumer closed",
			slog.String("queue_name", queueName),
			slog.String("exchange", exchange),
			slog.String("routing_key", routingKey),
		)

		c.Close()
	}(c)

	if err = c.Run(handler); err != nil {
		slog.Error("failed to run consumer",
			slog.String("error", err.Error()),
			slog.String("queue_name", queueName),
			slog.String("exchange", exchange),
			slog.String("routing_key", routingKey),
		)

		return err
	}

	<-ctx.Done()

	slog.Info("consumer shutting down",
		slog.String("queue_name", queueName),
		slog.String("exchange", exchange),
		slog.String("routing_key", routingKey),
	)

	return nil
}

func NewPublisher(conn *rabbitmq.Conn, exchange string) (*rabbitmq.Publisher, error) {
	p, err := rabbitmq.NewPublisher(conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeName(exchange),
	)
	if err != nil {
		slog.Error("failed to create publisher",
			slog.String("error", err.Error()),
			slog.String("exchange", exchange),
		)

		return nil, err
	}

	slog.Info("publisher created",
		slog.String("exchange", exchange),
	)

	return p, nil
}

func Publish(msg []byte, routingKey, exchange, correlationID string, pub *rabbitmq.Publisher) error {
	messageID := uuid.NewString()

	err := pub.Publish(msg, []string{routingKey},
		rabbitmq.WithPublishOptionsCorrelationID(correlationID),
		rabbitmq.WithPublishOptionsMessageID(messageID),
		rabbitmq.WithPublishOptionsExchange(exchange),
		rabbitmq.WithPublishOptionsContentType("application/json"),
	)
	if err != nil {
		slog.Error(err.Error())

		return err
	}

	slog.Info("message published",
		slog.String("message_id", messageID),
		slog.String("correlation_id", correlationID),
		slog.String("exchange", exchange),
		slog.String("routing_key", routingKey),
	)

	return nil
}
