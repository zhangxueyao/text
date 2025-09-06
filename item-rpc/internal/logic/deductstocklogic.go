package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductStockLogic {
	return &DeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 乐观锁：带 version
func (l *DeductStockLogic) DeductStock(in *itemrpc.DeductStockReq) (*itemrpc.DeductStockResp, error) {
	key := fmt.Sprintf("stock:sku:%d", in.Sku) // 失效的缓存 key
	topic := l.svcCtx.Config.TxnMsg.Topic      // 事务消息主题
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1) 读取当前版本
		var version int64
		stock, err := l.svcCtx.StockModel.FindOneBySku(ctx, in.Sku)
		if err != nil {
			return err
		}
		version = stock.Version
		// 2) 乐观锁扣减
		res, err := session.ExecCtx(ctx,
			"UPDATE stock SET qty=qty-?, version=version+1 WHERE sku=? AND version=? AND qty>=?",
			in.Delta, in.Sku, version, in.Delta)
		if err != nil {
			return err
		}
		aff, _ := res.RowsAffected()
		if aff == 0 {
			return sqlx.ErrNotFound
		} // 版本冲突/库存不足 → 交给上层处理或重试

		// 3) 写入事务消息（缓存失效）
		payload, _ := json.Marshal(map[string]any{"key": key})
		return l.svcCtx.TxnStore.AppendTx(ctx, session, topic, key, payload)
	})
	if err != nil {
		return nil, err
	}
	return &itemrpc.DeductStockResp{
		Success: true,
		Message: "Deduct Stock Success",
	}, nil
}
