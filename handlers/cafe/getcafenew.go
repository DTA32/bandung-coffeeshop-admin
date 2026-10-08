package cafe

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetCafeNewHandler struct{ *handlers.Handler }

func NewGetCafeNewHandler(h *handlers.Handler) *GetCafeNewHandler { return &GetCafeNewHandler{h} }

func (h *GetCafeNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts, err := cafeOptions(h.Handler, r)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	c := &models.Cafe{Status: "active"}
	h.Render(w, r, http.StatusOK, templates.CafeFormPage(c, opts, true, "", ""))
}
