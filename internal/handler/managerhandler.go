package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"manager/internal/logic"
	"manager/internal/svc"
	"manager/internal/types"
)

func ManagerHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Request
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewManagerLogic(r.Context(), svcCtx)
		resp, err := l.Manager(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
