package feed

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/cafe"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostFeedEditHandler struct{ *handlers.Handler }

func NewPostFeedEditHandler(h *handlers.Handler) *PostFeedEditHandler { return &PostFeedEditHandler{h} }

func (h *PostFeedEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, name, ok := cafe.FromPath(h.Handler, w, r)
	if !ok {
		return
	}
	fid, ok := utils.PathInt(r, "fid")
	if !ok || r.ParseForm() != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	loc := r.PathValue("id")
	in := ParseFeedForm(r.PostForm)
	if err := h.Ctl.Feed.Update(r.Context(), cafeID, fid, in, h.Cfg.DisplayTZ); err != nil {
		if controllers.IsNotFound(err) {
			h.NotFound(w, r, "Feed entry")
			return
		}
		feedFormError(h.Handler, w, r, err, templates.FeedForm{
			Action: r.URL.Path, Title: "Edit feed entry", CafeName: name, LocationID: loc, FreeImages: true,
		}, in)
		return
	}
	utils.RedirectOK(w, r, "/cafes/"+loc+"/feed", "Feed entry saved.")
}
