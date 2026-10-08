package location

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetLocationNewHandler struct{ *handlers.Handler }

func NewGetLocationNewHandler(h *handlers.Handler) *GetLocationNewHandler {
	return &GetLocationNewHandler{h}
}

func (h *GetLocationNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := &models.Location{Type: "poi", Status: "active"}
	h.Render(w, r, http.StatusOK, templates.LocationFormPage(l, true, "", ""))
}
