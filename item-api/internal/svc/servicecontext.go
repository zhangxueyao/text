package svc

import (
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zhangxueyao/item/item-api/internal/config"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"
)

type ServiceContext struct {
	Config       config.Config
	ItemRpc      itemrpc.ItemClient
	CaptchaStore *MemoryStore
	CodeStore    *MemoryStore
	UserStore    *UserStore
}

func NewServiceContext(c config.Config) *ServiceContext {

	rand.Seed(time.Now().UnixNano())
	cli := zrpc.MustNewClient(c.ItemRpc)

	sc := &ServiceContext{
		Config:       c,
		ItemRpc:      itemrpc.NewItemClient(cli.Conn()),
		CaptchaStore: NewMemoryStore(),
		CodeStore:    NewMemoryStore(),
		UserStore:    NewUserStore(),
	}

	return sc
}
