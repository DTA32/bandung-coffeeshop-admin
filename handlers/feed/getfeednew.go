package feed

import (
	"net/http"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/cafe"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
)

type GetFeedNewHandler struct{ *handlers.Handler }

func NewGetFeedNewHandler(h *handlers.Handler) *GetFeedNewHandler { return &GetFeedNewHandler{h} }

func (h *GetFeedNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, name, ok := cafe.FromPath(h.Handler, w, r)
	if !ok {
		return
	}
	loc := r.PathValue("id")
	h.Render(w, r, http.StatusOK, templates.FeedFormPage(templates.FeedForm{
		Action: "/cafes/" + loc + "/feed/new", Title: "New feed entry", CafeName: name, LocationID: loc,
		Author: "Admin", PublishedAt: time.Now().In(h.Cfg.DisplayTZ).Format("2006-01-02T15:04"),
		Visible: true, FreeImages: true,
	}))
}
