package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/google/uuid"
	"github.com/grafana/pyroscope-go"
	_ "github.com/grafana/pyroscope-go/godeltaprof/http/pprof"
	"github.com/hellofresh/health-go/v5"
	"github.com/robotjoosen/usvc-message-consumer/internal/server"
	"github.com/robotjoosen/usvc-message-consumer/pkg/config"
	"github.com/robotjoosen/usvc-message-consumer/pkg/rabbit"
	"github.com/wagslane/go-rabbitmq"
)

var (
	serviceName  = "message-consumer"
	buildName    = "usvc-message-consumer"
	buildVersion = "dev"
	buildCommit  = "n/a"
)

type (
	Settings struct {
		RabbitMQAddress     string `mapstructure:"MQ_ADDRESS"`
		RabbitMQRoutingKey  string `mapstructure:"MQ_ROUTING_KEY"`
		RabbitMQExchange    string `mapstructure:"MQ_EXCHANGE"`
		RabbitMQQueuePrefix string `mapstructure:"MQ_QUEUE_PREFIX"`
		ServerIP            string `mapstructure:"SERVER_IP"`
		ServerPort          int    `mapstructure:"SERVER_PORT"`
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
		"SERVER_IP":       "0.0.0.0",
		"SERVER_PORT":     8080,
		"OTEL_ADDRESS":    "http://pyroscope:4040",
	}); err != nil {
		slog.Error(err.Error())

		os.Exit(100)
	}

	slog.SetDefault(slog.With(
		slog.String("build_name", buildName),
		slog.String("build_version", buildVersion),
	))

	slog.Info("service started",
		slog.String("build_commit", buildCommit),
		slog.Any("settings", s),
	)

	startObservability(s)
	startMessageConsumer(s)
	startHealthz(s)

	<-make(chan interface{})
}

func startHealthz(s Settings) {
	h, err := health.New(
		health.WithSystemInfo(),
		health.WithComponent(health.Component{
			Name:    buildName + " - " + buildCommit,
			Version: buildVersion,
		}),
	)
	if err != nil {
		slog.Error("failed to initialise health handler",
			slog.String("error", err.Error()),
		)

		os.Exit(200)
	}

	server.New(s.ServerPort, map[string]http.HandlerFunc{
		"/healthz": h.HandlerFunc,
	}).Run()

	slog.Info("server started")
}

func startObservability(s Settings) {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	if _, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: serviceName,
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

		os.Exit(300)
	}

	slog.Info("observability started")
}

func startMessageConsumer(s Settings) {
	conn, err := rabbit.NewConnection(s.RabbitMQAddress)
	if err != nil {
		os.Exit(400)
	}

	go func() {
		if err = rabbit.RunConsumer(
			conn,
			s.RabbitMQExchange,
			s.RabbitMQRoutingKey,
			fmt.Sprintf("%s.%s", s.RabbitMQQueuePrefix, uuid.NewString()),
			handleMessage,
			context.Background(),
		); err != nil {
			os.Exit(401)
		}
	}()

	slog.Info("message consumer started")
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
