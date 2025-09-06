package logic

import (
	"context"
	"fmt"

	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"
	"github.com/zhangxueyao/item/item-rpc/pkg/cachex"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStockLogic {
	return &GetStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetStockLogic) GetStock(in *itemrpc.GetStockReq) (*itemrpc.GetStockResp, error) {
	key := fmt.Sprintf("stock:sku:%d", in.Sku)
	ttl := l.svcCtx.DefaultTTL
	v, err := cachex.GetWithPolicy[*itemrpc.GetStockResp](
		l.ctx, l.svcCtx.CacheMgr, key, ttl, 1, // PInvalidate
		func(ctx context.Context) (*itemrpc.GetStockResp, error) {
			// 真正回源 DB
			var s itemrpc.GetStockResp
			stock, err := l.svcCtx.StockModel.FindOneBySku(ctx, in.Sku)
			if err != nil {
				return nil, err
			}
			s.Sku = stock.Sku
			s.Qty = stock.Qty
			s.Version = stock.Version
			return &s, nil
		})
	if err != nil {
		return nil, err
	}
	return v, nil
}
