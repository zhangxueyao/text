package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	CacheRedis cache.CacheConf `json:"CacheRedis"`
	Redis      RedisConf
	Kafka      KafkaConf

	LocalCacheTTL          int
	EventDedupTTL          int
	LocalInvalidateChannel string
}

type KafkaConf struct {
	Brokers []string
	Topic   string
}

type RedisConf struct {
	Addr string
	Pass string
}
