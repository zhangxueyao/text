package logic

import (
	"context"
	"errors"
	"net/mail"
	"unicode"

	"github.com/zhangxueyao/item/item-api/internal/svc"
	"github.com/zhangxueyao/item/item-api/internal/types"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{ctx: ctx, svcCtx: svcCtx}
}

func isWeakPassword(pw string) bool {
	if len(pw) < 8 {
		return true
	}
	var hasLetter, hasDigit bool
	for _, r := range pw {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return !(hasLetter && hasDigit)
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.RegisterResp, error) {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, errors.New("invalid email")
	}
	if isWeakPassword(req.Password) {
		return nil, errors.New("weak password")
	}
	err := l.svcCtx.UserStore.Add(svc.User{
		Mobile:   req.Mobile,
		Password: req.Password,
		Email:    req.Email,
		Age:      req.Age,
		Gender:   req.Gender,
	})
	if err != nil {
		return nil, err
	}
	return &types.RegisterResp{Code: 0, Msg: "success"}, nil
}
