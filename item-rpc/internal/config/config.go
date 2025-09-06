package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/zrpc"
)

type RedisConf struct {
	Addr string `json:"Addr"`
	Pass string `json:"Pass"`
	DB   int    `json:"DB"`
}
type DBConf struct {
	DSN string `json:"DSN"`
}

type CacheConf struct {
	EnableDLock    bool `json:"EnableDLock"`    // 多实例强热点开启
	DefaultTTL     int  `json:"DefaultTTL"`     // 秒
	WatchdogMs     int  `json:"WatchdogMs"`     // 看门狗续锁间隔
	LockTTLMs      int  `json:"LockTTLMs"`      // 锁TTL
	LockMaxLeaseMs int  `json:"LockMaxLeaseMs"` // 锁最长存活
}

type TxnMsgConf struct {
	Enable     bool   `json:"Enable"`
	BatchSize  int    `json:"BatchSize"`
	IntervalMs int    `json:"IntervalMs"`
	MaxRetry   int    `json:"MaxRetry"`
	Topic      string `json:"Topic"`
	DLQTopic   string `json:"DLQTopic"`
}

type KqProducerConf struct {
	Brokers []string `json:"Brokers"`
	Topic   string   `json:"Topic"`
}

type Config struct {
	zrpc.RpcServerConf
	Redis      RedisConf      `json:"Redis"`
	DB         DBConf         `json:"DB"`
	Cache      CacheConf      `json:"CacheConf"`
	TxnMsg     TxnMsgConf     `json:"TxnMsg"`
	KqPusher   KqProducerConf `json:"KqPusher"`   // 事务消息 生产配置（指向 TxnMsg.Topic）
	KqConsumer kq.KqConf      `json:"KqConsumer"` // 失效消费者配置（消费 TxnMsg.Topic）
}
