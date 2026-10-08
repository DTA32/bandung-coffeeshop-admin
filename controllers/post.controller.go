package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

type PostController struct {
	repo *repository.PostRepository
}

func NewPostController(repo *repository.PostRepository) *PostController {
	return &PostController{repo: repo}
}

func DecodePost(p models.Proposal) (models.Post, error) {
	var v models.PostValue
	if err := json.Unmarshal(p.NewValue, &v); err != nil {
		return models.Post{}, fmt.Errorf("decode ig_post %d: %w", p.ID, err)
	}
	if v.Permalink == "" && p.EvidenceURL != nil {
		v.Permalink = *p.EvidenceURL
	}
	return models.Post{Proposal: p, Value: v}, nil
}

func (c *PostController) List(ctx context.Context, cafeID int) ([]models.Post, error) {
	items, err := c.repo.ListPending(ctx, cafeID)
	if err != nil {
		return nil, err
	}
	out := make([]models.Post, 0, len(items))
	for _, p := range items {
		post, err := DecodePost(p)
		if err != nil {
			return nil, err
		}
		out = append(out, post)
	}
	return out, nil
}

func (c *PostController) Get(ctx context.Context, id int64) (*models.Post, error) {
	p, err := c.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Field != models.FieldIGPost {
		return nil, ErrNotFound
	}
	post, err := DecodePost(*p)
	return &post, err
}

// Ignore rejects a pending post; the scraper never re-proposes it.
func (c *PostController) Ignore(ctx context.Context, id int64) error {
	return c.repo.Ignore(ctx, id)
}

// PostImageURLs maps the post's R2 keys to public image URLs.
func PostImageURLs(post models.Post, imageBase string) []string {
	out := make([]string, 0, len(post.Value.Media))
	for _, m := range post.Value.Media {
		if m.R2Key != "" {
			out = append(out, imageBase+"/"+strings.TrimLeft(m.R2Key, "/"))
		}
	}
	return out
}

// CreateFeed validates the form against the post's own images and inserts the
// feed entry while approving the post (one transaction in the repository).
func (c *PostController) CreateFeed(ctx context.Context, post models.Post, in FeedInput, imageBase string, tz *time.Location) (int, error) {
	allowed := PostImageURLs(post, imageBase)
	for _, im := range in.Images {
		if !slices.Contains(allowed, im) {
			return 0, invalid("image %q does not belong to this post", im)
		}
	}
	f, err := BuildFeed(in, tz, time.Now())
	if err != nil {
		return 0, err
	}
	return c.repo.CreateFeed(ctx, post.ID, f)
}
