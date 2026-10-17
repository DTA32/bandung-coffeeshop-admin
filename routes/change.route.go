package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/change"
)

// ChangeRoutes registers the change endpoints.
func ChangeRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /changes", change.NewGetChangesHandler(h))
	mux.Handle("POST /changes/batch", change.NewPostChangeBatchHandler(h))
	mux.Handle("POST /changes/{id}/approve", change.NewPostChangeApproveHandler(h))
	mux.Handle("POST /changes/{id}/reject", change.NewPostChangeRejectHandler(h))
	mux.Handle("POST /changes/cafe/{cafeID}/approve-all", change.NewPostChangeApproveAllHandler(h))
}
