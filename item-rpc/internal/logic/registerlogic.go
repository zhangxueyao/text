package logic

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zhangxueyao/item/item-rpc/internal/model"
	"github.com/zhangxueyao/item/item-rpc/internal/svc"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *itemrpc.RegisterReq) (*itemrpc.RegisterResp, error) {

	// 校验手机号是否已注册
	_, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.User.Mobile)
	if err == nil {
		return nil, errors.New("mobile is registered")
	}
	// 校验邮箱是否已注册
	_, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, in.User.Email)
	if err == nil {
		return nil, errors.New("email is registered")
	}

	// 插入用户
	_, err = l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Password: in.User.Password,
		Mobile:   in.User.Mobile,
		Email:    in.User.Email,
		Gender:   in.User.Gender,
		Age:      strconv.Itoa(int(in.User.Age)),
		CreateAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdateAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return &itemrpc.RegisterResp{
		Success: true,
		Message: "register success",
	}, nil
}
