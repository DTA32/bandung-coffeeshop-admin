package post

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetPostsHandler struct{ *handlers.Handler }

func NewGetPostsHandler(h *handlers.Handler) *GetPostsHandler { return &GetPostsHandler{h} }

func (h *GetPostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID := utils.QueryInt(r, "cafe")
	posts, err := h.Ctl.Post.List(r.Context(), cafeID)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.PostsPage(posts, cafeID, r.URL.Query().Get("ok")))
}
