package ratingtype

import (
	"net/http"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

type PostRatingTypeNewHandler struct{ *handlers.Handler }

func NewPostRatingTypeNewHandler(h *handlers.Handler) *PostRatingTypeNewHandler {
	return &PostRatingTypeNewHandler{h}
}

func (h *PostRatingTypeNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	rt := models.RatingType{Type: r.PostForm.Get("type"), Label: r.PostForm.Get("label"), LabelIndo: strings.TrimSpace(r.PostForm.Get("label_indo"))}
	err := h.Ctl.RatingType.Create(r.Context(), &rt)
	if err != nil {
		if msg := handlers.UserError(err); msg != "" {
			renderRatingTypes(h.Handler, w, r, http.StatusUnprocessableEntity, msg)
			return
		}
		h.ServerError(w, r, err)
		return
	}
	http.Redirect(w, r, "/rating-types", http.StatusSeeOther)
}
