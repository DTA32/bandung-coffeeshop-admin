package cafe

import (
	"net/url"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

func TestParseCafeForm(t *testing.T) {
	v := url.Values{
		"id":              {"savor-kogi"},
		"name":            {"Savor.KOGI"},
		"status":          {"active"},
		"instagram":       {"savor.kogi"},
		"phone":           {""},
		"hours_group":     {"all", "mon"},
		"hours_open":      {"08:00", ""},
		"hours_close":     {"00:00", ""},
		"hours_closed":    {"mon"},
		"is_featured":     {"1"},
		"weather":         {"clear", "rain"},
		"lat":             {"-6.8864916"},
		"lng":             {"107.59529"},
		"price_range_min": {"20000"},
		"about":           {"  Cozy spot  "},
		"about_indo":      {""},
		"tag":             {"3", "7"},
		"tag_visible_3":   {"1"},
		"image_url":       {"https://i/a.jpg", "https://i/b.jpg"},
		"image_desc":      {"A", "B"},
		"image_desc_indo": {"A-id", "B-id"},
	}
	c, err := ParseCafeForm(v)
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "savor-kogi" || !c.IsFeatured || c.Phone != nil || len(c.Hours) != 2 || *c.Hours[0].Open != "08:00" || !c.Hours[1].Closed || len(c.Weather) != 2 {
		t.Fatalf("basics: %+v", c)
	}
	if c.Coordinates == nil || c.Coordinates.Lat != -6.8864916 || c.Coordinates.Lng != 107.59529 {
		t.Fatalf("coords: %+v", c.Coordinates)
	}
	if *c.Price.PriceRangeMin != 20000 || c.Price.PriceRangeMax != nil {
		t.Fatalf("price: %+v", c.Price)
	}
	if c.About == nil || *c.About != "Cozy spot" || c.AboutIndo != nil {
		t.Fatalf("about: %v %v", c.About, c.AboutIndo)
	}
	if len(c.Tags) != 2 || !c.Tags[0].Visible || c.Tags[1].Visible {
		t.Fatalf("tags: %+v", c.Tags)
	}
	if len(c.Images) != 2 || c.Images[1].URL != "https://i/b.jpg" || c.Images[1].DescriptionIndo != "B-id" {
		t.Fatalf("images: %+v", c.Images)
	}
}

func TestParseCafeFormErrors(t *testing.T) {
	cases := map[string]url.Values{
		"half coordinates": {"name": {"x"}, "lat": {"-6.9"}},
		"bad price":        {"name": {"x"}, "price_range_min": {"cheap"}},
	}
	for name, v := range cases {
		t.Run(name, func(t *testing.T) {
			c, err := ParseCafeForm(v)
			if err == nil || handlers.UserError(err) == "" {
				t.Fatalf("expected user-facing error, got %v", err)
			}
			if c.Name != "x" {
				t.Fatal("payload must still be returned for re-rendering")
			}
		})
	}
}
