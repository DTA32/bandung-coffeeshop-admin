// Package review holds the review route handlers (BCF-31): the /reviews
// moderation queue, per-review ratings and partnership reviews.
package review

import (
	"context"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

// loadReview resolves the {id} path value, writing the 400/404/500 response
// itself and returning false when it cannot.
func loadReview(h *handlers.Handler, w http.ResponseWriter, r *http.Request) (*models.Review, bool) {
	id, ok := utils.PathInt(r, "id")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return nil, false
	}
	rv, err := h.Ctl.Review.Get(r.Context(), id)
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Review")
		return nil, false
	}
	if err != nil {
		h.ServerError(w, r, err)
		return nil, false
	}
	return rv, true
}

// formOptions loads the cafes and rating dimensions / buckets the review forms
// need.
func formOptions(h *handlers.Handler, r *http.Request) (models.ReviewFormOptions, error) {
	var opts models.ReviewFormOptions
	var err error
	if opts.Cafes, err = h.Ctl.Cafe.Options(r.Context()); err != nil {
		return opts, err
	}
	if opts.RatingTypes, err = h.Ctl.RatingType.List(r.Context()); err != nil {
		return opts, err
	}
	opts.RatingCategories, err = h.Ctl.RatingCategory.List(r.Context())
	return opts, err
}

// moderate runs a queue action on the {id} review and answers with its card
// for the HTMX outerHTML swap. The form value tab is the status tab the card
// is shown on: a review that left it renders as a compact resolved card.
// Validation-type failures (e.g. the status changed meanwhile) re-render the
// card with the message.
func moderate(h *handlers.Handler, w http.ResponseWriter, r *http.Request, apply func(context.Context, int) error) {
	id, ok := utils.PathInt(r, "id")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	err := apply(r.Context(), id)
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Review")
		return
	}
	msg := ""
	if err != nil {
		if msg = handlers.UserError(err); msg == "" {
			h.ServerError(w, r, err)
			return
		}
	}
	rv, err := h.Ctl.Review.Get(r.Context(), id)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	tab := controllers.NormalizeReviewListParams(models.ReviewListParams{Status: r.FormValue("tab")}).Status
	h.Render(w, r, http.StatusOK, templates.ReviewRow(*rv, tab, msg))
}

// queueURL is the /reviews tab a review is listed on.
func queueURL(rv *models.Review) string {
	return "/reviews?status=" + rv.Status
}
