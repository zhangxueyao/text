package logic

import (
	"context"
	"strconv"

	"github.com/zhangxueyao/item/item-api/internal/svc"
	"github.com/zhangxueyao/item/item-api/internal/types"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetItemLogic {
	return &GetItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetItemLogic) GetItem(id int64) (resp *types.Item, err error) {
	// 调 RPC
	item, err := l.svcCtx.ItemRpc.GetItem(l.ctx, &itemrpc.GetItemReq{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return &types.Item{
		Id:   strconv.FormatInt(item.Id, 10),
		Name: item.Name,
	}, nil
}
