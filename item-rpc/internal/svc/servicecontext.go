package svc

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
	cachex "github.com/zhangxueyao/item/item-rpc/cache"
	"github.com/zhangxueyao/item/item-rpc/internal/config"
	"github.com/zhangxueyao/item/item-rpc/internal/model"
	"github.com/zhangxueyao/item/item-rpc/internal/mq/producer"
	"github.com/zhangxueyao/item/item-rpc/pkg"
)

type ServiceContext struct {
	Config     config.Config
	DB         sqlx.SqlConn
	Cache      cache.Cache  // go-zero 分布式缓存（Redis）
	LocalCache cachex.Local // 自定义本地缓存（ristretto/bigcache 封装）
	Redis      *redis.Client
	KafkaProd  producer.Producer

	ItemModel   model.ItemModel
	UserModel   model.UserModel
	OutboxModel model.OutboxModel
	Snowflake   *pkg.SnowflakeIDGenerator
}

func NewServiceContext(c config.Config) *ServiceContext {
	// DB
	db := sqlx.NewMysql(c.DataSource)
	// go-zero cache（Redis）
	cc := cache.New(
		c.CacheRedis,            // ClusterConf
		syncx.NewSingleFlight(), // 并发去重器
		nil,                     // 统计指标(可先传nil，后续需要再加)
		sqlx.ErrNotFound,        // “未找到”哨兵错误（你用的是 sqlx，就传 sqlx.ErrNotFound）
	)

	// 本地缓存
	lc := cachex.NewLocalCache(time.Duration(c.LocalCacheTTL) * time.Second)

	// Redis
	rdb := redis.NewClient(&redis.Options{Addr: c.Redis.Addr, Password: c.Redis.Pass})

	// Kafka Producer
	kp := producer.NewSyncProducer(c.Kafka.Brokers, c.Kafka.Topic)

	snowflake := pkg.NewSnowflakeIDGenerator(1)

	sc := &ServiceContext{
		Config:      c,
		DB:          db,
		Cache:       cc,
		LocalCache:  lc,
		Redis:       rdb,
		KafkaProd:   kp,
		ItemModel:   model.NewItemModel(db),
		UserModel:   model.NewUserModel(db),
		OutboxModel: model.NewOutboxModel(db),
		Snowflake:   snowflake,
	}
	// 订阅本地失效频道，跨实例同步清理 LocalCache
	go func() {
		ctx := context.Background()
		sub := rdb.Subscribe(ctx, c.LocalInvalidateChannel)
		ch := sub.Channel()
		for msg := range ch {
			sc.LocalCache.Del(msg.Payload)
		}
	}()

	logx.Infof("ServiceContext initialized")
	return sc
}
