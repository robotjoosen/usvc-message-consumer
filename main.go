package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"os"

	"github.com/robotjoosen/usvc-messsage-consumer/pkg/config"
	"github.com/robotjoosen/usvc-messsage-consumer/pkg/rabbit"
	"github.com/wagslane/go-rabbitmq"
)

const (
	rmqRoutingKey = "default"
	rmqExchange   = "traceability"
)

type (
	Settings struct {
		RabbitMQAddress string `mapstructure:"MQ_ADDRESS"`
	}

	Record struct {
		ID string `json:"ID"`
	}

	Message struct {
		CorrelationID string `json:"correlation_id"`
		ActionType    string `json:"action_type"`
		Data          Record `json:"data"`
	}
)

func initialize() Settings {
	s := Settings{}
	if _, err := config.Load(&s, map[string]any{
		"MQ_ADDRESS": "",
	}); err != nil {
		os.Exit(1)
	}

	return s
}

func main() {
	s := initialize()

	conn := rabbit.NewConnection(s.RabbitMQAddress)

	rabbit.NewConsumer(
		conn,
		rmqExchange,
		"usvc-message-consumer."+uuid.NewString(),
		rmqRoutingKey,
		func(d rabbitmq.Delivery) (action rabbitmq.Action) {
			var msg Message
			dLog := slog.With(
				slog.String("routing_key", d.RoutingKey),
				slog.String("message_id", d.MessageId),
				slog.String("correlation_id", d.CorrelationId),
			)

			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				dLog.Error("failed to unmarshal message")

				return rabbitmq.NackDiscard
			}

			dLog.Info("received message",
				slog.String("action_type", msg.ActionType),
				slog.Any("record", msg.Data),
			)

			return rabbitmq.Ack
		},
	)

	<-make(chan interface{})
}
