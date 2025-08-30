package outbox

import (
	"context"
	"time"

	"github.com/zhangxueyao/item/item-rpc/internal/model"
	"github.com/zhangxueyao/item/item-rpc/internal/mq/producer"

	"github.com/zeromicro/go-zero/core/logx"
)

type Relay struct {
	outbox model.OutboxModel
	prod   producer.Producer
}

func NewRelay(o model.OutboxModel, p producer.Producer) *Relay {
	return &Relay{outbox: o, prod: p}
}

func (r *Relay) Run(ctx context.Context) {
	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			events, err := r.outbox.ListNew(ctx, 1000)
			if err != nil {
				logx.Error(err)
				continue
			}
			for _, e := range events {
				if err := r.prod.Send(ctx, e.AggregateID, e); err != nil {
					logx.Errorf("publish failed id=%d: %v", e.EventID, err)
					continue
				}
				_ = r.outbox.MarkPublished(ctx, e.EventID)
			}
		}
	}
}
