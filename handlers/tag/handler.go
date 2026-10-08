// Package tag holds the tag route handlers.
package tag

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

func renderTags(h *handlers.Handler, w http.ResponseWriter, r *http.Request, status int, errMsg string) {
	tags, err := h.Ctl.Tag.List(r.Context())
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, status, templates.TagsPage(tags, errMsg))
}
