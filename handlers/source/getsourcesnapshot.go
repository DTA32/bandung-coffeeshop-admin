package source

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetSourceSnapshotHandler struct{ *handlers.Handler }

func NewGetSourceSnapshotHandler(h *handlers.Handler) *GetSourceSnapshotHandler {
	return &GetSourceSnapshotHandler{h}
}

func (h *GetSourceSnapshotHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.PathInt64(r, "id")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	s, err := h.Ctl.Source.Snapshot(r.Context(), id)
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Snapshot")
		return
	}
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.SnapshotPage(s))
}
