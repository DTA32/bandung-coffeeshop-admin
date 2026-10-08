package location

import (
	"net/url"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func ParseLocationForm(v url.Values) models.Location {
	return models.Location{
		ID:              strings.TrimSpace(v.Get("id")),
		Name:            v.Get("name"),
		Description:     strings.TrimSpace(v.Get("description")),
		DescriptionIndo: strings.TrimSpace(v.Get("description_indo")),
		GmapsID:         handlers.OptString(v, "gmaps_id"),
		Type:            v.Get("type"),
		Status:          v.Get("status"),
		Geojson:         handlers.OptString(v, "geojson"),
		Images:          handlers.ParseImages(v),
	}
}
