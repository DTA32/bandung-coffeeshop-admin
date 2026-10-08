package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/tag"
)

// TagRoutes registers the tag endpoints.
func TagRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /tags", tag.NewGetTagsHandler(h))
	mux.Handle("POST /tags", tag.NewPostTagNewHandler(h))
	mux.Handle("POST /tags/{id}", tag.NewPostTagEditHandler(h))
	mux.Handle("POST /tags/{id}/delete", tag.NewPostTagDeleteHandler(h))
}
