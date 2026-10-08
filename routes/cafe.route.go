package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/cafe"
)

// CafeRoutes registers the cafe endpoints.
func CafeRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /cafes", cafe.NewGetCafesHandler(h))
	mux.Handle("GET /cafes/new", cafe.NewGetCafeNewHandler(h))
	mux.Handle("POST /cafes/new", cafe.NewPostCafeNewHandler(h))
	mux.Handle("GET /cafes/{id}", cafe.NewGetCafeHandler(h))
	mux.Handle("POST /cafes/{id}", cafe.NewPostCafeEditHandler(h))
	mux.Handle("POST /cafes/{id}/delete", cafe.NewPostCafeDeleteHandler(h))
	mux.Handle("POST /cafes/tags", cafe.NewPostCafeTagNewHandler(h))
}
