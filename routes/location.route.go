package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/location"
)

// LocationRoutes registers the location endpoints.
func LocationRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /locations", location.NewGetLocationsHandler(h))
	mux.Handle("GET /locations/new", location.NewGetLocationNewHandler(h))
	mux.Handle("POST /locations/new", location.NewPostLocationNewHandler(h))
	mux.Handle("GET /locations/{id}", location.NewGetLocationHandler(h))
	mux.Handle("POST /locations/{id}", location.NewPostLocationEditHandler(h))
	mux.Handle("POST /locations/{id}/delete", location.NewPostLocationDeleteHandler(h))
}
