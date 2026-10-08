package post

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/feed"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type PostPostFeedHandler struct{ *handlers.Handler }

func NewPostPostFeedHandler(h *handlers.Handler) *PostPostFeedHandler { return &PostPostFeedHandler{h} }

func (h *PostPostFeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	post, ok := loadPost(h.Handler, w, r)
	if !ok {
		return
	}
	if r.ParseForm() != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	in := feed.ParseFeedForm(r.PostForm)
	_, err := h.Ctl.Post.CreateFeed(r.Context(), *post, in, h.Cfg.ImageBaseURL, h.Cfg.DisplayTZ)
	if err != nil {
		msg := handlers.UserError(err)
		if msg == "" {
			h.ServerError(w, r, err)
			return
		}
		choices := controllers.PostImageURLs(*post, h.Cfg.ImageBaseURL)
		h.Render(w, r, http.StatusUnprocessableEntity, templates.FeedFormPage(templates.FeedForm{
			Action: r.URL.Path, Title: "Create feed from post", CafeName: post.CafeName, LocationID: post.LocationID,
			Content: in.Content, Author: in.Author, Link: in.Link, PublishedAt: in.PublishedAt, ExpiresAt: in.ExpiresAt,
			Visible: in.Visible, Choices: choices, Selected: in.Images, Post: post, Error: msg,
		}))
		return
	}
	utils.RedirectOK(w, r, "/posts", "Feed entry created for "+post.CafeName+".")
}
