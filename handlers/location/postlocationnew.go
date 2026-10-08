package location

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostLocationNewHandler struct{ *handlers.Handler }

func NewPostLocationNewHandler(h *handlers.Handler) *PostLocationNewHandler {
	return &PostLocationNewHandler{h}
}

func (h *PostLocationNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	saveLocation(h.Handler, w, r, true)
}
