package ratingtype

import (
	"net/http"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type PostRatingTypeEditHandler struct{ *handlers.Handler }

func NewPostRatingTypeEditHandler(h *handlers.Handler) *PostRatingTypeEditHandler {
	return &PostRatingTypeEditHandler{h}
}

func (h *PostRatingTypeEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	rt := models.RatingType{Type: r.PathValue("type"), Label: r.PostForm.Get("label"), LabelIndo: strings.TrimSpace(r.PostForm.Get("label_indo"))}
	err := h.Ctl.RatingType.Update(r.Context(), &rt)
	if err != nil && handlers.UserError(err) == "" && !controllers.IsNotFound(err) {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.RatingTypeRow(rt, handlers.SavedOr(err)))
}
