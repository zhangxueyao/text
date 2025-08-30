package logic

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zhangxueyao/item/item-rpc/internal/model"
	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateItemLogic {
	return &UpdateItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateItemLogic) UpdateItem(in *itemrpc.UpdateItemReq) (*itemrpc.UpdateItemResp, error) {
	it := &model.Item{Id: in.Id, Name: in.Name}

	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := l.svcCtx.ItemModel.TxUpdate(ctx, session, it); err != nil {
			return err
		}
		// 写 Outbox
		payload, _ := json.Marshal(it) // 最小必要字段
		eventID, err := l.svcCtx.Snowflake.GenerateID()
		if err != nil {
			return err
		}
		evt := &model.OutboxEvent{
			EventID:     eventID, // 你的 Snowflake 生成器
			Aggregate:   "item-api",
			AggregateID: it.Id,
			EventType:   "UPDATED",
			Payload:     payload,
		}
		return l.svcCtx.OutboxModel.TxInsert(ctx, session, evt)
	})
	if err != nil {
		return nil, err
	}
	return &itemrpc.UpdateItemResp{}, nil
}
