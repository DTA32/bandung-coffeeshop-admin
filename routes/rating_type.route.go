package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/ratingtype"
)

// RatingTypeRoutes registers the rating type endpoints.
func RatingTypeRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /rating-types", ratingtype.NewGetRatingTypesHandler(h))
	mux.Handle("POST /rating-types", ratingtype.NewPostRatingTypeNewHandler(h))
	mux.Handle("POST /rating-types/{type}", ratingtype.NewPostRatingTypeEditHandler(h))
	mux.Handle("POST /rating-types/{type}/delete", ratingtype.NewPostRatingTypeDeleteHandler(h))
}
