package logic

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zhangxueyao/item/item-api/internal/svc"
	"github.com/zhangxueyao/item/item-api/internal/types"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	code, ok := l.svcCtx.CodeStore.Get(req.Mobile)
	if !ok || code != req.Code {
		return nil, errors.New("invalid code")
	}
	l.svcCtx.CodeStore.Delete(req.Mobile)

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
