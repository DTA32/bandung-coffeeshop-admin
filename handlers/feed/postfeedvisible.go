package feed

import (
	"net/http"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/cafe"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostFeedVisibleHandler struct{ *handlers.Handler }

func NewPostFeedVisibleHandler(h *handlers.Handler) *PostFeedVisibleHandler {
	return &PostFeedVisibleHandler{h}
}

func (h *PostFeedVisibleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, _, ok := cafe.FromPath(h.Handler, w, r)
	if !ok {
		return
	}
	fid, ok := utils.PathInt(r, "fid")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	visible := r.URL.Query().Get("v") == "1"
	if err := h.Ctl.Feed.SetVisible(r.Context(), cafeID, fid, visible); err != nil {
		if controllers.IsNotFound(err) {
			http.Error(w, "feed entry not found", http.StatusNotFound)
			return
		}
		h.ServerError(w, r, err)
		return
	}
	f, err := h.Ctl.Feed.Get(r.Context(), cafeID, fid)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.FeedRow(r.PathValue("id"), *f, time.Now()))
}
