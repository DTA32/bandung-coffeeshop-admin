package review

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostReviewEditHandler struct{ *handlers.Handler }

func NewPostReviewEditHandler(h *handlers.Handler) *PostReviewEditHandler {
	return &PostReviewEditHandler{h}
}

func (h *PostReviewEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rv, ok := loadPartnership(h.Handler, w, r)
	if !ok {
		return
	}
	savePartnership(h.Handler, w, r, rv.ID)
}
