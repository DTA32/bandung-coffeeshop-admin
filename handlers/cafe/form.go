package cafe

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func parseCoordinates(v url.Values) (*models.Coordinates, error) {
	lat, err := handlers.OptFloat(v, "lat")
	if err != nil {
		return nil, err
	}
	lng, err := handlers.OptFloat(v, "lng")
	if err != nil {
		return nil, err
	}
	if lat == nil && lng == nil {
		return nil, nil
	}
	if lat == nil || lng == nil {
		return nil, handlers.FormErrorf("%s", "set both latitude and longitude, or neither")
	}
	return &models.Coordinates{Lat: *lat, Lng: *lng}, nil
}

// ParseCafeForm maps the cafe form onto the payload. It always returns the
// best-effort payload so the form can be re-rendered with the user's input.
func ParseCafeForm(v url.Values) (models.Cafe, error) {
	c := models.Cafe{
		ID:              strings.TrimSpace(v.Get("id")),
		Name:            v.Get("name"),
		Description:     strings.TrimSpace(v.Get("description")),
		DescriptionIndo: strings.TrimSpace(v.Get("description_indo")),
		GmapsID:         handlers.OptString(v, "gmaps_id"),
		Status:          v.Get("status"),
		Instagram:       handlers.OptString(v, "instagram"),
		Phone:           handlers.OptString(v, "phone"),
		Menu:            handlers.OptString(v, "menu"),
		Hours:           handlers.ParseHours(v),
		IsFeatured:      handlers.Checked(v, "is_featured"),
		Weather:         append([]string{}, v["weather"]...),
		About:           handlers.OptString(v, "about"),
		AboutIndo:       handlers.OptString(v, "about_indo"),
		Images:          handlers.ParseImages(v),
	}
	var firstErr error
	keep := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	var err error
	c.Coordinates, err = parseCoordinates(v)
	keep(err)

	priceFields := []struct {
		key string
		dst **int
	}{
		{"price_range_min", &c.Price.PriceRangeMin}, {"price_range_max", &c.Price.PriceRangeMax},
		{"coffee_price_min", &c.Price.CoffeePriceMin}, {"coffee_price_max", &c.Price.CoffeePriceMax},
		{"snack_price_min", &c.Price.SnackPriceMin}, {"snack_price_max", &c.Price.SnackPriceMax},
		{"food_price_min", &c.Price.FoodPriceMin}, {"food_price_max", &c.Price.FoodPriceMax},
	}
	for _, pf := range priceFields {
		*pf.dst, err = handlers.OptInt(v, pf.key)
		keep(err)
	}

	for _, s := range v["tag"] {
		id, err := strconv.Atoi(s)
		if err != nil {
			keep(handlers.FormErrorf("%s", fmt.Sprintf("invalid tag id %q", s)))
			continue
		}
		c.Tags = append(c.Tags, models.CafeTag{TagID: id, Visible: handlers.Checked(v, "tag_visible_"+s)})
	}
	return c, firstErr
}
