package routes

import (
	"io/fs"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/middleware"
)

// BasicRoutes registers the endpoints that belong to no entity: health and static files.
func BasicRoutes(mux *http.ServeMux, h *handlers.Handler, static fs.FS) {
	mux.Handle("GET /health", handlers.NewGetHealthHandler(h))
	mux.Handle("GET /static/", http.StripPrefix("/static/", middleware.CacheStatic(http.FileServerFS(static))))
}
