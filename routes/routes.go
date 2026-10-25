// Package routes registers every endpoint, one <entity>.route.go per entity.
package routes

import (
	"io/fs"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/middleware"
)

// New wires every endpoint and the middleware chain. static is the embedded
// static/ tree.
func New(h *handlers.Handler, static fs.FS) http.Handler {
	mux := http.NewServeMux()

	BasicRoutes(mux, h, static)
	DashboardRoutes(mux, h)
	CafeRoutes(mux, h)
	LocationRoutes(mux, h)
	TagRoutes(mux, h)
	RatingTypeRoutes(mux, h)
	RatingCategoryRoutes(mux, h)
	FeedRoutes(mux, h)
	ChangeRoutes(mux, h)
	PostRoutes(mux, h)
	SourceRoutes(mux, h)
	UserRoutes(mux, h)
	ReviewRoutes(mux, h)
	ImageRoutes(mux, h)

	var root http.Handler = mux
	root = middleware.SameOrigin(root)
	root = middleware.SecurityHeaders(root)
	root = middleware.Logging(root)
	root = middleware.Recover(root)
	return root
}
