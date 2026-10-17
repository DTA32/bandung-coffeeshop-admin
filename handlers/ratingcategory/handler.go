// Package ratingcategory holds the ratingcategory route handlers.
package ratingcategory

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

func renderRatings(h *handlers.Handler, w http.ResponseWriter, r *http.Request, status int, errMsg string) {
	types, err := h.Ctl.RatingType.List(r.Context())
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	lp := utils.ParseListParams(r, "")
	pg, err := h.Ctl.RatingCategory.ListPage(r.Context(), lp.Page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	next := lp.Next("/rating-categories", nil, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, status, templates.RatingCategoryRows(pg.Items, types, next))
		return
	}
	h.Render(w, r, status, templates.RatingCategoriesPage(types, pg.Items, next, errMsg))
}
