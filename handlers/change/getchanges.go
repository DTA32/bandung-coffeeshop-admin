package change

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetChangesHandler struct{ *handlers.Handler }

func NewGetChangesHandler(h *handlers.Handler) *GetChangesHandler { return &GetChangesHandler{h} }

func (h *GetChangesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID := utils.QueryInt(r, "cafe")
	groups, err := h.Ctl.Change.List(r.Context(), cafeID)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	q := r.URL.Query()
	h.Render(w, r, http.StatusOK, templates.ChangesPage(groups, cafeID, q.Get("ok"), q.Get("err")))
}
