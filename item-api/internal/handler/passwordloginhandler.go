package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zhangxueyao/item/item-api/internal/logic"
	"github.com/zhangxueyao/item/item-api/internal/svc"
	"github.com/zhangxueyao/item/item-api/internal/types"
)

func PasswordLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PasswordLoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewPasswordLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
