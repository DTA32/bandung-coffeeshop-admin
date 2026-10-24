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
		Hours:     []models.CafeHour{{Day: ptr("tue"), Open: ptr(" 08:00 "), Close: ptr("22:00")}, {Closed: true, Day: ptr("mon")}},
		About:     ptr("  "),
		Images:    []models.Image{{URL: " https://a/1.jpg "}, {URL: ""}},
	}
	if err := NormalizeCafe(&c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "Savor" || c.Status != "active" || *c.Instagram != "savor.kogi" || c.Phone != nil {
		t.Fatalf("unexpected normalisation: %+v", c)
	}
	if len(c.Hours) != 2 || *c.Hours[0].Day != "mon" || !c.Hours[0].Closed || *c.Hours[1].Open != "08:00" {
		t.Fatalf("hours must be trimmed and sorted mon→sun: %+v", c.Hours)
	}
	if c.Weather == nil || len(c.Weather) != 0 {
		t.Fatalf("weather must default to an empty, non-nil slice (column is NOT NULL)")
	}
	if c.About != nil {
		t.Fatalf("blank about must become NULL: %q", *c.About)
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
		{"bad hour", func(c *models.Cafe) { c.Hours = []models.CafeHour{{Open: ptr("24:00"), Close: ptr("02:00")}} }, "invalid hour"},
		{"bad weather", func(c *models.Cafe) { c.Weather = []string{"snow"} }, "invalid weather"},
		{"bad menu", func(c *models.Cafe) { c.Menu = ptr("drive.google.com/x") }, "menu must be"},
		{"bad menu gallery", func(c *models.Cafe) { c.Menu = ptr("#image-0") }, "menu must be"},
		{"bad coords", func(c *models.Cafe) { c.Coordinates = &models.Coordinates{Lat: 100} }, "invalid coordinates"},
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
