package controllers

import (
	"context"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

// DashboardController composes the home page from several repositories.
type DashboardController struct {
	repo    *repository.DashboardRepository
	sources *repository.SourceRepository
}

func NewDashboardController(repo *repository.DashboardRepository, sources *repository.SourceRepository) *DashboardController {
	return &DashboardController{repo: repo, sources: sources}
}

func (c *DashboardController) Get(ctx context.Context) (*models.Dashboard, error) {
	d, err := c.repo.Counts(ctx)
	if err != nil {
		return nil, err
	}
	d.LastRuns, err = c.sources.LastRuns(ctx)
	return d, err
}
