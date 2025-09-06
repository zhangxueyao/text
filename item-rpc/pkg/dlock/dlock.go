package dlock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

type Lock struct {
	rdb                 *redis.Client
	key, token          string
	ttl, maxLease, tick time.Duration
	cancel              context.CancelFunc
	leased              time.Duration
}

func randToken() string { b := make([]byte, 16); rand.Read(b); return hex.EncodeToString(b) }

func Acquire(ctx context.Context, rdb *redis.Client, key string, ttl, maxLease, tick time.Duration) (*Lock, bool, error) {
	token := randToken()
	ok, err := rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil || !ok {
		return nil, ok, err
	}

	cctx, cancel := context.WithCancel(context.Background())
	l := &Lock{rdb: rdb, key: key, token: token, ttl: ttl, maxLease: maxLease, tick: tick, cancel: cancel}
	go l.watchdog(cctx)
	return l, true, nil
}

func (l *Lock) watchdog(ctx context.Context) {
	t := time.NewTicker(l.tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// 超过最大租约直接放弃续期
			l.leased += l.tick
			if l.maxLease > 0 && l.leased >= l.maxLease {
				return
			}
			val, _ := l.rdb.Get(ctx, l.key).Result()
			if val == l.token {
				l.rdb.Expire(ctx, l.key, l.ttl) // 续锁
			} else {
				return
			}
		}
	}
}

func (l *Lock) Release(ctx context.Context) {
	l.cancel()
	// Lua 安全释放
	_ = l.rdb.Eval(ctx, `
        if redis.call("get", KEYS[1])==ARGV[1] then
            return redis.call("del", KEYS[1]) else return 0 end
    `, []string{l.key}, l.token).Err()
}
