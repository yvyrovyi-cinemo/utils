package kafkalib

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type kafkaMetrics struct {
	consumerLagGaugeVec *prometheus.GaugeVec
}

func initMetrics() (*kafkaMetrics, error) {
	consumerLagGaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_partition_lag",
			Help: "a lag of partition consumer",
			ConstLabels: map[string]string{
				"service_name": "sarama",
			},
		},
		[]string{"topic", "partition", "consumer_group"},
	)

	if err := prometheus.Register(consumerLagGaugeVec); err != nil {
		return nil, fmt.Errorf("failed to register consumer lag kafkaMetrics: %w", err)
	}

	return &kafkaMetrics{
		consumerLagGaugeVec: consumerLagGaugeVec,
	}, nil
}

func (m *kafkaMetrics) setConsumerLag(
	topic string,
	partition int32,
	consumerGroup string,
	val int64,
) {
	m.consumerLagGaugeVec.WithLabelValues(topic, fmt.Sprintf("%d", partition), consumerGroup).Set(float64(val))
}
