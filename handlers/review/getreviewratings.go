package review

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetReviewRatingsHandler struct{ *handlers.Handler }

func NewGetReviewRatingsHandler(h *handlers.Handler) *GetReviewRatingsHandler {
	return &GetReviewRatingsHandler{h}
}

// ServeHTTP renders the ratings editor of any non-deleted review; the rest of
// a user review is shown read-only.
func (h *GetReviewRatingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rv, ok := loadReview(h.Handler, w, r)
	if !ok {
		return
	}
	if rv.Status == models.ReviewStatusDeleted {
		h.NotFound(w, r, "Review")
		return
	}
	opts, err := formOptions(h.Handler, r)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.ReviewRatingsPage(*rv, opts, ""))
}
