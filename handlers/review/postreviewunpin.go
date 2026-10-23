package review

import (
	"context"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostReviewUnpinHandler struct{ *handlers.Handler }

func NewPostReviewUnpinHandler(h *handlers.Handler) *PostReviewUnpinHandler {
	return &PostReviewUnpinHandler{h}
}

func (h *PostReviewUnpinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	moderate(h.Handler, w, r, func(ctx context.Context, id int) error {
		return h.Ctl.Review.SetPinned(ctx, id, false)
	})
}
