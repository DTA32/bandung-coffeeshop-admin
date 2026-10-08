package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/feed"
)

// FeedRoutes registers the feed endpoints.
func FeedRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /cafes/{id}/feed", feed.NewGetFeedsHandler(h))
	mux.Handle("GET /cafes/{id}/feed/new", feed.NewGetFeedNewHandler(h))
	mux.Handle("POST /cafes/{id}/feed/new", feed.NewPostFeedNewHandler(h))
	mux.Handle("GET /cafes/{id}/feed/{fid}/edit", feed.NewGetFeedEditHandler(h))
	mux.Handle("POST /cafes/{id}/feed/{fid}/edit", feed.NewPostFeedEditHandler(h))
	mux.Handle("POST /cafes/{id}/feed/{fid}/visible", feed.NewPostFeedVisibleHandler(h))
}
