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
	f := templates.FeedForm{
		Action:      "/posts/" + strconv.FormatInt(post.ID, 10) + "/feed",
		Title:       "Create feed from post",
		CafeName:    post.CafeName,
		LocationID:  post.LocationID,
		Content:     strings.TrimSpace(post.Value.Caption),
		Author:      "Admin",
		Link:        post.Value.Permalink,
		PublishedAt: defaultPublishedAt(post.Value.TakenAt, time.Now()).In(h.Cfg.DisplayTZ).Format("2006-01-02T15:04"),
		Visible:     true,
		Choices:     choices,
		Selected:    choices,
		Post:        post,
	}
	h.Render(w, r, http.StatusOK, templates.FeedFormPage(f))
}

// defaultPublishedAt prefills the feed's published_at with when the post went
// up on Instagram, falling back to now when the scraper didn't record it.
func defaultPublishedAt(takenAt, now time.Time) time.Time {
	if takenAt.IsZero() {
		return now
	}
	return takenAt
}
