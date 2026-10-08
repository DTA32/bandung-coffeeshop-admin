package location

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetLocationsHandler struct{ *handlers.Handler }

func NewGetLocationsHandler(h *handlers.Handler) *GetLocationsHandler { return &GetLocationsHandler{h} }

func (h *GetLocationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	typ := r.URL.Query().Get("type")
	items, err := h.Ctl.Location.List(r.Context(), typ)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.LocationListPage(items, typ, r.URL.Query().Get("ok")))
}
