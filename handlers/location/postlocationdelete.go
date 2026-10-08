package location

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostLocationDeleteHandler struct{ *handlers.Handler }

func NewPostLocationDeleteHandler(h *handlers.Handler) *PostLocationDeleteHandler {
	return &PostLocationDeleteHandler{h}
}

func (h *PostLocationDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Ctl.Location.Delete(r.Context(), r.PathValue("id"))
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Location")
		return
	}
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	utils.RedirectOK(w, r, "/locations", "Location deleted.")
}
