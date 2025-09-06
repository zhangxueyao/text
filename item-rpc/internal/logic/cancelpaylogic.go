package logic

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelPayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelPayLogic {
	return &CancelPayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelPayLogic) CancelPay(in *itemrpc.PayCancelReq) (*itemrpc.PayAck, error) {
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 幂等读取（可能 Try 未写成功：空回滚场景 → 直接写 CANCEL 记录后返回）
		var curState sql.NullInt64
		if err := session.QueryRowCtx(ctx, &curState,
			"SELECT state FROM tcc_branch WHERE xid=? AND branch=? FOR UPDATE", in.Xid, "pay"); err != nil && err != sqlx.ErrNotFound {
			return err
		}
		if curState.Valid {
			if curState.Int64 == stateCAL {
				return nil
			}
			if curState.Int64 == stateCFM {
				// 已确认则不再回滚
				return nil
			}
		}

		// 若已 TRY 过：解冻回滚
		_, err := session.ExecCtx(ctx, `
			UPDATE account
			SET freeze_amount = freeze_amount - ?, balance = balance + ?, version=version+1
			WHERE id=? AND freeze_amount >= ?`,
			in.Amount, in.Amount, in.UserId, in.Amount)
		if err != nil {
			// 空回滚：没 Try 也允许成功（不报错）
			// 可选择仅在存在TRY时才做这步
		}

		// 标记/插入 CANCEL
		_, err = session.ExecCtx(ctx, `
			INSERT INTO tcc_branch(xid, branch, state, expire_at)
			VALUES(?, 'pay', ?, NOW())
			ON DUPLICATE KEY UPDATE state=VALUES(state), expire_at=VALUES(expire_at)
		`, in.Xid, stateCAL)
		return err
	})

	return &itemrpc.PayAck{
		Ok: err == nil,
	}, nil
}
