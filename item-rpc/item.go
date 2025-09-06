package main

import (
	"flag"
	"fmt"

	"github.com/zhangxueyao/item/item-rpc/internal/config"
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
	// 1) 构建业务上下文（里面已把 kq 队列、事务消息调度器等加入了 ServiceGroup）
	ctx := svc.NewServiceContext(c)
	// 2) 统一启动/停止后台服务（事务消息调度器、kq 消费者等）
	go ctx.Group.Start()
	defer ctx.Group.Stop()
	// 3) 创建并启动 gRPC Server
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		itemrpc.RegisterItemServer(grpcServer, server.NewItemServer(ctx))
		// 仅在开发/测试环境打开反射
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
