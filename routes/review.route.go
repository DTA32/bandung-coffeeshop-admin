package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/review"
)

// ReviewRoutes registers the review endpoints.
func ReviewRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /reviews", review.NewGetReviewsHandler(h))
	mux.Handle("POST /reviews/{id}/approve", review.NewPostReviewApproveHandler(h))
	mux.Handle("POST /reviews/{id}/reject", review.NewPostReviewRejectHandler(h))
	mux.Handle("POST /reviews/{id}/pin", review.NewPostReviewPinHandler(h))
	mux.Handle("POST /reviews/{id}/unpin", review.NewPostReviewUnpinHandler(h))
	mux.Handle("GET /reviews/{id}/ratings", review.NewGetReviewRatingsHandler(h))
	mux.Handle("POST /reviews/{id}/ratings", review.NewPostReviewRatingsHandler(h))
}
