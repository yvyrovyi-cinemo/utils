package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yvyrovyi-cinemo/utils/kafkalib"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("ERR:", err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	config := kafkalib.ConsumerConfig{
		Config: kafkalib.Config{
			ClientID:   "example-consumer",
			Brokers:    "localhost:9095",
			TLSEnabled: false,
		},
		GroupID: "example-group",
		Topics:  []string{"test_topic"},
	}

	client, err := kafkalib.NewConsumer(config, logger)
	if err != nil {
		return fmt.Errorf("creating kafka consumer: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context, message *kafkalib.Message) error {
		prt := int32(-1)

		if message.Partition != nil {
			prt = *message.Partition
		}
		fmt.Printf(`{"topic:": "%s", "key": "%s", "partition": %d, "payload": %s}`+"\n",
			message.Topic,
			message.Key,
			prt,
			message.Payload,
		)

		return nil
	})

}
