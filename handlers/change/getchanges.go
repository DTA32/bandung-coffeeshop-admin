package change

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetChangesHandler struct{ *handlers.Handler }

func NewGetChangesHandler(h *handlers.Handler) *GetChangesHandler { return &GetChangesHandler{h} }

func (h *GetChangesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lp := utils.ParseListParams(r, "")
	cafeID := utils.QueryInt(r, "cafe")
	pg, err := h.Ctl.Change.List(r.Context(), cafeID, utils.QueryInt(r, "after"))
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	cafeParam := ""
	if cafeID != 0 {
		cafeParam = strconv.Itoa(cafeID)
	}
	after := ""
	if n := len(pg.Items); n > 0 {
		after = strconv.Itoa(pg.Items[n-1].CafeID)
	}
	next := lp.Next("/changes", url.Values{"cafe": {cafeParam}, "after": {after}}, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, http.StatusOK, templates.ChangeGroups(pg.Items, next))
		return
	}
	q := r.URL.Query()
	h.Render(w, r, http.StatusOK, templates.ChangesPage(pg.Items, cafeID, next, q.Get("ok"), q.Get("err")))
}
