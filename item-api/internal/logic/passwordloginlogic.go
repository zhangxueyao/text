package logic

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zhangxueyao/item/item-api/internal/svc"
	"github.com/zhangxueyao/item/item-api/internal/types"
)

type PasswordLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasswordLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PasswordLoginLogic {
	return &PasswordLoginLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *PasswordLoginLogic) Login(req *types.PasswordLoginReq) (*types.LoginResp, error) {
	if !l.svcCtx.UserStore.ValidatePassword(req.Mobile, req.Password) {
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
