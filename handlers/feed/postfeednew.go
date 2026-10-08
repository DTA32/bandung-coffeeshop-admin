package feed

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/cafe"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostFeedNewHandler struct{ *handlers.Handler }

func NewPostFeedNewHandler(h *handlers.Handler) *PostFeedNewHandler { return &PostFeedNewHandler{h} }

func (h *PostFeedNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cafeID, name, ok := cafe.FromPath(h.Handler, w, r)
	if !ok {
		return
	}
	if r.ParseForm() != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	loc := r.PathValue("id")
	in := ParseFeedForm(r.PostForm)
	if _, err := h.Ctl.Feed.Create(r.Context(), cafeID, in, h.Cfg.DisplayTZ); err != nil {
		feedFormError(h.Handler, w, r, err, templates.FeedForm{
			Action: r.URL.Path, Title: "New feed entry", CafeName: name, LocationID: loc, FreeImages: true,
		}, in)
		return
	}
	utils.RedirectOK(w, r, "/cafes/"+loc+"/feed", "Feed entry created.")
}
