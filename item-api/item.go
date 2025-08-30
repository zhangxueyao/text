package main

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zhangxueyao/item/item-api/internal/handler"

	"github.com/zhangxueyao/item/item-api/internal/config"
	"github.com/zhangxueyao/item/item-api/internal/svc"
)

func main() {
	var c config.Config
	conf.MustLoad("etc/item-api.yaml", &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	handler.RegisterHandlers(server, ctx)

	server.Start()
}
