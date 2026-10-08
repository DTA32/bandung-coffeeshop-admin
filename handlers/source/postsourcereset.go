package source

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostSourceResetHandler struct{ *handlers.Handler }

func NewPostSourceResetHandler(h *handlers.Handler) *PostSourceResetHandler {
	return &PostSourceResetHandler{h}
}

func (h *PostSourceResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, source, ok := targetFromPath(h.Handler, w, r)
	if !ok {
		return
	}
	if err := h.Ctl.Source.ResetFailures(r.Context(), cafeID, source); err != nil {
		h.ServerError(w, r, err)
		return
	}
	renderTarget(h.Handler, w, r, cafeID, source, "")
}
