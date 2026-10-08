// Package dashboard holds the dashboard route handlers.
package dashboard

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetDashboardHandler struct{ *handlers.Handler }

func NewGetDashboardHandler(h *handlers.Handler) *GetDashboardHandler { return &GetDashboardHandler{h} }

func (h *GetDashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d, err := h.Ctl.Dashboard.Get(r.Context())
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.DashboardPage(d))
}
