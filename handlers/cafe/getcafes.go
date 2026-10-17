package cafe

import (
	"net/http"
	"net/url"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetCafesHandler struct{ *handlers.Handler }

func NewGetCafesHandler(h *handlers.Handler) *GetCafesHandler { return &GetCafesHandler{h} }

func (h *GetCafesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lp := utils.ParseListParams(r, templates.CafeListID)
	sort := models.NormalizeCafeSort(r.URL.Query().Get("sort"))
	pg, err := h.Ctl.Cafe.List(r.Context(), lp.Q, sort, lp.Page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	next := lp.Next("/cafes", url.Values{"q": {lp.Q}, "sort": {sort}}, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, http.StatusOK, templates.CafeListRows(pg.Items, lp.Q, next, lp.Page == 1))
		return
	}
	h.Render(w, r, http.StatusOK, templates.CafeListPage(pg.Items, lp.Q, sort, next, r.URL.Query().Get("ok")))
}
