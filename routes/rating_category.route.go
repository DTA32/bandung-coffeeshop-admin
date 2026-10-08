package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/ratingcategory"
)

// RatingCategoryRoutes registers the rating category endpoints.
func RatingCategoryRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /rating-categories", ratingcategory.NewGetRatingCategoriesHandler(h))
	mux.Handle("POST /rating-categories", ratingcategory.NewPostRatingCategoryNewHandler(h))
	mux.Handle("POST /rating-categories/{id}", ratingcategory.NewPostRatingCategoryEditHandler(h))
	mux.Handle("POST /rating-categories/{id}/delete", ratingcategory.NewPostRatingCategoryDeleteHandler(h))
}
