package controllers

import (
	"context"
	"regexp"
	"slices"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

var (
	gmapsFeatureRe = regexp.MustCompile(`^0x[0-9a-f]+:0x[0-9a-f]+$`)
	digitsRe       = regexp.MustCompile(`^[0-9]+$`)
)

type SourceController struct {
	repo *repository.SourceRepository
}

func NewSourceController(repo *repository.SourceRepository) *SourceController {
	return &SourceController{repo: repo}
}

func (c *SourceController) ListTargets(ctx context.Context, f models.TargetFilter) ([]models.SourceTarget, error) {
	return c.repo.ListTargets(ctx, f)
}

func (c *SourceController) RecentRuns(ctx context.Context, limit int) ([]models.ScrapeRun, error) {
	return c.repo.RecentRuns(ctx, limit)
}

func (c *SourceController) Snapshot(ctx context.Context, id int64) (*models.Snapshot, error) {
	return c.repo.GetSnapshot(ctx, id)
}

func (c *SourceController) SetEnabled(ctx context.Context, cafeID int, source string, enabled bool) error {
	return c.repo.SetTargetEnabled(ctx, cafeID, source, enabled)
}

func (c *SourceController) ResetFailures(ctx context.Context, cafeID int, source string) error {
	return c.repo.ResetTargetFailures(ctx, cafeID, source)
}

func (c *SourceController) SetExternalID(ctx context.Context, cafeID int, source, externalID string) error {
	if !slices.Contains(models.Sources, source) {
		return invalid("unknown source %q", source)
	}
	v := NilIfBlank(&externalID)
	if v != nil {
		switch source {
		case "linktree":
			if !isHTTPURL(*v) {
				return invalid("linktree target must be an http(s) URL")
			}
		case "gmaps":
			if !gmapsFeatureRe.MatchString(*v) {
				return invalid("gmaps target must be a feature id like 0x2e68…:0x9d13…")
			}
		case "instagram":
			if !digitsRe.MatchString(*v) {
				return invalid("instagram target is the numeric user id (the handle comes from the cafe form)")
			}
		}
	}
	return c.repo.UpsertTarget(ctx, cafeID, source, v)
}
