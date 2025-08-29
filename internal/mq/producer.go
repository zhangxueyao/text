package mq

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-queue/kq"
	"manager/internal/types"
)

type UserRegisterProducer struct {
	pusher *kq.Pusher
}

// 初始化生产者（实际是Pusher）
func NewUserRegisterProducer(conf kq.KqConf) *UserRegisterProducer {
	p := kq.NewPusher(conf.Brokers, conf.Topic) // 使用NewPusher创建实例
	return &UserRegisterProducer{pusher: p}
}

// 发送消息
func (p *UserRegisterProducer) Send(ctx context.Context, msg types.UserRegisterMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.pusher.Push(ctx, string(data)) // 使用Push方法发送消息
}

func (p *UserRegisterProducer) Close() error {
	// Pusher不需要显式关闭，底层连接池会自动管理
	return nil
}
