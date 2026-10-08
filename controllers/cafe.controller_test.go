package controllers

import (
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func ptr[T any](v T) *T { return &v }

func TestNormalizeCafe(t *testing.T) {
	c := models.Cafe{
		Name:      "  Savor  ",
		Instagram: ptr(" @savor.kogi "),
		Phone:     ptr("  "),
		OpenHour:  ptr("08:00"),
		CloseHour: ptr(""),
		Images:    []models.Image{{URL: " https://a/1.jpg "}, {URL: ""}},
	}
	if err := NormalizeCafe(&c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "Savor" || c.Status != "active" || *c.Instagram != "savor.kogi" || c.Phone != nil || c.CloseHour != nil {
		t.Fatalf("unexpected normalisation: %+v", c)
	}
	if c.Weather == nil || len(c.Weather) != 0 {
		t.Fatalf("weather must default to an empty, non-nil slice (column is NOT NULL)")
	}
	if len(c.Images) != 1 || c.Images[0].URL != "https://a/1.jpg" {
		t.Fatalf("images: %+v", c.Images)
	}

	bad := []struct {
		name string
		mut  func(*models.Cafe)
		want string
	}{
		{"no name", func(c *models.Cafe) { c.Name = " " }, "name is required"},
		{"bad status", func(c *models.Cafe) { c.Status = "gone" }, "invalid status"},
		{"bad hour", func(c *models.Cafe) { c.OpenHour = ptr("24:00") }, "invalid hour"},
		{"bad weather", func(c *models.Cafe) { c.Weather = []string{"snow"} }, "invalid weather"},
		{"bad menu", func(c *models.Cafe) { c.Menu = ptr("drive.google.com/x") }, "menu must be"},
		{"bad coords", func(c *models.Cafe) { c.Coordinates = &models.Coordinates{Lat: 100} }, "invalid coordinates"},
		{"bad score", func(c *models.Cafe) { c.Review = &models.Review{OverallScore: ptr(5.5)} }, "between 0 and 5"},
		{"dup tag", func(c *models.Cafe) { c.Tags = []models.CafeTag{{TagID: 1}, {TagID: 1}} }, "duplicate tag"},
		{"dup image", func(c *models.Cafe) { c.Images = []models.Image{{URL: "https://x"}, {URL: "https://x"}} }, "duplicate image"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			c := models.Cafe{Name: "ok"}
			tc.mut(&c)
			err := NormalizeCafe(&c)
			if err == nil || !strings.Contains(err.Error(), tc.want) || !IsValidation(err) {
				t.Fatalf("got %v, want validation error containing %q", err, tc.want)
			}
		})
	}
}

func TestValidateRatings(t *testing.T) {
	bounds := map[string]models.RatingBounds{
		"wifi":       {Min: ptr(0.0), Max: ptr(5.0)},
		"price-rank": {Min: ptr(0.0), Max: ptr(100000.0)},
		"new":        {},
	}
	ok := []models.Rating{{CategoryType: "wifi", Score: 5}, {CategoryType: "price-rank", Score: 45000}, {CategoryType: "new", Score: 999}}
	if err := ValidateRatings(ok, bounds); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRatings([]models.Rating{{CategoryType: "wifi", Score: 6}}, bounds); err == nil {
		t.Fatal("expected out of bounds")
	}
	if err := ValidateRatings([]models.Rating{{CategoryType: "nope", Score: 1}}, bounds); err == nil {
		t.Fatal("expected unknown type")
	}
}
