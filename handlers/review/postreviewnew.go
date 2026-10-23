package review

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostReviewNewHandler struct{ *handlers.Handler }

func NewPostReviewNewHandler(h *handlers.Handler) *PostReviewNewHandler {
	return &PostReviewNewHandler{h}
}

func (h *PostReviewNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	savePartnership(h.Handler, w, r, 0)
}
