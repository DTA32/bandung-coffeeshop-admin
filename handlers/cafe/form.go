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
func ParseCafeForm(v url.Values, ratingTypes []models.RatingType) (models.Cafe, error) {
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
		OpenHour:        handlers.OptString(v, "open_hour"),
		CloseHour:       handlers.OptString(v, "close_hour"),
		IsFeatured:      handlers.Checked(v, "is_featured"),
		Weather:         append([]string{}, v["weather"]...),
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

	// Review: kept when it already exists or any review field is filled.
	rv := models.Review{
		IsSubjective: handlers.Checked(v, "is_subjective"),
		Content:      handlers.OptString(v, "review_content"),
		ContentIndo:  handlers.OptString(v, "review_content_indo"),
		VisitedAt:    handlers.OptString(v, "visited_at"),
	}
	rv.ID, err = handlers.OptInt(v, "review_id")
	keep(err)
	rv.OverallScore, err = handlers.OptFloat(v, "overall_score")
	keep(err)
	rv.WFCScore, err = handlers.OptFloat(v, "wfc_score")
	keep(err)
	rv.PersonalScore, err = handlers.OptFloat(v, "personal_score")
	keep(err)
	if rv.ID != nil || rv.IsSubjective || rv.Content != nil || rv.ContentIndo != nil || rv.VisitedAt != nil ||
		rv.OverallScore != nil || rv.WFCScore != nil || rv.PersonalScore != nil {
		c.Review = &rv
	}

	for _, rt := range ratingTypes {
		if !handlers.Checked(v, "rating_enabled_"+rt.Type) {
			continue
		}
		score, err := handlers.OptFloat(v, "rating_score_"+rt.Type)
		if err == nil && score == nil {
			err = handlers.FormErrorf("%s", fmt.Sprintf("%s is enabled but has no score", rt.Label))
		}
		keep(err)
		r := models.Rating{
			CategoryType:                 rt.Type,
			ShortDescriptionOverride:     handlers.OptString(v, "rating_override_"+rt.Type),
			ShortDescriptionOverrideIndo: handlers.OptString(v, "rating_override_indo_"+rt.Type),
		}
		if score != nil {
			r.Score = *score
		}
		c.Ratings = append(c.Ratings, r)
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
