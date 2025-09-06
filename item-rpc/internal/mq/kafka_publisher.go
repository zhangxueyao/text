package mq

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type HandleFunc func(ctx context.Context, key string) error

type KafkaConsumer struct {
	r      *kafka.Reader
	handle HandleFunc
}

func NewKafkaConsumer(brokers []string, group, topic string, h HandleFunc) *KafkaConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: group,
		Topic:   topic,
	})
	return &KafkaConsumer{r: r, handle: h}
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	for {
		m, err := c.r.ReadMessage(ctx)
		if err != nil {
			return err
		}
		key := string(m.Key)
		if key == "" {
			key = string(m.Value)
		}
		if err := c.handle(ctx, key); err != nil {
			log.Printf("invalidate handle err: %v key=%s", err, key)
		}
	}
}

func (c *KafkaConsumer) Close() error { return c.r.Close() }
