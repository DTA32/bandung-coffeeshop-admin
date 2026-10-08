// Package ratingtype holds the ratingtype route handlers.
package ratingtype

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

func renderRatingTypes(h *handlers.Handler, w http.ResponseWriter, r *http.Request, status int, errMsg string) {
	types, err := h.Ctl.RatingType.List(r.Context())
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, status, templates.RatingTypesPage(types, errMsg))
}
