package ratingcategory

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostRatingCategoryEditHandler struct{ *handlers.Handler }

func NewPostRatingCategoryEditHandler(h *handlers.Handler) *PostRatingCategoryEditHandler {
	return &PostRatingCategoryEditHandler{h}
}

func (h *PostRatingCategoryEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.PathInt(r, "id")
	if !ok || r.ParseForm() != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	rc, err := ParseRatingCategoryForm(r.PostForm)
	rc.ID = id
	if err == nil {
		err = handlers.FriendlyUnique(h.Ctl.RatingCategory.Update(r.Context(), id, &rc), "slug for this dimension")
	}
	if err != nil && handlers.UserError(err) == "" {
		h.ServerError(w, r, err)
		return
	}
	types, terr := h.Ctl.RatingType.List(r.Context())
	if terr != nil {
		h.ServerError(w, r, terr)
		return
	}
	h.Render(w, r, http.StatusOK, templates.RatingCategoryRow(rc, types, handlers.SavedOr(err)))
}
