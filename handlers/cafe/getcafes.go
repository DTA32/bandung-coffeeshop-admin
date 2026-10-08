package cafe

import (
	"net/http"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetCafesHandler struct{ *handlers.Handler }

func NewGetCafesHandler(h *handlers.Handler) *GetCafesHandler { return &GetCafesHandler{h} }

func (h *GetCafesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	items, err := h.Ctl.Cafe.List(r.Context(), q)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.CafeListPage(items, q, r.URL.Query().Get("ok")))
}
