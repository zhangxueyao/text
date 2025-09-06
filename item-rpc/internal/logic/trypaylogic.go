package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zhangxueyao/item/item-rpc/internal/model"
	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TryPayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTryPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TryPayLogic {
	return &TryPayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

const (
	stateTRY = 0
	stateCFM = 1
	stateCAL = 2
)

type tccPayload struct {
	UserID  int64 `json:"user_id"`
	OrderID int64 `json:"order_id"`
	Amount  int64 `json:"amount"`
}

func (l *TryPayLogic) TryPay(in *itemrpc.PayTryReq) (*itemrpc.PayAck, error) {
	pl := tccPayload{UserID: in.UserId, OrderID: in.OrderId, Amount: in.Amount}
	js, _ := json.Marshal(pl)
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 幂等：若已 Confirm/Cancel，直接返回成功（防悬挂/乱序）
		txn, err := l.svcCtx.TccBranchModel.FindOneByXidBranch(ctx, in.Xid, "pay")
		if err != nil && err != model.ErrNotFound {
			return err
		}
		if txn != nil {
			switch txn.State {
			case stateCFM, stateCAL:
				return nil // 已处理
			case stateTRY:
				// 已try过，刷新过期时间即可（续租防悬挂）
				_, err := session.ExecCtx(ctx,
					"UPDATE tcc_branch SET expire_at=? WHERE xid=? AND branch=?",
					time.Now().Add(2*time.Minute), in.Xid, "pay")
				return err
			}
		}
		// 业务冻结：余额 >= amount，扣 balance，加 freeze_amount
		// 乐观：用 balance>=amount 条件约束即可（也可引入 version）
		res, err := session.ExecCtx(ctx, `
			UPDATE account
			SET balance = balance - ?, freeze_amount = freeze_amount + ?, version=version+1
			WHERE id = ? AND balance >= ?`, in.Amount, in.Amount, in.UserId, in.Amount)
		if err != nil {
			return err
		}
		// 检查影响行数
		aff, _ := res.RowsAffected()
		if aff == 0 {
			return errors.New("insufficient balance")
		}
		// 记录/插入 TRY
		// 记录/插入 TRY
		_, err = session.ExecCtx(ctx, `
			INSERT INTO tcc_branch(xid, branch, state, payload, expire_at)
			VALUES(?, 'pay', ?, ?, ?)
			ON DUPLICATE KEY UPDATE state=VALUES(state), payload=VALUES(payload), expire_at=VALUES(expire_at)
		`, in.Xid, stateTRY, js, time.Now().Add(2*time.Minute))
		return err
	})

	return &itemrpc.PayAck{Ok: err == nil}, err
}
