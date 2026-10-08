package tag

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostTagDeleteHandler struct{ *handlers.Handler }

func NewPostTagDeleteHandler(h *handlers.Handler) *PostTagDeleteHandler {
	return &PostTagDeleteHandler{h}
}

func (h *PostTagDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, ok := utils.PathInt(r, "id")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.Ctl.Tag.Delete(r.Context(), id); err != nil && !controllers.IsNotFound(err) {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.Deleted("Tag"))
}
