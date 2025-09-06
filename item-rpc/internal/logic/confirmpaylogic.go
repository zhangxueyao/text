package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmPayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmPayLogic {
	return &ConfirmPayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Confirm: 实际扣款，写支付流水，减去冻结
func (l *ConfirmPayLogic) ConfirmPay(in *itemrpc.PayConfirmReq) (*itemrpc.PayAck, error) {
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		topic := l.svcCtx.Config.TxnMsg.Topic // 事务消息主题
		// 幂等读取
		var curState int
		if err := session.QueryRowCtx(ctx, &curState,
			"SELECT state FROM tcc_branch WHERE xid=? AND branch=? FOR UPDATE", in.Xid, "pay"); err != nil {
			// 没有 TRY 记录：可认为空回滚已发生/悬挂，直接返回成功或按策略处理
			return nil
		}
		if curState == stateCFM {
			return nil // 幂等
		}
		if curState == stateCAL {
			// 已回滚则不再确认
			return nil
		}

		// 减冻结，不再回加 balance；并落支付流水（幂等唯一约束）
		_, err := session.ExecCtx(ctx, `
			UPDATE account
			SET freeze_amount = freeze_amount - ?, version=version+1
			WHERE id=? AND freeze_amount >= ?`,
			in.Amount, in.UserId, in.Amount)
		if err != nil {
			return err
		}
		id, err := l.svcCtx.Snowflake.GenerateID()
		if err != nil {
			return err
		}
		_, err = session.ExecCtx(ctx, `
			INSERT INTO pay_txn(id, user_id, order_id, amount, status)
			VALUES(?, ?, ?, ?, 1)
			ON DUPLICATE KEY UPDATE status=VALUES(status)`,
			id, in.UserId, in.OrderId, in.Amount)
		if err != nil {
			return err
		}

		// 标记 CONFIRM
		_, err = session.ExecCtx(ctx,
			"UPDATE tcc_branch SET state=? WHERE xid=? AND branch=?",
			stateCFM, in.Xid, "pay")
		if err != nil {
			return err
		}

		// 写 Outbox（由现有 Dispatcher/KQ 异步删缓存/发积分等）
		payload := map[string]any{
			"event":    "pay_success",
			"user_id":  in.UserId,
			"order_id": in.OrderId,
			"amount":   in.Amount,
			"at":       time.Now().Unix(),
		}
		return l.svcCtx.TxnStore.AppendTx(ctx, session, topic, in.Xid, payload)
	})

	return &itemrpc.PayAck{Ok: err == nil}, err
}
