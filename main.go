package main

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/pyroscope-go"
	"log/slog"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/google/uuid"
	_ "github.com/grafana/pyroscope-go/godeltaprof/http/pprof"
	"github.com/robotjoosen/usvc-message-consumer/pkg/config"
	"github.com/robotjoosen/usvc-message-consumer/pkg/rabbit"
	"github.com/wagslane/go-rabbitmq"
)

var (
	buildName    = "api-message-generator"
	buildVersion = "dev"
	buildCommit  = "n/a"
)

type (
	Settings struct {
		RabbitMQAddress     string `mapstructure:"MQ_ADDRESS"`
		RabbitMQRoutingKey  string `mapstructure:"MQ_ROUTING_KEY"`
		RabbitMQExchange    string `mapstructure:"MQ_EXCHANGE"`
		RabbitMQQueuePrefix string `mapstructure:"MQ_QUEUE_PREFIX"`
		OTELAddress         string `mapstructure:"OTEL_ADDRESS"`
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
		"OTEL_ADDRESS":    "http://pyroscope:4040",
	}); err != nil {
		slog.Error(err.Error())

		os.Exit(1)
	}

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	if _, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "message-consumer", // todo: figure out what the nicest way is to set an application name
		ServerAddress:   s.OTELAddress,
		Logger:          nil,
		Tags: map[string]string{
			"name":    buildName,
			"version": buildVersion,
			"commit":  buildCommit,
		},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	}); err != nil {
		slog.Error("failed to start pyroscope",
			slog.String("error", err.Error()),
		)

		os.Exit(2)
	}

	slog.Info("service started",
		slog.String("build_name", buildName),
		slog.String("build_version", buildVersion),
		slog.String("build_commit", buildCommit),
		slog.Any("settings", s),
	)

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
		slog.String("message_id", d.MessageId),
		slog.String("correlation_id", d.CorrelationId),
		slog.String("routing_key", d.RoutingKey),
	)

	var msg Message

	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		dLog.Error("failed to unmarshal message")

		return rabbitmq.NackDiscard
	}

	dLog.Info("message received",
		slog.String("action_type", msg.ActionType),
		slog.Any("record", msg.Data),
	)

	return rabbitmq.Ack
}
