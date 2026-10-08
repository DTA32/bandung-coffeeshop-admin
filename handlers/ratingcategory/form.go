package ratingcategory

import (
	"net/url"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func ParseRatingCategoryForm(v url.Values) (models.RatingCategory, error) {
	rc := models.RatingCategory{
		Type:                 v.Get("type"),
		Name:                 v.Get("name"),
		NameIndo:             strings.TrimSpace(v.Get("name_indo")),
		Slug:                 handlers.OptString(v, "slug"),
		ShortDescription:     strings.TrimSpace(v.Get("short_description")),
		ShortDescriptionIndo: strings.TrimSpace(v.Get("short_description_indo")),
		LongDescription:      strings.TrimSpace(v.Get("long_description")),
		LongDescriptionIndo:  strings.TrimSpace(v.Get("long_description_indo")),
	}
	lb, err := handlers.OptFloat(v, "lower_bound")
	if err != nil {
		return rc, err
	}
	ub, err := handlers.OptFloat(v, "upper_bound")
	if err != nil {
		return rc, err
	}
	if lb == nil || ub == nil {
		return rc, handlers.FormErrorf("%s", "lower and upper bounds are required")
	}
	rc.LowerBound, rc.UpperBound = *lb, *ub
	return rc, nil
}
