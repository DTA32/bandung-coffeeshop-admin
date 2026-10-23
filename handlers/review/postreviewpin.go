package review

import (
	"context"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostReviewPinHandler struct{ *handlers.Handler }

func NewPostReviewPinHandler(h *handlers.Handler) *PostReviewPinHandler {
	return &PostReviewPinHandler{h}
}

func (h *PostReviewPinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	moderate(h.Handler, w, r, func(ctx context.Context, id int) error {
		return h.Ctl.Review.SetPinned(ctx, id, true)
	})
}
