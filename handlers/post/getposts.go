package post

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetPostsHandler struct{ *handlers.Handler }

func NewGetPostsHandler(h *handlers.Handler) *GetPostsHandler { return &GetPostsHandler{h} }

func (h *GetPostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lp := utils.ParseListParams(r, "")
	cafeID := utils.QueryInt(r, "cafe")
	pg, err := h.Ctl.Post.List(r.Context(), cafeID, lp.Page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	cafeParam := ""
	if cafeID != 0 {
		cafeParam = strconv.Itoa(cafeID)
	}
	next := lp.Next("/posts", url.Values{"cafe": {cafeParam}}, pg.HasMore)
	if lp.Fragment {
		h.Render(w, r, http.StatusOK, templates.PostCards(pg.Items, next))
		return
	}
	h.Render(w, r, http.StatusOK, templates.PostsPage(pg.Items, cafeID, next, r.URL.Query().Get("ok")))
}
