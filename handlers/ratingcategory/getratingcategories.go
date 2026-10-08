package ratingcategory

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type GetRatingCategoriesHandler struct{ *handlers.Handler }

func NewGetRatingCategoriesHandler(h *handlers.Handler) *GetRatingCategoriesHandler {
	return &GetRatingCategoriesHandler{h}
}

func (h *GetRatingCategoriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderRatings(h.Handler, w, r, http.StatusOK, "")
}
