package controllers

import (
	"context"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

type RatingCategoryController struct {
	repo *repository.RatingCategoryRepository
}

func NewRatingCategoryController(repo *repository.RatingCategoryRepository) *RatingCategoryController {
	return &RatingCategoryController{repo: repo}
}

// ListPage returns one page of rating categories for the categories page.
func (c *RatingCategoryController) ListPage(ctx context.Context, page int) (models.Page[models.RatingCategory], error) {
	return c.repo.ListPage(ctx, page)
}

func (c *RatingCategoryController) List(ctx context.Context) ([]models.RatingCategory, error) {
	return c.repo.List(ctx)
}

func (c *RatingCategoryController) Create(ctx context.Context, rc *models.RatingCategory) error {
	if err := NormalizeRatingCategory(rc); err != nil {
		return err
	}
	_, err := c.repo.Create(ctx, *rc)
	return err
}

func (c *RatingCategoryController) Update(ctx context.Context, id int, rc *models.RatingCategory) error {
	if err := NormalizeRatingCategory(rc); err != nil {
		return err
	}
	return c.repo.Update(ctx, id, *rc)
}

func (c *RatingCategoryController) Delete(ctx context.Context, id int) error {
	return c.repo.Delete(ctx, id)
}

func NormalizeRatingCategory(rc *models.RatingCategory) error {
	rc.Name = strings.TrimSpace(rc.Name)
	if rc.Name == "" {
		return invalid("name is required")
	}
	if rc.Type == "" {
		return invalid("type is required")
	}
	rc.Slug = NilIfBlank(rc.Slug)
	if rc.Slug != nil && !slugRe.MatchString(*rc.Slug) {
		return invalid("invalid slug: use lowercase letters, numbers and hyphens")
	}
	if rc.LowerBound > rc.UpperBound {
		return invalid("lower bound must not exceed upper bound")
	}
	return nil
}
