package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/dashboard"
)

// DashboardRoutes registers the dashboard endpoints.
func DashboardRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /{$}", dashboard.NewGetDashboardHandler(h))
}
