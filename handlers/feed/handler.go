// Package feed holds the feed route handlers.
package feed

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

func feedFormError(h *handlers.Handler, w http.ResponseWriter, r *http.Request, err error, f templates.FeedForm, in controllers.FeedInput) {
	msg := handlers.UserError(err)
	if msg == "" {
		h.ServerError(w, r, err)
		return
	}
	f.Content, f.Author, f.Link = in.Content, in.Author, in.Link
	f.PublishedAt, f.ExpiresAt, f.Visible, f.Selected = in.PublishedAt, in.ExpiresAt, in.Visible, in.Images
	f.Error = msg
	h.Render(w, r, http.StatusUnprocessableEntity, templates.FeedFormPage(f))
}
