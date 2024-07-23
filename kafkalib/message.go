package kafkalib

type Message struct {
	Topic     string
	Key       []byte
	Payload   []byte
	Partition *int32
}
