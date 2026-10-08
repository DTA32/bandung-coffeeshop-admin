package change

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostChangeRejectHandler struct{ *handlers.Handler }

func NewPostChangeRejectHandler(h *handlers.Handler) *PostChangeRejectHandler {
	return &PostChangeRejectHandler{h}
}

func (h *PostChangeRejectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.PathInt64(r, "id")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	p, err := h.Ctl.Change.Get(r.Context(), id)
	if err == nil {
		err = h.Ctl.Change.Reject(r.Context(), id)
	}
	renderProposalResult(h.Handler, w, r, p, err, models.StatusRejected)
}
