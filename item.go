package main

import (
	"context"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"manager/internal/config"
	"manager/internal/mq/consumer"
	"manager/internal/outbox"
	"manager/internal/svc"
)

func main() {
	var c config.Config
	conf.MustLoad("etc/item-api-api.yaml", &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	s := svc.NewServiceContext(c)

	// 启动 Outbox Relay
	go outbox.NewRelay(s.OutboxModel, s.KafkaProd).Run(context.Background())

	// 启动 Consumer（伪示意：接入你的 Kafka 客户端消费循环）
	updater := consumer.NewCacheUpdater(s)
	_ = updater // 在你的 Kafka 消费回调里调用 updater.Handle

	// TODO: 注册路由（GetItem / UpdateItem）
	server.Start()
}
