package logic

import (
	"context"
	"time"

	cachex "github.com/zhangxueyao/item-rpc/internal/cache"
	"github.com/zhangxueyao/item-rpc/internal/model"
	"github.com/zhangxueyao/item-rpc/internal/svc"
	"github.com/zhangxueyao/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetItemLogic {
	return &GetItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetItemLogic) GetItem(in *itemrpc.GetItemReq) (*itemrpc.ItemResp, error) {
	// todo: add your logic here and delete this line
	key := cachex.ItemKey(in.Id)

	// 1) 本地
	if v, ok := l.svcCtx.LocalCache.Get(key); ok {
		return &itemrpc.ItemResp{
			Id:   in.Id,
			Name: v.(*model.Item).Name,
		}, nil
	}

	// 2) go-zero Redis 缓存封装（TakeCtx: miss 时回源 DB 并写回 Redis）
	var it model.Item
	err := l.svcCtx.Cache.TakeCtx(l.ctx, &it, key, func(val any) error {
		item, err := l.svcCtx.ItemModel.FindOne(l.ctx, in.Id)
		if err != nil {
			return err
		}
		*val.(**model.Item) = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 3) 回填 Local（短 TTL）
	l.svcCtx.LocalCache.SetWithTTL(key, &it, time.Duration(l.svcCtx.Config.LocalCacheTTL)*time.Second)
	return &itemrpc.ItemResp{
		Id:   in.Id,
		Name: it.Name,
	}, nil
}
