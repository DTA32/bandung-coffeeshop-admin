package cafe

import (
	"net/url"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestParseCafeForm(t *testing.T) {
	v := url.Values{
		"id":                   {"savor-kogi"},
		"name":                 {"Savor.KOGI"},
		"status":               {"active"},
		"instagram":            {"savor.kogi"},
		"phone":                {""},
		"open_hour":            {"08:00"},
		"close_hour":           {"00:00"},
		"is_featured":          {"1"},
		"weather":              {"clear", "rain"},
		"lat":                  {"-6.8864916"},
		"lng":                  {"107.59529"},
		"price_range_min":      {"20000"},
		"review_id":            {"12"},
		"overall_score":        {"4.5"},
		"rating_enabled_wifi":  {"1"},
		"rating_score_wifi":    {"4"},
		"rating_override_wifi": {"Fast"},
		"rating_score_noise":   {"3"}, // not enabled -> ignored
		"tag":                  {"3", "7"},
		"tag_visible_3":        {"1"},
		"image_url":            {"https://i/a.jpg", "https://i/b.jpg"},
		"image_desc":           {"A", "B"},
		"image_desc_indo":      {"A-id", "B-id"},
	}
	types := []models.RatingType{{Type: "wifi", Label: "Wifi"}, {Type: "noise", Label: "Noise"}}
	c, err := ParseCafeForm(v, types)
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "savor-kogi" || !c.IsFeatured || c.Phone != nil || *c.OpenHour != "08:00" || len(c.Weather) != 2 {
		t.Fatalf("basics: %+v", c)
	}
	if c.Coordinates == nil || c.Coordinates.Lat != -6.8864916 || c.Coordinates.Lng != 107.59529 {
		t.Fatalf("coords: %+v", c.Coordinates)
	}
	if *c.Price.PriceRangeMin != 20000 || c.Price.PriceRangeMax != nil {
		t.Fatalf("price: %+v", c.Price)
	}
	if c.Review == nil || *c.Review.ID != 12 || *c.Review.OverallScore != 4.5 {
		t.Fatalf("review: %+v", c.Review)
	}
	if len(c.Ratings) != 1 || c.Ratings[0].CategoryType != "wifi" || c.Ratings[0].Score != 4 || *c.Ratings[0].ShortDescriptionOverride != "Fast" {
		t.Fatalf("ratings: %+v", c.Ratings)
	}
	if len(c.Tags) != 2 || !c.Tags[0].Visible || c.Tags[1].Visible {
		t.Fatalf("tags: %+v", c.Tags)
	}
	if len(c.Images) != 2 || c.Images[1].URL != "https://i/b.jpg" || c.Images[1].DescriptionIndo != "B-id" {
		t.Fatalf("images: %+v", c.Images)
	}
}

func TestParseCafeFormErrors(t *testing.T) {
	types := []models.RatingType{{Type: "wifi", Label: "Wifi"}}
	cases := map[string]url.Values{
		"half coordinates": {"name": {"x"}, "lat": {"-6.9"}},
		"bad price":        {"name": {"x"}, "price_range_min": {"cheap"}},
		"rating no score":  {"name": {"x"}, "rating_enabled_wifi": {"1"}},
		"bad score":        {"name": {"x"}, "overall_score": {"great"}},
	}
	for name, v := range cases {
		t.Run(name, func(t *testing.T) {
			c, err := ParseCafeForm(v, types)
			if err == nil || handlers.UserError(err) == "" {
				t.Fatalf("expected user-facing error, got %v", err)
			}
			if c.Name != "x" {
				t.Fatal("payload must still be returned for re-rendering")
			}
		})
	}
}

func TestParseCafeFormNoReview(t *testing.T) {
	c, err := ParseCafeForm(url.Values{"name": {"x"}}, nil)
	if err != nil || c.Review != nil {
		t.Fatalf("an untouched review section must not create a review: %+v %v", c.Review, err)
	}
}
