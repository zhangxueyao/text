package handler

import (
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zhangxueyao/item-rpc/internal/logic"
	"github.com/zhangxueyao/item-rpc/internal/svc"
)

func GetItemHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(rest.RouteContext(r.Context()).Vars()["id"], 10, 64)
		logic := logic.NewGetItemLogic(r.Context(), svcCtx)
		resp, err := logic.GetItem(id)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
