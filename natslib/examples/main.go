package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/yvyrovyi-cinemo/utils/natslib"
)

func main() {
	const NatsConnStr = "nats://localhost:4222"

	ctx := context.Background()

	config := natslib.Config{
		URL:                              NatsConnStr,
		DrainTimeout:                     10 * time.Second,
		SubCheckPeriod:                   10 * time.Second,
		MaxReconnect:                     60,
		ReconnectWait:                    1 * time.Second,
		SubscriptionPendingMessagesLimit: 1000,
	}

	natsClient, err := natslib.Connect(ctx, config, slog.Default())
	if err != nil {
		fmt.Println("ERR: connecting nats:", err)
		os.Exit(1)
	}
	defer natsClient.Close()

	const topicName = "test.topic"

	sub, err := natsClient.Subscribe(ctx, topicName)
	if err != nil {
		fmt.Println("ERR: subscribing:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		buf, err := sub.Receive(ctx)
		if err != nil {
			fmt.Println("ERR: receiving:", err)
			return
		}

		fmt.Printf("message: %s\n", buf)
	}()

	if err := natsClient.Publish(ctx, topicName, []byte("hello")); err != nil {
		fmt.Println("ERR: publishing:", err)
	}

	wg.Wait()
}
