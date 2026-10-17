// Package ratingtype holds the ratingtype route handlers.
package ratingtype

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

func renderRatingTypes(h *handlers.Handler, w http.ResponseWriter, r *http.Request, status int, errMsg string) {
	lp := utils.ParseListParams(r, "")
	pg, err := h.Ctl.RatingType.ListPage(r.Context(), lp.Page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	next := lp.Next("/rating-types", nil, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, status, templates.RatingTypeRows(pg.Items, next))
		return
	}
	h.Render(w, r, status, templates.RatingTypesPage(pg.Items, next, errMsg))
}
