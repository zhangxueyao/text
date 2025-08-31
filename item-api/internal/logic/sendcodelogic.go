package logic

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zhangxueyao/item/item-api/internal/svc"
	"github.com/zhangxueyao/item/item-api/internal/types"
)

type SendCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCodeLogic {
	return &SendCodeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *SendCodeLogic) SendCode(req *types.SendCodeReq) (*types.SendCodeResp, error) {
	captcha, ok := l.svcCtx.CaptchaStore.Get(req.CaptchaId)
	if !ok || captcha != req.Captcha {
		return nil, errors.New("invalid captcha")
	}
	l.svcCtx.CaptchaStore.Delete(req.CaptchaId)

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	l.svcCtx.CodeStore.Set(req.Mobile, code, 5*time.Minute)
	logx.Infof("send sms code %s to %s", code, req.Mobile)
	return &types.SendCodeResp{Code: 0, Msg: "ok"}, nil
}
