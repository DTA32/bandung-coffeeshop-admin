package source

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostSourceExternalIDHandler struct{ *handlers.Handler }

func NewPostSourceExternalIDHandler(h *handlers.Handler) *PostSourceExternalIDHandler {
	return &PostSourceExternalIDHandler{h}
}

func (h *PostSourceExternalIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, source, ok := targetFromPath(h.Handler, w, r)
	if !ok {
		return
	}
	if r.ParseForm() != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	err := h.Ctl.Source.SetExternalID(r.Context(), cafeID, source, r.PostForm.Get("external_id"))
	if err != nil && handlers.UserError(err) == "" {
		h.ServerError(w, r, err)
		return
	}
	renderTarget(h.Handler, w, r, cafeID, source, handlers.SavedOr(err))
}
