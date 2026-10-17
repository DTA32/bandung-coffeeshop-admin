package controllers

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

type LocationController struct {
	repo *repository.LocationRepository
}

func NewLocationController(repo *repository.LocationRepository) *LocationController {
	return &LocationController{repo: repo}
}

func (c *LocationController) List(ctx context.Context, typ, q string, page int) (models.Page[models.LocationListItem], error) {
	return c.repo.List(ctx, typ, q, page)
}

func (c *LocationController) Get(ctx context.Context, id string) (*models.Location, error) {
	return c.repo.Get(ctx, id)
}

func (c *LocationController) Create(ctx context.Context, p *models.Location) error {
	if !slugRe.MatchString(p.ID) {
		return invalid("invalid id: use lowercase letters, numbers and hyphens")
	}
	if err := NormalizeLocation(p); err != nil {
		return err
	}
	return c.repo.Create(ctx, *p)
}

func (c *LocationController) Update(ctx context.Context, id string, p *models.Location) error {
	p.ID = id
	if err := NormalizeLocation(p); err != nil {
		return err
	}
	return c.repo.Update(ctx, id, *p)
}

// Delete soft-deletes the location.
func (c *LocationController) Delete(ctx context.Context, id string) error {
	return c.repo.SoftDelete(ctx, id)
}

func NormalizeLocation(p *models.Location) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return invalid("name is required")
	}
	if !slices.Contains(models.LocationTypes, p.Type) {
		return invalid("invalid type %q", p.Type)
	}
	if p.Status == "" {
		p.Status = "active"
	}
	if !slices.Contains(models.LocationStatuses, p.Status) {
		return invalid("invalid status %q", p.Status)
	}
	p.GmapsID = NilIfBlank(p.GmapsID)
	p.Geojson = NilIfBlank(p.Geojson)
	if p.Geojson != nil {
		var g struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(*p.Geojson), &g); err != nil || g.Type == "" {
			return invalid("geojson must be a GeoJSON geometry object")
		}
		want := map[string][]string{
			"poi":      {"Point"},
			"area":     {"Polygon", "MultiPolygon"},
			"district": {"Polygon", "MultiPolygon"},
		}[p.Type]
		if !slices.Contains(want, g.Type) {
			return invalid("%s needs a %s geometry, got %s", p.Type, strings.Join(want, " or "), g.Type)
		}
	}
	return NormalizeImages(&p.Images)
}

// NormalizeImages drops empty rows and rejects duplicate URLs, which would
// otherwise silently collapse into one row via the (location_id, url) upsert.
func NormalizeImages(images *[]models.Image) error {
	out := make([]models.Image, 0, len(*images))
	seen := make(map[string]struct{}, len(*images))
	for _, im := range *images {
		im.URL = strings.TrimSpace(im.URL)
		if im.URL == "" {
			continue
		}
		if _, ok := seen[im.URL]; ok {
			return invalid("duplicate image url %s", im.URL)
		}
		seen[im.URL] = struct{}{}
		out = append(out, im)
	}
	*images = out
	return nil
}
