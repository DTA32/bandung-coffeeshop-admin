package ratingtype

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type PostRatingTypeDeleteHandler struct{ *handlers.Handler }

func NewPostRatingTypeDeleteHandler(h *handlers.Handler) *PostRatingTypeDeleteHandler {
	return &PostRatingTypeDeleteHandler{h}
}

func (h *PostRatingTypeDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	typ := r.PathValue("type")
	err := h.Ctl.RatingType.Delete(r.Context(), typ)
	if err != nil && !controllers.IsNotFound(err) {
		if msg := handlers.UserError(err); msg != "" {
			types, _ := h.Ctl.RatingType.List(r.Context())
			for _, rt := range types {
				if rt.Type == typ {
					h.Render(w, r, http.StatusOK, templates.RatingTypeRow(rt, msg))
					return
				}
			}
		}
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.Deleted("Dimension"))
}
