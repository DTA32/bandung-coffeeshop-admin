package review

import (
	"net/http"
	"net/url"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostReviewRemoveHandler struct{ *handlers.Handler }

func NewPostReviewRemoveHandler(h *handlers.Handler) *PostReviewRemoveHandler {
	return &PostReviewRemoveHandler{h}
}

// ServeHTTP soft-deletes a partnership review (status 'deleted'). From the
// queue (HTMX) it answers with the card; from the edit form it redirects.
func (h *PostReviewRemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if utils.IsHTMX(r) {
		moderate(h.Handler, w, r, h.Ctl.Review.Remove)
		return
	}
	rv, ok := loadPartnership(h.Handler, w, r)
	if !ok {
		return
	}
	err := h.Ctl.Review.Remove(r.Context(), rv.ID)
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Review")
		return
	}
	if err != nil {
		msg := handlers.UserError(err)
		if msg == "" {
			h.ServerError(w, r, err)
			return
		}
		http.Redirect(w, r, "/reviews?err="+url.QueryEscape(msg), http.StatusSeeOther)
		return
	}
	utils.RedirectOK(w, r, "/reviews?status=deleted", "Removed "+rv.ReviewerName+"'s review of "+rv.CafeName+".")
}
