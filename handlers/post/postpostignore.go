package post

import (
	"errors"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type PostPostIgnoreHandler struct{ *handlers.Handler }

func NewPostPostIgnoreHandler(h *handlers.Handler) *PostPostIgnoreHandler {
	return &PostPostIgnoreHandler{h}
}

func (h *PostPostIgnoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	post, ok := loadPost(h.Handler, w, r)
	if !ok {
		return
	}
	if err := h.Ctl.Post.Ignore(r.Context(), post.ID); err != nil && !errors.Is(err, controllers.ErrProposalNotPending) {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.PostIgnored(*post))
}
