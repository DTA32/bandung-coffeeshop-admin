package location

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetLocationHandler struct{ *handlers.Handler }

func NewGetLocationHandler(h *handlers.Handler) *GetLocationHandler { return &GetLocationHandler{h} }

func (h *GetLocationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l, err := h.Ctl.Location.Get(r.Context(), r.PathValue("id"))
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Location")
		return
	}
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.LocationFormPage(l, false, "", r.URL.Query().Get("ok")))
}
