package mqutils

import (
	"PrintServer/deploy"
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

func CreateWriter(mqcfg []string, topic string, partition int) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(mqcfg...),
		Topic:        topic,
		BatchTimeout: time.Millisecond,
	}

	return w
}

type MQMessage struct {
	Offset int64
	Key    []byte
	Value  []byte
}

func ConnectAndGetValuesWithTimeout(mqcfg deploy.AccessConfig, topic string, partition int, offset int64, timeout time.Duration) (messages []MQMessage, err error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{mqcfg.GetPath()},
		Topic:     topic,
		Partition: partition,
		MinBytes:  10e1, // 10B
		MaxBytes:  10e6, // 10MB
	})
	r.SetOffset(offset)

	// Завершаем работу
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			break
		}
		messages = append(messages, MQMessage{Offset: m.Offset, Key: m.Key, Value: m.Value})
	}
	r.Close()

	return messages, err
}
