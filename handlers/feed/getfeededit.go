package feed

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/cafe"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetFeedEditHandler struct{ *handlers.Handler }

func NewGetFeedEditHandler(h *handlers.Handler) *GetFeedEditHandler { return &GetFeedEditHandler{h} }

func (h *GetFeedEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, name, ok := cafe.FromPath(h.Handler, w, r)
	if !ok {
		return
	}
	fid, ok := utils.PathInt(r, "fid")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	f, err := h.Ctl.Feed.Get(r.Context(), cafeID, fid)
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Feed entry")
		return
	}
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	loc := r.PathValue("id")
	pub := f.PublishedAt
	h.Render(w, r, http.StatusOK, templates.FeedFormPage(templates.FeedForm{
		Action: r.URL.Path, Title: "Edit feed entry", CafeName: name, LocationID: loc,
		Content: f.Content, Author: f.Author, Link: utils.DerefStr(f.Link),
		PublishedAt: pub.In(h.Cfg.DisplayTZ).Format("2006-01-02T15:04"),
		ExpiresAt:   utils.FormatLocal(f.ExpiresAt, h.Cfg.DisplayTZ),
		Visible:     f.Visible, Selected: f.Images, FreeImages: true,
	}))
}
