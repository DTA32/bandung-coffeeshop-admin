package change

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostChangeApproveAllHandler struct{ *handlers.Handler }

func NewPostChangeApproveAllHandler(h *handlers.Handler) *PostChangeApproveAllHandler {
	return &PostChangeApproveAllHandler{h}
}

func (h *PostChangeApproveAllHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, ok := utils.PathInt(r, "cafeID")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	n, err := h.Ctl.Change.ApproveAllForCafe(r.Context(), cafeID)
	if err != nil {
		msg := handlers.UserError(err)
		if msg == "" {
			h.ServerError(w, r, err)
			return
		}
		http.Redirect(w, r, "/changes?err="+url.QueryEscape(fmt.Sprintf("Approved %d, then stopped: %s", n, msg)), http.StatusSeeOther)
		return
	}
	utils.RedirectOK(w, r, "/changes", fmt.Sprintf("Approved %d change(s).", n))
}
