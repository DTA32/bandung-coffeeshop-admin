package cafe

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostCafeNewHandler struct{ *handlers.Handler }

func NewPostCafeNewHandler(h *handlers.Handler) *PostCafeNewHandler { return &PostCafeNewHandler{h} }

func (h *PostCafeNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	saveCafe(h.Handler, w, r, true)
}
