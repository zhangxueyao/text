package svc

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/queue"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zhangxueyao/item/item-rpc/internal/config"
	"github.com/zhangxueyao/item/item-rpc/internal/model"
	"github.com/zhangxueyao/item/item-rpc/pkg"
	"github.com/zhangxueyao/item/item-rpc/pkg/cachex"
	"github.com/zhangxueyao/item/item-rpc/pkg/txnmsg"
)

// 适配器：让闭包实现 service.Service 接口
type funcService struct {
	start func()
	stop  func()
}

func (f *funcService) Start() { f.start() }
func (f *funcService) Stop()  { f.stop() }

type ServiceContext struct {
	Config      config.Config
	DB          sqlx.SqlConn
	Rdb         *redis.Client
	KqPusher    *kq.Pusher
	KqQueue     queue.MessageQueue
	CacheMgr    *cachex.Manager
	TxnStore    *txnmsg.Store
	Dispatcher  *txnmsg.Dispatcher
	DefaultTTL  time.Duration
	ItemModel   model.ItemModel
	UserModel   model.UserModel
	StockModel  model.StockModel
	OutboxModel model.OutboxModel
	Snowflake   *pkg.SnowflakeIDGenerator
	Group       *service.ServiceGroup
}

func NewServiceContext(c config.Config) *ServiceContext {
	// DB
	db := sqlx.NewMysql(c.DB.DSN)

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: c.Redis.Addr, Password: c.Redis.Pass, DB: c.Redis.DB,
	})

	// cache manager
	local, _ := ristretto.NewCache(&ristretto.Config{NumCounters: 1e6, MaxCost: 1 << 28, BufferItems: 64})
	cm := &cachex.Manager{
		Rdb: rdb, Local: local, EnableDLock: c.Cache.EnableDLock,
		LockTTL:      time.Duration(c.Cache.LockTTLMs) * time.Millisecond,
		LockMaxLease: time.Duration(c.Cache.LockMaxLeaseMs) * time.Millisecond,
		Watchdog:     time.Duration(c.Cache.WatchdogMs) * time.Millisecond,
	}
	// kq pusher/consumer
	pusher := kq.NewPusher(c.KqPusher.Brokers, c.KqPusher.Topic)
	store := txnmsg.NewStore(db)
	dispatcher := txnmsg.NewDispatcher(store, pusher, c.TxnMsg.BatchSize, time.Duration(c.TxnMsg.IntervalMs)*time.Millisecond, c.TxnMsg.MaxRetry)

	newQueue := kq.MustNewQueue(c.KqConsumer, kq.WithHandle(func(ctx context.Context, key, val string) error {
		// payload 可包含 {"key":"stock:sku:123"}
		// 简化：直接用 key，即消息键就是要失效的缓存 key
		cm.Invalidate(ctx, key)
		return nil
	}))
	snowflake := pkg.NewSnowflakeIDGenerator(1)
	// 统一生命周期
	g := service.NewServiceGroup()

	// 把“事务消息调度器”纳入 ServiceGroup（用 funcService 适配）
	if c.TxnMsg.Enable {
		g.Add(&funcService{
			start: func() {
				go func() {
					_ = dispatcher.Start(context.Background())
				}()

			},
			stop: func() {
				dispatcher.Stop(context.Background())
			},
		})
	}

	// kq.Queue 自带 Start/Stop，实现了 service.Service，直接加入
	g.Add(newQueue) // 启动 kq 消费者

	return &ServiceContext{
		Config:      c,
		DB:          db,
		Rdb:         rdb,
		KqPusher:    pusher,
		KqQueue:     newQueue,
		CacheMgr:    cm,
		TxnStore:    store,
		Dispatcher:  dispatcher,
		DefaultTTL:  time.Duration(c.Cache.DefaultTTL) * time.Second,
		ItemModel:   model.NewItemModel(db),
		UserModel:   model.NewUserModel(db),
		OutboxModel: model.NewOutboxModel(db),
		StockModel:  model.NewStockModel(db),
		Snowflake:   snowflake,
		Group:       g,
	}
}
