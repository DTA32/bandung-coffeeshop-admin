package change

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostChangeApproveHandler struct{ *handlers.Handler }

func NewPostChangeApproveHandler(h *handlers.Handler) *PostChangeApproveHandler {
	return &PostChangeApproveHandler{h}
}

func (h *PostChangeApproveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.PathInt64(r, "id")
	if !ok || r.ParseForm() != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var edit *controllers.Edit
	if r.PostForm.Get("edit") != "" {
		edit = &controllers.Edit{Value: r.PostForm.Get("value"), Hours: handlers.ParseHours(r.PostForm)}
	}
	p, err := h.Ctl.Change.Approve(r.Context(), id, edit)
	renderProposalResult(h.Handler, w, r, p, err, models.StatusApproved)
}
