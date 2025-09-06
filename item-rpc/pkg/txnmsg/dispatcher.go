package txnmsg

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/zeromicro/go-queue/kq"
)

type Dispatcher struct {
	Store     *Store
	Pusher    *kq.Pusher
	BatchSize int
	Interval  time.Duration
	MaxRetry  int
	owner     string
}

func NewDispatcher(store *Store, pusher *kq.Pusher, bs int, itv time.Duration, maxRetry int) *Dispatcher {
	host, _ := os.Hostname()
	return &Dispatcher{
		Store: store, Pusher: pusher,
		BatchSize: bs, Interval: itv, MaxRetry: maxRetry,
		owner: fmt.Sprintf("%s-%d", host, time.Now().UnixNano()),
	}
}

func (d *Dispatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			n, err := d.Store.Claim(ctx, d.owner, d.BatchSize)
			if err != nil || n == 0 {
				continue
			}
			msgs, err := d.Store.FetchClaimed(ctx, d.owner, d.BatchSize)
			if err != nil {
				continue
			}
			for _, m := range msgs {
				if err := d.Pusher.KPush(ctx, m.Key, string(m.Body)); err != nil {
					if m.Retry+1 >= d.MaxRetry {
						_ = d.Store.MarkFailed(ctx, m.ID)
					} else {
						_ = d.Store.RetryLater(ctx, m.ID, m.Retry)
					}
					continue
				}
				_ = d.Store.MarkSent(ctx, m.ID)
			}
		}
	}
}

func (d *Dispatcher) Stop(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
}
