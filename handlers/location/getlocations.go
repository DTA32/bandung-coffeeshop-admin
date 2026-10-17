package location

import (
	"net/http"
	"net/url"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetLocationsHandler struct{ *handlers.Handler }

func NewGetLocationsHandler(h *handlers.Handler) *GetLocationsHandler { return &GetLocationsHandler{h} }

func (h *GetLocationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lp := utils.ParseListParams(r, templates.LocationListID)
	typ := r.URL.Query().Get("type")
	pg, err := h.Ctl.Location.List(r.Context(), typ, lp.Q, lp.Page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	next := lp.Next("/locations", url.Values{"q": {lp.Q}, "type": {typ}}, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, http.StatusOK, templates.LocationListRows(pg.Items, lp.Q, next, lp.Page == 1))
		return
	}
	h.Render(w, r, http.StatusOK, templates.LocationListPage(pg.Items, typ, lp.Q, next, r.URL.Query().Get("ok")))
}
