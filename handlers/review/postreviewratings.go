package review

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostReviewRatingsHandler struct{ *handlers.Handler }

func NewPostReviewRatingsHandler(h *handlers.Handler) *PostReviewRatingsHandler {
	return &PostReviewRatingsHandler{h}
}

// ServeHTTP replaces the review's ratings (status and updated_at untouched)
// and returns to the review's queue tab.
func (h *PostReviewRatingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rv, ok := loadReview(h.Handler, w, r)
	if !ok {
		return
	}
	if r.ParseForm() != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	opts, err := formOptions(h.Handler, r)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	ratings, err := ParseRatings(r.PostForm, opts.RatingTypes)
	if err == nil {
		err = h.Ctl.Review.SaveRatings(r.Context(), rv.ID, ratings)
	}
	if err != nil {
		if controllers.IsNotFound(err) {
			h.NotFound(w, r, "Review")
			return
		}
		msg := handlers.UserError(err)
		if msg == "" {
			h.ServerError(w, r, err)
			return
		}
		rv.Ratings = ratings
		h.Render(w, r, http.StatusUnprocessableEntity, templates.ReviewRatingsPage(*rv, opts, msg))
		return
	}
	utils.RedirectOK(w, r, queueURL(rv), "Saved ratings of "+rv.ReviewerName+"'s review of "+rv.CafeName+".")
}
