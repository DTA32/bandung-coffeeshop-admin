package source

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostSourceEnabledHandler struct{ *handlers.Handler }

func NewPostSourceEnabledHandler(h *handlers.Handler) *PostSourceEnabledHandler {
	return &PostSourceEnabledHandler{h}
}

func (h *PostSourceEnabledHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, source, ok := targetFromPath(h.Handler, w, r)
	if !ok {
		return
	}
	if err := h.Ctl.Source.SetEnabled(r.Context(), cafeID, source, r.URL.Query().Get("v") == "1"); err != nil {
		h.ServerError(w, r, err)
		return
	}
	renderTarget(h.Handler, w, r, cafeID, source, "")
}
