package cafe

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetCafeHandler struct{ *handlers.Handler }

func NewGetCafeHandler(h *handlers.Handler) *GetCafeHandler { return &GetCafeHandler{h} }

func (h *GetCafeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := h.Ctl.Cafe.Get(r.Context(), r.PathValue("id"))
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Cafe")
		return
	}
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	opts, err := cafeOptions(h.Handler, r)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.CafeFormPage(c, opts, false, "", r.URL.Query().Get("ok")))
}
