package controllers

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

type CafeController struct {
	repo *repository.CafeRepository
}

func NewCafeController(repo *repository.CafeRepository) *CafeController {
	return &CafeController{repo: repo}
}

// TouchTag bumps one saved cafe tag's updated_at; returns the new timestamp.
func (c *CafeController) TouchTag(ctx context.Context, locationID string, tagID int) (time.Time, error) {
	return c.repo.TouchTag(ctx, locationID, tagID)
}

func (c *CafeController) List(ctx context.Context, q, sort string, page int) (models.Page[models.CafeListItem], error) {
	return c.repo.List(ctx, q, sort, page)
}

// Options lists every cafe for dropdowns (no paging).
func (c *CafeController) Options(ctx context.Context) ([]models.CafeListItem, error) {
	return c.repo.Options(ctx)
}

func (c *CafeController) Get(ctx context.Context, id string) (*models.Cafe, error) {
	return c.repo.Get(ctx, id)
}

// IDByLocation resolves a cafe's numeric id and name from its location id.
func (c *CafeController) IDByLocation(ctx context.Context, locationID string) (int, string, error) {
	return c.repo.IDByLocation(ctx, locationID)
}

func (c *CafeController) Create(ctx context.Context, p *models.Cafe) error {
	if !slugRe.MatchString(p.ID) {
		return invalid("invalid id: use lowercase letters, numbers and hyphens")
	}
	if err := NormalizeCafe(p); err != nil {
		return err
	}
	return c.repo.Create(ctx, *p)
}

func (c *CafeController) Update(ctx context.Context, id string, p *models.Cafe) error {
	p.ID = id
	if err := NormalizeCafe(p); err != nil {
		return err
	}
	return c.repo.Update(ctx, id, *p)
}

// Delete soft-deletes the cafe's location.
func (c *CafeController) Delete(ctx context.Context, id string) error {
	return c.repo.SoftDelete(ctx, id)
}

// NormalizeCafe trims, defaults and validates a cafe payload in place
// (everything that does not need the database).
func NormalizeCafe(p *models.Cafe) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return invalid("name is required")
	}
	if p.Status == "" {
		p.Status = "active"
	}
	if !slices.Contains(models.LocationStatuses, p.Status) {
		return invalid("invalid status %q", p.Status)
	}

	p.GmapsID = NilIfBlank(p.GmapsID)
	p.Instagram = NilIfBlank(p.Instagram)
	if p.Instagram != nil {
		h := strings.TrimPrefix(strings.TrimSpace(*p.Instagram), "@")
		p.Instagram = &h
	}
	p.Phone = NilIfBlank(p.Phone)
	p.Menu = NilIfBlank(p.Menu)
	if p.Menu != nil && !isMenuValue(*p.Menu) {
		return invalid("menu must be an http(s) URL or #image-N (N >= 1)")
	}
	p.OpenHour = NilIfBlank(p.OpenHour)
	p.CloseHour = NilIfBlank(p.CloseHour)
	if err := validateHour(p.OpenHour); err != nil {
		return err
	}
	if err := validateHour(p.CloseHour); err != nil {
		return err
	}

	if p.Weather == nil {
		p.Weather = []string{}
	}
	for _, w := range p.Weather {
		if !slices.Contains(models.Weathers, w) {
			return invalid("invalid weather %q", w)
		}
	}

	if c := p.Coordinates; c != nil && (c.Lat < -90 || c.Lat > 90 || c.Lng < -180 || c.Lng > 180) {
		return invalid("invalid coordinates")
	}

	p.About = NilIfBlank(p.About)
	p.AboutIndo = NilIfBlank(p.AboutIndo)

	seen := make(map[int]struct{}, len(p.Tags))
	for _, tg := range p.Tags {
		if _, ok := seen[tg.TagID]; ok {
			return invalid("duplicate tag in payload")
		}
		seen[tg.TagID] = struct{}{}
	}
	return NormalizeImages(&p.Images)
}
