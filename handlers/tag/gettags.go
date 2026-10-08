package tag

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type GetTagsHandler struct{ *handlers.Handler }

func NewGetTagsHandler(h *handlers.Handler) *GetTagsHandler { return &GetTagsHandler{h} }

func (h *GetTagsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderTags(h.Handler, w, r, http.StatusOK, r.URL.Query().Get("err"))
}
