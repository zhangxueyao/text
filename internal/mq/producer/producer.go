package producer

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	Send(ctx context.Context, key int64, v any) error
}

type syncProducer struct {
	sp    sarama.SyncProducer
	topic string
}

func NewSyncProducer(brokers []string, topic string) Producer {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Idempotent = true
	p, _ := sarama.NewSyncProducer(brokers, cfg)
	return &syncProducer{sp: p, topic: topic}
}

func (p *syncProducer) Send(ctx context.Context, key int64, v any) error {
	b, _ := json.Marshal(v)
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(string(rune(key))), // 简化：真实项目请用稳定字节序列化
		Value: sarama.ByteEncoder(b),
	}
	_, _, err := p.sp.SendMessage(msg)
	return err
}
