package logic

import (
	"context"

	"manager/internal/svc"
	"manager/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManagerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManagerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManagerLogic {
	return &ManagerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ManagerLogic) Manager(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}
