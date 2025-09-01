package logic

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zhangxueyao/item/item-api/internal/svc"
	"github.com/zhangxueyao/item/item-api/internal/types"
	"github.com/zhangxueyao/item/item-rpc/itemrpc"
)

type PasswordLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasswordLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasswordLoginLogic {
	return &PasswordLoginLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *PasswordLoginLogic) Login(req *types.PasswordLoginReq) (*types.LoginResp, error) {
	user, err := l.svcCtx.ItemRpc.GetUser(l.ctx, &itemrpc.GetUserReq{
		Mobile: req.Mobile,
	})
	if err != nil {
		return nil, err
	}
	if user.User.Password != req.Password {
		return nil, errors.New("invalid credentials")
	}
	now := time.Now().Unix()
	expire := now + l.svcCtx.Config.Auth.AccessExpire
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"mobile": req.Mobile,
		"iat":    now,
		"exp":    expire,
	})
	tokenStr, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		return nil, err
	}
	return &types.LoginResp{Token: tokenStr, Expire: expire}, nil
}
