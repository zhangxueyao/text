package mq

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"manager/internal/types"
)

type UserRegisterConsumer struct {
	consumer interface{} // 无需指定具体类型
}

// 实现ConsumeHandler接口的结构体
type userRegisterConsumeHandler struct {
	handler func(ctx context.Context, msg *types.UserRegisterMessage) error
}

// 实现Consume方法（关键修正）
func (h *userRegisterConsumeHandler) Consume(ctx context.Context, key string, value string) error {
	var m types.UserRegisterMessage
	if err := json.Unmarshal([]byte(value), &m); err != nil {
		logx.Errorf("解析消息失败: %v, key: %s, value: %s", err, key, value)
		return nil
	}
	return h.handler(ctx, &m)
}
func NewUserRegisterConsumer(conf kq.KqConf, handler func(ctx context.Context, msg *types.UserRegisterMessage) error) *UserRegisterConsumer {
	// 使用实现了ConsumeHandler接口的结构体
	h := &userRegisterConsumeHandler{
		handler: handler,
	}

	// 正确传递处理器接口
	c := kq.MustNewQueue(conf, h)

	return &UserRegisterConsumer{consumer: c}
}

// 启动消费者
func (c *UserRegisterConsumer) Start() error {
	// 通过类型断言调用Start方法
	if q, ok := c.consumer.(interface{ Start() }); ok {
		q.Start()
	}
	return nil
}

// 停止消费者
func (c *UserRegisterConsumer) Stop() error {
	// 通过类型断言调用Stop方法
	if q, ok := c.consumer.(interface{ Stop() }); ok {
		q.Stop()
	}
	return nil
}
