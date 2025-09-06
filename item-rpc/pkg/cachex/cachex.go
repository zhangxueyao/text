package cachex

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"github.com/zhangxueyao/item/item-rpc/pkg/dlock"
)

type Policy int

const (
	PInvalidate      Policy = iota + 1 // 强一致：删除失效 + 事件驱动
	PWriteThrough                      // 弱一致：写透
	PListWithVersion                   // 聚合：版本号
)

type Manager struct {
	Rdb                             *redis.Client
	SF                              singleflight.Group
	Local                           *ristretto.Cache
	EnableDLock                     bool
	LockTTL, LockMaxLease, Watchdog time.Duration
}

type Loader[T any] func(ctx context.Context) (T, error)

func jittered(ttl time.Duration) time.Duration {
	n := time.Now().UnixNano()
	off := (n%3 - 1)
	return ttl + time.Duration(int64(ttl)/10*off)
}

func GetWithPolicy[T any](
	ctx context.Context, m *Manager, key string, ttl time.Duration, pol Policy, loader Loader[T],
) (T, error) {
	var zero T
	// 1) 本地
	if v, ok := m.Local.Get(key); ok {
		return v.(T), nil
	}
	// 2) Redis
	if bs, err := m.Rdb.Get(ctx, key).Bytes(); err == nil {
		var v T
		if json.Unmarshal(bs, &v) == nil {
			m.Local.SetWithTTL(key, v, 1, ttl)
			return v, nil
		}
	}
	// 3) singleflight + （可选）分布式锁
	anyV, err, _ := m.SF.Do(key, func() (any, error) {
		// 再查一次 Redis（双检）
		if bs, err := m.Rdb.Get(ctx, key).Bytes(); err == nil {
			var vv T
			if json.Unmarshal(bs, &vv) == nil {
				m.Local.SetWithTTL(key, vv, 1, ttl)
				return vv, nil
			}
		}
		// 分布式锁（只保护“回源重建”）
		var lk *dlock.Lock
		if m.EnableDLock {
			l, ok, e := dlock.Acquire(ctx, m.Rdb, "lock:"+key, m.LockTTL, m.LockMaxLease, m.Watchdog)
			if e != nil {
				return zero, e
			}
			if !ok { // 取不到锁：轮询等回填
				deadline := time.Now().Add(2 * time.Second)
				for time.Now().Before(deadline) {
					if bs, err := m.Rdb.Get(ctx, key).Bytes(); err == nil {
						var v2 T
						if json.Unmarshal(bs, &v2) == nil {
							m.Local.SetWithTTL(key, v2, 1, ttl)
							return v2, nil
						}
					}
					time.Sleep(40 * time.Millisecond)
				}
				return zero, fmt.Errorf("rebuild timeout")
			}
			lk = l
			defer lk.Release(context.Background())
		}
		// 真正回源
		v, err := loader(ctx)
		if err != nil {
			return zero, err
		}
		// 默认退化成 Invalidate
		blob, _ := json.Marshal(v)
		_ = m.Rdb.Set(ctx, key, blob, ttl).Err()
		m.Local.SetWithTTL(key, v, 1, ttl)
		return v, nil
	})
	if err != nil {
		return zero, err
	}
	return anyV.(T), nil
}

func (m *Manager) UpdateStrong(
	ctx context.Context,
	tx func(context.Context) error,
	publish func(context.Context, string) error,
	keys ...string,
) error {
	if err := tx(ctx); err != nil {
		return err
	}
	for _, k := range keys {
		_ = publish(ctx, k)
	}
	return nil
}

func UpdateWeak[T any](
	ctx context.Context, m *Manager, key string, ttl time.Duration, v T,
	tx func(context.Context) error,
) error {
	if err := tx(ctx); err != nil {
		return err
	}
	blob, _ := json.Marshal(v)
	if err := m.Rdb.Set(ctx, key, blob, jittered(ttl)).Err(); err != nil {
		_ = m.Rdb.Del(ctx, key).Err()
	}
	m.Local.SetWithTTL(key, v, 1, ttl)
	return nil
}

func (m *Manager) BumpVersion(ctx context.Context, verKey string) error {
	return m.Rdb.Incr(ctx, verKey).Err()
}

func (m *Manager) Invalidate(ctx context.Context, key string) {
	_ = m.Rdb.Del(ctx, key).Err()
	m.Local.Del(key)
}
