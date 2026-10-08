package feed

import (
	"net/url"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

// ParseFeedForm reads the feed form; images come from post checkboxes
// ("image") and/or the free-text list ("images_text", one URL per line).
func ParseFeedForm(v url.Values) controllers.FeedInput {
	images := append([]string{}, v["image"]...)
	for _, line := range strings.Split(v.Get("images_text"), "\n") {
		if l := strings.TrimSpace(line); l != "" {
			images = append(images, l)
		}
	}
	return controllers.FeedInput{
		Content:     v.Get("content"),
		Images:      images,
		Author:      v.Get("author"),
		Link:        v.Get("link"),
		PublishedAt: v.Get("published_at"),
		ExpiresAt:   v.Get("expires_at"),
		Visible:     handlers.Checked(v, "visible"),
	}
}
