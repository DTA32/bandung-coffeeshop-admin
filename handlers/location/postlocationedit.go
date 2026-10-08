package location

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostLocationEditHandler struct{ *handlers.Handler }

func NewPostLocationEditHandler(h *handlers.Handler) *PostLocationEditHandler {
	return &PostLocationEditHandler{h}
}

func (h *PostLocationEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	saveLocation(h.Handler, w, r, false)
}
