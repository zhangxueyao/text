package logic

import (
	"context"
	"strconv"

	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *itemrpc.GetUserReq) (*itemrpc.GetUserResp, error) {
	// 校验手机号是否已注册
	user, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err != nil {
		return nil, err
	}
	age, err := strconv.Atoi(user.Age)
	if err != nil {
		return nil, err
	}
	return &itemrpc.GetUserResp{
		User: &itemrpc.User{
			Password: user.Password,
			Mobile:   user.Mobile,
			Email:    user.Email,
			Gender:   user.Gender,
			Age:      int32(age),
		},
	}, nil
}
