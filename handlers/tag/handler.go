// Package tag holds the tag route handlers.
package tag

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

func renderTags(h *handlers.Handler, w http.ResponseWriter, r *http.Request, status int, errMsg string) {
	lp := utils.ParseListParams(r, "")
	pg, err := h.Ctl.Tag.ListPage(r.Context(), lp.Page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	next := lp.Next("/tags", nil, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, status, templates.TagRows(pg.Items, next))
		return
	}
	h.Render(w, r, status, templates.TagsPage(pg.Items, next, errMsg))
}
