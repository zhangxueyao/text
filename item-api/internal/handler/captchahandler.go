package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zhangxueyao/item/item-api/internal/logic"
	"github.com/zhangxueyao/item/item-api/internal/svc"
)

func CaptchaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewCaptchaLogic(r.Context(), svcCtx)
		resp, err := l.Captcha()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
