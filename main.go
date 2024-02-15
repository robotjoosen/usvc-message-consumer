package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/robotjoosen/usvc-messsage-consumer/pkg/config"
	"github.com/robotjoosen/usvc-messsage-consumer/pkg/rabbit"
	"github.com/wagslane/go-rabbitmq"
)

type (
	Settings struct {
		RabbitMQAddress     string `mapstructure:"MQ_ADDRESS"`
		RabbitMQRoutingKey  string `mapstructure:"MQ_ROUTING_KEY"`
		RabbitMQExchange    string `mapstructure:"MQ_EXCHANGE"`
		RabbitMQQueuePrefix string `mapstructure:"MQ_QUEUE_PREFIX"`
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

func main() {
	s := Settings{}
	if _, err := config.Load(&s, map[string]any{
		"MQ_ADDRESS":      "",
		"MQ_ROUTING_KEY":  "",
		"MQ_EXCHANGE":     "",
		"MQ_QUEUE_PREFIX": "usvc-message-consumer",
	}); err != nil {
		slog.Error(err.Error())

		os.Exit(1)
	}

	rabbit.NewConsumer(
		rabbit.NewConnection(s.RabbitMQAddress),
		s.RabbitMQExchange,
		s.RabbitMQRoutingKey,
		fmt.Sprintf("%s.%s", s.RabbitMQQueuePrefix, uuid.NewString()),
		handleMessage,
	)

	<-make(chan interface{})
}

func handleMessage(d rabbitmq.Delivery) (action rabbitmq.Action) {
	dLog := slog.With(
		slog.String("routing_key", d.RoutingKey),
		slog.String("message_id", d.MessageId),
		slog.String("correlation_id", d.CorrelationId),
	)

	var msg Message

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
}
