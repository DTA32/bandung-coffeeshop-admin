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
