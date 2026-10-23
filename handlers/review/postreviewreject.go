package review

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostReviewRejectHandler struct{ *handlers.Handler }

func NewPostReviewRejectHandler(h *handlers.Handler) *PostReviewRejectHandler {
	return &PostReviewRejectHandler{h}
}

func (h *PostReviewRejectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	moderate(h.Handler, w, r, h.Ctl.Review.Reject)
}
