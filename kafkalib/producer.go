package kafkalib

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/IBM/sarama"
)

type (
	Producer struct {
		producer sarama.AsyncProducer
		logger   *slog.Logger
	}

	ProducerConfig struct {
		Config `yaml:",inline"`
	}
)

func NewProducer(config ProducerConfig, logger *slog.Logger) (*Producer, error) {
	producerConfig := sarama.NewConfig()
	producerConfig.ClientID = config.ClientID
	producerConfig.Metadata.AllowAutoTopicCreation = false

	// we MUST read producer.Errors() chan !!
	producerConfig.Producer.Return.Errors = true

	if config.DebugLog {
		sarama.Logger = NewSaramaLogger(logger.With(LogsLabelComponent, "sarama"))
	}

	producer, err := sarama.NewAsyncProducer(strings.Split(config.Brokers, ","), producerConfig)
	if err != nil {
		return nil, fmt.Errorf("creating producer: %w", err)
	}

	logger = logger.With(LogsLabelComponent, "kafkalib-producer")

	go func() {
		for err := range producer.Errors() {
			logger.Error("kafka producer error", "error", err)
		}
	}()

	return &Producer{
		producer: producer,
		logger:   logger,
	}, nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

func (p *Producer) Produce(msg *Message) error {
	p.producer.Input() <- &sarama.ProducerMessage{
		Topic:    msg.Topic,
		Key:      sarama.ByteEncoder(msg.Key),
		Value:    sarama.ByteEncoder(msg.Payload),
		Metadata: nil,
	}

	return nil
}
