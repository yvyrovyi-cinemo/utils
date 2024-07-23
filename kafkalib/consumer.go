package kafkalib

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	circlebuff "git.internal.cinemo.com/~yvyrovyi/utils/circlebuf"

	"github.com/IBM/sarama"
)

type Consumer struct {
	ConsumerConfig
	consumerGroup              sarama.ConsumerGroup
	topics                     []string
	handler                    MessageHandler
	logger                     *slog.Logger
	metrics                    *kafkaMetrics
	onAssignPartitionHandler   RebalanceHandler
	onUnassignPartitionHandler RebalanceHandler
}

type ConsumerConfig struct {
	Config                  `yaml:",inline"`
	GroupID                 string        `env:"GROUP_ID" yaml:"group_id"`
	Topics                  []string      `env:"TOPICS" envSeparator:"," yaml:"topics"`
	FaultTolerancePeriod    time.Duration `env:"FAULT_TOLERANCE_PERIOD" yaml:"fault_tolerance_period"`
	FaultToleranceThreshold int           `env:"FAULT_TOLERANCE_THRESHOLD" yaml:"fault_tolerance_threshold"`
}

func (c ConsumerConfig) WithDefaults() ConsumerConfig {
	if c.FaultTolerancePeriod == 0 {
		c.FaultTolerancePeriod = 30 * time.Second
	}

	if c.FaultToleranceThreshold == 0 {
		c.FaultToleranceThreshold = 3
	}

	return c
}

type MessageHandler func(context.Context, *Message) error
type RebalanceHandler func(ctx context.Context, topic string) error

func NewConsumer(config ConsumerConfig, logger *slog.Logger) (*Consumer, error) {
	consumerConfig := sarama.NewConfig()
	if config.DebugLog {
		sarama.Logger = NewSaramaLogger(logger.With(LogsLabelComponent, "sarama"))
	}

	logger = logger.With(LogsLabelComponent, "kafkalib-consumer")

	consumerConfig.ClientID = config.ClientID
	consumerConfig.Consumer.Group.Rebalance.GroupStrategies =
		[]sarama.BalanceStrategy{
			sarama.NewBalanceStrategyRoundRobin(),
		}

	cg, err := sarama.NewConsumerGroup(strings.Split(config.Brokers, ","), config.GroupID, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("creating consumer group: %w", err)
	}

	metrics, err := initMetrics()
	if err != nil {
		logger.Error("failed to init kafkaMetrics", "error", err)
	}

	return &Consumer{
		ConsumerConfig: config.WithDefaults(),
		consumerGroup:  cg,
		topics:         config.Topics,
		logger:         logger,
		metrics:        metrics,
	}, nil
}

func (c *Consumer) SetOnAssignHandler(h RebalanceHandler) {
	c.onAssignPartitionHandler = h
}

func (c *Consumer) SetOnUnassignHandler(h RebalanceHandler) {
	c.onUnassignPartitionHandler = h
}

func (c *Consumer) Run(ctx context.Context, handler MessageHandler) error {
	chanErr := make(chan error)
	c.handler = handler

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			close(chanErr)
			wg.Done()
		}()

		for {
			c.logger.Info("connecting to broker",
				"client_id", c.ClientID,
				"group_id", c.GroupID,
				"topics", c.Topics,
			)

			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.consumerGroup.Consume(ctx, c.topics, c); err != nil {
				// check if context was cancelled, signaling that the consumer should stop
				if ctx.Err() != nil {
					c.logger.Info("consumer exited Consume on ctx.Done")
					return
				}

				chanErr <- err
			}
		}
	}()

	c.logger.Info("kafka consumer is up and running")

	var resErr error

	circleBuff := circlebuff.New[int](c.FaultToleranceThreshold, c.FaultTolerancePeriod)

waitLoop:
	for {
		select {
		case <-ctx.Done():
			break waitLoop

		case err := <-chanErr:
			c.logger.Error("consumer error", "error", err)
			circleBuff.AddValue(1)

			faultsNum, err := circleBuff.Sum()
			if errors.Is(err, circlebuff.ErrNoValues) {
				continue
			}

			if faultsNum >= c.FaultToleranceThreshold {
				c.logger.Error("kafka consumer fault threshold exceeded")
				resErr = err
				break waitLoop
			}
		}
	}

	cancel()
	wg.Wait()

	if err := c.consumerGroup.Close(); err != nil {
		return fmt.Errorf("closing Kafka client: %w", err)
	}

	c.logger.Info("sarama consumer closed")

	return resErr
}

// Implementing ConsumerGroupHandler.
// Copied from example:
// https://github.com/Shopify/sarama/blob/main/examples/consumergroup/main.go

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	ctx := session.Context()

	if c.onAssignPartitionHandler != nil {
		for topic := range session.Claims() {
			if err := c.onAssignPartitionHandler(ctx, topic); err != nil {
				return fmt.Errorf("on assign partition: %w", err)
			}
		}
	}

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	ctx := session.Context()

	if c.onUnassignPartitionHandler != nil {
		for topic := range session.Claims() {
			if err := c.onUnassignPartitionHandler(ctx, topic); err != nil {
				return fmt.Errorf("on unassign partition: %w", err)
			}
		}
	}

	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	c.logger.Debug("ConsumeClaim",
		"topic", claim.Topic(),
		"partition", claim.Partition(),
	)

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message := <-claim.Messages():
			if err := c.messageHandler(session.Context(), message); err != nil {
				return err
			}

			session.MarkMessage(message, "")

			c.metrics.setConsumerLag(
				claim.Topic(), claim.Partition(), c.GroupID,
				claim.HighWaterMarkOffset()-message.Offset,
			)

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			c.logger.Debug("consume got ctx.Done",
				"topic", claim.Topic(),
				"partition_id", claim.Partition(),
			)
			return nil
		}
	}
}

func (c *Consumer) messageHandler(ctx context.Context, msg *sarama.ConsumerMessage) error {
	if msg == nil {
		return nil
	}

	svcMessage := Message{
		Topic:     msg.Topic,
		Key:       msg.Key,
		Payload:   msg.Value,
		Partition: &msg.Partition,
	}

	return c.handler(ctx, &svcMessage)
}
