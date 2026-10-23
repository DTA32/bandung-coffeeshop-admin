package user

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type PostUserReviveHandler struct{ *handlers.Handler }

func NewPostUserReviveHandler(h *handlers.Handler) *PostUserReviveHandler {
	return &PostUserReviveHandler{h}
}

func (h *PostUserReviveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setStatus(h.Handler, w, r, h.Ctl.User.Revive)
}
