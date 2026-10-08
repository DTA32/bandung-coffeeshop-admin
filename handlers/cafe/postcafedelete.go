package cafe

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostCafeDeleteHandler struct{ *handlers.Handler }

func NewPostCafeDeleteHandler(h *handlers.Handler) *PostCafeDeleteHandler {
	return &PostCafeDeleteHandler{h}
}

func (h *PostCafeDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Ctl.Cafe.Delete(r.Context(), r.PathValue("id"))
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Cafe")
		return
	}
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	utils.RedirectOK(w, r, "/cafes", "Cafe deleted.")
}
