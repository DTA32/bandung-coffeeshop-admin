package controllers

import (
	"context"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

type TagController struct {
	repo *repository.TagRepository
}

func NewTagController(repo *repository.TagRepository) *TagController {
	return &TagController{repo: repo}
}

func (c *TagController) List(ctx context.Context) ([]models.Tag, error) {
	return c.repo.List(ctx)
}

// ListPage returns one page of tags for the tags page.
func (c *TagController) ListPage(ctx context.Context, page int) (models.Page[models.Tag], error) {
	return c.repo.ListPage(ctx, page)
}

// Create normalises t, inserts it and sets t.ID.
func (c *TagController) Create(ctx context.Context, t *models.Tag) error {
	if err := NormalizeTag(t); err != nil {
		return err
	}
	id, err := c.repo.Create(ctx, *t)
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

func (c *TagController) Update(ctx context.Context, id int, t *models.Tag) error {
	if err := NormalizeTag(t); err != nil {
		return err
	}
	return c.repo.Update(ctx, id, *t)
}

// Delete removes the tag and its cafe assignments.
func (c *TagController) Delete(ctx context.Context, id int) error {
	return c.repo.Delete(ctx, id)
}

func NormalizeTag(t *models.Tag) error {
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" {
		return invalid("tag name is required")
	}
	t.Slug = NilIfBlank(t.Slug)
	if t.Slug != nil && !slugRe.MatchString(*t.Slug) {
		return invalid("invalid slug: use lowercase letters, numbers and hyphens")
	}
	return nil
}
