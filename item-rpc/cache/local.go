package cachex

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

type Local interface {
	Get(key string) (any, bool)
	SetWithTTL(key string, val any, ttl time.Duration)
	Del(key string)
}

type localCache struct {
	rc  *ristretto.Cache
	ttl time.Duration
}

func NewLocalCache(defaultTTL time.Duration) Local {
	rc, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6, MaxCost: 1 << 28, BufferItems: 64,
	})
	return &localCache{rc: rc, ttl: defaultTTL}
}

func (l *localCache) Get(key string) (any, bool) {
	v, ok := l.rc.Get(key)
	return v, ok
}

func (l *localCache) SetWithTTL(key string, val any, ttl time.Duration) {
	if ttl <= 0 {
		ttl = l.ttl
	}
	l.rc.SetWithTTL(key, val, 1, ttl)
}

func (l *localCache) Del(key string) { l.rc.Del(key) }
