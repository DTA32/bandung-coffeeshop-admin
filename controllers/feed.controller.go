package controllers

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

type FeedController struct {
	repo *repository.FeedRepository
}

func NewFeedController(repo *repository.FeedRepository) *FeedController {
	return &FeedController{repo: repo}
}

func (c *FeedController) List(ctx context.Context, cafeID, page int) (models.Page[models.Feed], error) {
	return c.repo.List(ctx, cafeID, page)
}

func (c *FeedController) Get(ctx context.Context, cafeID, id int) (*models.Feed, error) {
	return c.repo.Get(ctx, cafeID, id)
}

func (c *FeedController) Create(ctx context.Context, cafeID int, in FeedInput, tz *time.Location) (int, error) {
	f, err := BuildFeed(in, tz, time.Now())
	if err != nil {
		return 0, err
	}
	f.CafeID = cafeID
	return c.repo.Create(ctx, f)
}

func (c *FeedController) Update(ctx context.Context, cafeID, id int, in FeedInput, tz *time.Location) error {
	existing, err := c.repo.Get(ctx, cafeID, id)
	if err != nil {
		return err
	}
	f, err := BuildFeed(in, tz, existing.PublishedAt)
	if err != nil {
		return err
	}
	f.ID, f.CafeID = id, cafeID
	return c.repo.Update(ctx, f)
}

func (c *FeedController) SetVisible(ctx context.Context, cafeID, id int, visible bool) error {
	return c.repo.SetVisible(ctx, cafeID, id, visible)
}

// FeedInput is the submitted feed form.
type FeedInput struct {
	Content     string
	Images      []string // full URLs
	Author      string
	Link        string
	PublishedAt string // datetime-local, in the display timezone; blank = now
	ExpiresAt   string // datetime-local; blank = never
	Visible     bool
}

const datetimeLocal = "2006-01-02T15:04"

// BuildFeed validates the form and returns the feed row (times in UTC).
func BuildFeed(in FeedInput, tz *time.Location, now time.Time) (models.Feed, error) {
	f := models.Feed{
		Content: strings.TrimSpace(in.Content),
		Author:  strings.TrimSpace(in.Author),
		Visible: in.Visible,
		Images:  []string{},
	}
	if f.Content == "" {
		return f, invalid("content is required")
	}
	if f.Author == "" {
		f.Author = "Admin"
	}
	if l := strings.TrimSpace(in.Link); l != "" {
		if !isHTTPURL(l) {
			return f, invalid("link must be an http(s) URL")
		}
		f.Link = &l
	}
	for _, im := range in.Images {
		im = strings.TrimSpace(im)
		if im == "" || slices.Contains(f.Images, im) {
			continue
		}
		if !isHTTPURL(im) {
			return f, invalid("image %q must be an http(s) URL", im)
		}
		f.Images = append(f.Images, im)
	}

	f.PublishedAt = now.UTC()
	if v := strings.TrimSpace(in.PublishedAt); v != "" {
		t, err := time.ParseInLocation(datetimeLocal, v, tz)
		if err != nil {
			return f, invalid("invalid publish time %q", v)
		}
		f.PublishedAt = t.UTC()
	}
	if v := strings.TrimSpace(in.ExpiresAt); v != "" {
		t, err := time.ParseInLocation(datetimeLocal, v, tz)
		if err != nil {
			return f, invalid("invalid expiry time %q", v)
		}
		t = t.UTC()
		if !t.After(f.PublishedAt) {
			return f, invalid("expiry must be after the publish time")
		}
		f.ExpiresAt = &t
	}
	return f, nil
}
