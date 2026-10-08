package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/post"
)

// PostRoutes registers the post endpoints.
func PostRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /posts", post.NewGetPostsHandler(h))
	mux.Handle("GET /posts/{id}/feed", post.NewGetPostFeedHandler(h))
	mux.Handle("POST /posts/{id}/feed", post.NewPostPostFeedHandler(h))
	mux.Handle("POST /posts/{id}/ignore", post.NewPostPostIgnoreHandler(h))
}
