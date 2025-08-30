package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/zhangxueyao/item/item-rpc/internal/config"
	"github.com/zhangxueyao/item/item-rpc/internal/mq/consumer"
	"github.com/zhangxueyao/item/item-rpc/internal/outbox"
	"github.com/zhangxueyao/item/item-rpc/internal/server"
	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/item.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	// 启动 Outbox Relay
	go outbox.NewRelay(ctx.OutboxModel, ctx.KafkaProd).Run(context.Background())
	// 启动 Consumer（伪示意：接入你的 Kafka 客户端消费循环）
	updater := consumer.NewCacheUpdater(ctx)
	_ = updater // 在你的 Kafka 消费回调里调用 updater.Handle

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		itemrpc.RegisterItemServer(grpcServer, server.NewItemServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
