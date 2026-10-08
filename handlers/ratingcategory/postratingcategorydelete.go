package ratingcategory

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostRatingCategoryDeleteHandler struct{ *handlers.Handler }

func NewPostRatingCategoryDeleteHandler(h *handlers.Handler) *PostRatingCategoryDeleteHandler {
	return &PostRatingCategoryDeleteHandler{h}
}

func (h *PostRatingCategoryDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.PathInt(r, "id")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.Ctl.RatingCategory.Delete(r.Context(), id); err != nil && !controllers.IsNotFound(err) {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.Deleted("Bucket"))
}
