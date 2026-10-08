package cafe

import (
	"net/http"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetCafesHandler struct{ *handlers.Handler }

func NewGetCafesHandler(h *handlers.Handler) *GetCafesHandler { return &GetCafesHandler{h} }

func (h *GetCafesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	sort := models.NormalizeCafeSort(r.URL.Query().Get("sort"))
	items, err := h.Ctl.Cafe.List(r.Context(), q, sort)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.CafeListPage(items, q, sort, r.URL.Query().Get("ok")))
}
