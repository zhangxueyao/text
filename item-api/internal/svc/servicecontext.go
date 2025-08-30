package svc

import (
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zhangxueyao/item/item-api/internal/config"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"
)

type ServiceContext struct {
	Config  config.Config
	ItemRpc itemrpc.ItemClient
}

func NewServiceContext(c config.Config) *ServiceContext {

	cli := zrpc.MustNewClient(c.ItemRpc)

	sc := &ServiceContext{
		Config:  c,
		ItemRpc: itemrpc.NewItemClient(cli.Conn()),
	}

	return sc
}
