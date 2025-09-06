package mq

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	w *kafka.Writer
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		w: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaPublisher) PublishInvalidate(ctx context.Context, key string) error {
	return p.w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: []byte(key), // 简化：直接用 key 作消息体；也可用 JSON
	})
}

func (p *KafkaPublisher) Close() error { return p.w.Close() }
