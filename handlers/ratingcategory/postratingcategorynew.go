package ratingcategory

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostRatingCategoryNewHandler struct{ *handlers.Handler }

func NewPostRatingCategoryNewHandler(h *handlers.Handler) *PostRatingCategoryNewHandler {
	return &PostRatingCategoryNewHandler{h}
}

func (h *PostRatingCategoryNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	rc, err := ParseRatingCategoryForm(r.PostForm)
	if err == nil {
		err = handlers.FriendlyUnique(h.Ctl.RatingCategory.Create(r.Context(), &rc), "slug for this dimension")
	}
	if err != nil {
		if msg := handlers.UserError(err); msg != "" {
			renderRatings(h.Handler, w, r, http.StatusUnprocessableEntity, msg)
			return
		}
		h.ServerError(w, r, err)
		return
	}
	http.Redirect(w, r, "/rating-categories", http.StatusSeeOther)
}
