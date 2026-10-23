package controllers

import (
	"context"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

type RatingTypeController struct {
	repo *repository.RatingTypeRepository
}

func NewRatingTypeController(repo *repository.RatingTypeRepository) *RatingTypeController {
	return &RatingTypeController{repo: repo}
}

// ListPage returns one page of rating types for the rating types page.
func (c *RatingTypeController) ListPage(ctx context.Context, page int) (models.Page[models.RatingType], error) {
	return c.repo.ListPage(ctx, page)
}

func (c *RatingTypeController) List(ctx context.Context) ([]models.RatingType, error) {
	return c.repo.List(ctx)
}

func (c *RatingTypeController) Create(ctx context.Context, rt *models.RatingType) error {
	if err := NormalizeRatingType(rt); err != nil {
		return err
	}
	return c.repo.Create(ctx, *rt)
}

// Update edits the labels of rt.Type (the key is immutable).
func (c *RatingTypeController) Update(ctx context.Context, rt *models.RatingType) error {
	if err := NormalizeRatingType(rt); err != nil {
		return err
	}
	return c.repo.Update(ctx, rt.Type, *rt)
}

// Delete refuses while any category or cafe rating references the type.
func (c *RatingTypeController) Delete(ctx context.Context, typeKey string) error {
	return c.repo.Delete(ctx, typeKey)
}

func NormalizeRatingType(rt *models.RatingType) error {
	rt.Type = strings.TrimSpace(rt.Type)
	rt.Label = strings.TrimSpace(rt.Label)
	if !slugRe.MatchString(rt.Type) {
		return invalid("invalid type key: use lowercase letters, numbers and hyphens")
	}
	if rt.Label == "" {
		return invalid("label is required")
	}
	return nil
}

// ValidateRatings enforces each dimension's own range (e.g. price-rank spans
// 0..100000). A type without buckets has no range, so it can't be rated (the
// BE rejects it for user reviews too).
func ValidateRatings(ratings []models.Rating, bounds map[string]models.RatingBounds) error {
	for _, rt := range ratings {
		b, ok := bounds[rt.CategoryType]
		if !ok {
			return invalid("invalid rating type %q", rt.CategoryType)
		}
		if b.Min == nil || b.Max == nil {
			return invalid("%s has no rating categories yet", rt.CategoryType)
		}
		if rt.Score < *b.Min || rt.Score > *b.Max {
			return invalid("%s score %g is outside %g–%g", rt.CategoryType, rt.Score, *b.Min, *b.Max)
		}
	}
	return nil
}
