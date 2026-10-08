package post

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

type GetPostFeedHandler struct{ *handlers.Handler }

func NewGetPostFeedHandler(h *handlers.Handler) *GetPostFeedHandler { return &GetPostFeedHandler{h} }

func (h *GetPostFeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	post, ok := loadPost(h.Handler, w, r)
	if !ok {
		return
	}
	if post.Status != models.StatusPending {
		utils.RedirectOK(w, r, "/posts", "That post was already handled.")
		return
	}
	choices := controllers.PostImageURLs(*post, h.Cfg.ImageBaseURL)
	now := time.Now()
	f := templates.FeedForm{
		Action:      "/posts/" + strconv.FormatInt(post.ID, 10) + "/feed",
		Title:       "Create feed from post",
		CafeName:    post.CafeName,
		LocationID:  post.LocationID,
		Content:     strings.TrimSpace(post.Value.Caption),
		Author:      "Admin",
		Link:        post.Value.Permalink,
		PublishedAt: now.In(h.Cfg.DisplayTZ).Format("2006-01-02T15:04"),
		Visible:     true,
		Choices:     choices,
		Selected:    choices,
		Post:        post,
	}
	h.Render(w, r, http.StatusOK, templates.FeedFormPage(f))
}
