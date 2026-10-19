// Package models holds the DTOs shared by repository, controllers, handlers
// and templates, one file per entity (<entity>.model.go). Ported from
// bandung-coffeeshop-be-console@252f8ed model/admin.go and brought up to the
// current schema (weather, phone, menu).
package models

import "time"

type Coordinates struct {
	Lat float64
	Lng float64
}

type CafeListItem struct {
	ID         string // location id (slug)
	CafeID     int
	Name       string
	Status     string
	IsFeatured bool
	Instagram  *string
	UpdatedAt  string
}

type CafePrice struct {
	PriceRangeMin  *int
	PriceRangeMax  *int
	CoffeePriceMin *int
	CoffeePriceMax *int
	SnackPriceMin  *int
	SnackPriceMax  *int
	FoodPriceMin   *int
	FoodPriceMax   *int
}

type CafeTag struct {
	TagID     int
	Visible   bool
	UpdatedAt time.Time // cafe_tag.updated_at (UTC); zero for tags not saved yet
}

// Cafe is the full editable representation of a cafe across every related
// table, used for both reads and create/update writes.
type Cafe struct {
	ID              string // location id (slug)
	CafeID          int
	Name            string
	Description     string // location.description: holds the address (see 002_cafe_seeder.py)
	DescriptionIndo string
	GmapsID         *string
	Status          string
	Instagram       *string
	Phone           *string
	Menu            *string
	OpenHour        *string // HH:MM
	CloseHour       *string // HH:MM
	IsFeatured      bool
	Weather         []string
	Coordinates     *Coordinates
	About           *string // cafe.description: the "About this cafe" blurb (EN)
	AboutIndo       *string // cafe.description_indo
	Price           CafePrice
	Tags            []CafeTag
	Images          []Image
}

// CafeFormOptions is the reference data the cafe form needs.
type CafeFormOptions struct {
	Tags []Tag
}

// Weathers mirrors constants.Weather* in bandung-coffeeshop-be and the
// cafe_weather_valid CHECK in 005_cafe_weather.sql.
var Weathers = []string{"clear", "cloudy", "rain"}

// Cafe list sort keys (the ?sort= query param of GET /cafes).
const (
	CafeSortUpdatedAsc  = "updated_asc"
	CafeSortUpdatedDesc = "updated_desc"
)

// NormalizeCafeSort returns a known sort key, defaulting to least recently updated.
func NormalizeCafeSort(s string) string {
	if s == CafeSortUpdatedDesc {
		return CafeSortUpdatedDesc
	}
	return CafeSortUpdatedAsc
}
