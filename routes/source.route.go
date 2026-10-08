package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/source"
)

// SourceRoutes registers the source endpoints.
func SourceRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /sources", source.NewGetSourcesHandler(h))
	mux.Handle("POST /sources", source.NewPostSourceNewHandler(h))
	mux.Handle("POST /sources/{cafeID}/{source}/external-id", source.NewPostSourceExternalIDHandler(h))
	mux.Handle("POST /sources/{cafeID}/{source}/enabled", source.NewPostSourceEnabledHandler(h))
	mux.Handle("POST /sources/{cafeID}/{source}/reset", source.NewPostSourceResetHandler(h))
	mux.Handle("GET /snapshots/{id}", source.NewGetSourceSnapshotHandler(h))
}
