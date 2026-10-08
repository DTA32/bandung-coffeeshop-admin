package source

import (
	"net/http"
	"strconv"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostSourceNewHandler struct{ *handlers.Handler }

func NewPostSourceNewHandler(h *handlers.Handler) *PostSourceNewHandler {
	return &PostSourceNewHandler{h}
}

func (h *PostSourceNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	cafeID, err := strconv.Atoi(r.PostForm.Get("cafe_id"))
	if err != nil {
		renderSources(h.Handler, w, r, http.StatusUnprocessableEntity, "", "Pick a cafe.")
		return
	}
	err = h.Ctl.Source.SetExternalID(r.Context(), cafeID, r.PostForm.Get("source"), r.PostForm.Get("external_id"))
	if err != nil {
		if msg := handlers.UserError(err); msg != "" {
			renderSources(h.Handler, w, r, http.StatusUnprocessableEntity, "", msg)
			return
		}
		h.ServerError(w, r, err)
		return
	}
	utils.RedirectOK(w, r, "/sources", "Target saved.")
}
