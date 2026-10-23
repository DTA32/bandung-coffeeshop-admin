package review

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostReviewApproveHandler struct{ *handlers.Handler }

func NewPostReviewApproveHandler(h *handlers.Handler) *PostReviewApproveHandler {
	return &PostReviewApproveHandler{h}
}

func (h *PostReviewApproveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	moderate(h.Handler, w, r, h.Ctl.Review.Approve)
}
