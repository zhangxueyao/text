package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cachex "manager/internal/cache"
	"manager/internal/model"
	"manager/internal/svc"
)

type CacheUpdater struct {
	svc   *svc.ServiceContext
	topic string
}

func NewCacheUpdater(s *svc.ServiceContext) *CacheUpdater { return &CacheUpdater{svc: s} }

// 伪代码：集成你选用的 Kafka client 的消费循环；这里只写处理函数
func (c *CacheUpdater) Handle(ctx context.Context, val []byte) error {
	var evt model.OutboxEvent
	if err := json.Unmarshal(val, &evt); err != nil {
		return err
	}

	// 幂等：event_id 去重（Redis SET NX）
	seenKey := fmt.Sprintf("evt:%d", evt.EventID)
	ok, err := c.svc.Redis.SetNX(ctx, seenKey, 1, time.Duration(c.svc.Config.EventDedupTTL)*time.Second).Result()
	if err != nil {
		return err
	}
	if !ok {
		// 已处理
		return nil
	}

	key := cachex.ItemKey(evt.AggregateID)

	switch evt.EventType {
	case "DELETED":
		// 失效 Redis
		_ = c.svc.Cache.DelCtx(ctx, key)
		// 本地失效广播
		_ = c.svc.Redis.Publish(ctx, c.svc.Config.LocalInvalidateChannel, key).Err()

	case "CREATED", "UPDATED":
		// 写透 Redis（用事件 payload 的“最小视图”避免脏字段）
		var it model.Item
		_ = json.Unmarshal(evt.Payload, &it)
		// go-zero cache 没有直接 SetCtx 任意对象的 API；这里可用 Redis 原生写：
		b, _ := json.Marshal(it)
		_ = c.svc.Redis.Set(ctx, key, b, 10*time.Minute).Err()

		// 本地失效广播（确保各实例 LocalCache 失效）
		_ = c.svc.Redis.Publish(ctx, c.svc.Config.LocalInvalidateChannel, key).Err()
	}

	return nil
}
