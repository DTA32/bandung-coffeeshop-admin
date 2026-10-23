package user

import (
	"net/http"
	"net/url"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetUsersHandler struct{ *handlers.Handler }

func NewGetUsersHandler(h *handlers.Handler) *GetUsersHandler { return &GetUsersHandler{h} }

// ServeHTTP renders the full page, or a fragment: the next rows for the
// infinite-scroll sentinel (page >= 2), or the results (tabs + table) for the
// typeahead / tab requests that target the list container.
func (h *GetUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lp := utils.ParseListParams(r, templates.UserListID)
	pg, err := h.Ctl.User.List(r.Context(), models.UserListParams{
		Q:      lp.Q,
		Status: r.URL.Query().Get("status"),
		Page:   lp.Page,
	})
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	f := templates.UserFilter{Q: lp.Q, Status: normalizeStatus(r.URL.Query().Get("status"))}
	next := lp.Next("/users", url.Values{"q": {f.Q}, "status": {statusParam(f.Status)}}, pg.HasMore)
	switch {
	case lp.Fragment && lp.Page >= 2:
		h.Render(w, r, http.StatusOK, templates.UserRows(pg.Items, next))
	case lp.Fragment:
		h.Render(w, r, http.StatusOK, templates.UserResults(pg.Items, f, next))
	default:
		h.Render(w, r, http.StatusOK, templates.UsersPage(pg.Items, f, next))
	}
}

// normalizeStatus mirrors controllers.NormalizeUserListParams for the filter
// the page echoes back.
func normalizeStatus(s string) string {
	if s == models.UserStatusActive || s == models.UserStatusDeleted {
		return s
	}
	return models.UserFilterAll
}

// statusParam drops the default "all" from URLs.
func statusParam(s string) string {
	if s == models.UserFilterAll {
		return ""
	}
	return s
}
