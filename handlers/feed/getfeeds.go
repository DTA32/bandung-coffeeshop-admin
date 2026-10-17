package feed

import (
	"net/http"
	"net/url"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/cafe"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetFeedsHandler struct{ *handlers.Handler }

func NewGetFeedsHandler(h *handlers.Handler) *GetFeedsHandler { return &GetFeedsHandler{h} }

func (h *GetFeedsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, name, ok := cafe.FromPath(h.Handler, w, r)
	if !ok {
		return
	}
	lp := utils.ParseListParams(r, "")
	pg, err := h.Ctl.Feed.List(r.Context(), cafeID, lp.Page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	locationID := r.PathValue("id")
	next := lp.Next("/cafes/"+url.PathEscape(locationID)+"/feed", nil, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, http.StatusOK, templates.FeedRows(locationID, pg.Items, time.Now(), next))
		return
	}
	h.Render(w, r, http.StatusOK, templates.FeedListPage(locationID, name, pg.Items, time.Now(), next, r.URL.Query().Get("ok")))
}
