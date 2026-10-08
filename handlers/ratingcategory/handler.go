// Package ratingcategory holds the ratingcategory route handlers.
package ratingcategory

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

func renderRatings(h *handlers.Handler, w http.ResponseWriter, r *http.Request, status int, errMsg string) {
	types, err := h.Ctl.RatingType.List(r.Context())
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	cats, err := h.Ctl.RatingCategory.List(r.Context())
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, status, templates.RatingCategoriesPage(types, cats, errMsg))
}
