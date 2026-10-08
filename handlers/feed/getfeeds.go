package feed

import (
	"net/http"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/cafe"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetFeedsHandler struct{ *handlers.Handler }

func NewGetFeedsHandler(h *handlers.Handler) *GetFeedsHandler { return &GetFeedsHandler{h} }

func (h *GetFeedsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, name, ok := cafe.FromPath(h.Handler, w, r)
	if !ok {
		return
	}
	items, err := h.Ctl.Feed.List(r.Context(), cafeID)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.FeedListPage(r.PathValue("id"), name, items, time.Now(), r.URL.Query().Get("ok")))
}
