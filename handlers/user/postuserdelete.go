package user

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostUserDeleteHandler struct{ *handlers.Handler }

func NewPostUserDeleteHandler(h *handlers.Handler) *PostUserDeleteHandler {
	return &PostUserDeleteHandler{h}
}

func (h *PostUserDeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setStatus(h.Handler, w, r, h.Ctl.User.Delete)
}
