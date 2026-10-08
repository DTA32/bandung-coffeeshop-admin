package cafe

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostCafeEditHandler struct{ *handlers.Handler }

func NewPostCafeEditHandler(h *handlers.Handler) *PostCafeEditHandler { return &PostCafeEditHandler{h} }

func (h *PostCafeEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	saveCafe(h.Handler, w, r, false)
}
