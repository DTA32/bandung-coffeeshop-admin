package user

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetUsersHandler struct{ *handlers.Handler }

func NewGetUsersHandler(h *handlers.Handler) *GetUsersHandler { return &GetUsersHandler{h} }

// ServeHTTP renders the full page, or only the results (tabs, table,
// pagination) for the HTMX search / tab / page requests. History restores
// (HX-History-Restore-Request) need the full page.
func (h *GetUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page, err := h.Ctl.User.List(r.Context(), models.UserListParams{
		Q:      r.URL.Query().Get("q"),
		Status: r.URL.Query().Get("status"),
		Page:   utils.QueryInt(r, "page"),
	})
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	if utils.IsHTMX(r) && r.Header.Get("HX-History-Restore-Request") != "true" {
		h.Render(w, r, http.StatusOK, templates.UserResults(page))
		return
	}
	h.Render(w, r, http.StatusOK, templates.UsersPage(page))
}
