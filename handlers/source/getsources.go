package source

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

type GetSourcesHandler struct{ *handlers.Handler }

func NewGetSourcesHandler(h *handlers.Handler) *GetSourcesHandler { return &GetSourcesHandler{h} }

func (h *GetSourcesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderSources(h.Handler, w, r, http.StatusOK, r.URL.Query().Get("ok"), "")
}
