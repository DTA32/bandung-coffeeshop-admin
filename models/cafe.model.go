// Package models holds the DTOs shared by repository, controllers, handlers
// and templates, one file per entity (<entity>.model.go). Ported from
// bandung-coffeeshop-be-console@252f8ed model/admin.go and brought up to the
// current schema (weather, phone, menu).
package models

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

// Review is the single editable review for a cafe. ID is set when the row
// already exists so an update targets it instead of inserting a new one.
type Review struct {
	ID            *int
	IsSubjective  bool
	OverallScore  *float64
	WFCScore      *float64
	PersonalScore *float64
	Content       *string
	ContentIndo   *string
	VisitedAt     *string // YYYY-MM-DD
}

type Rating struct {
	CategoryType                 string
	Score                        float64
	ShortDescriptionOverride     *string
	ShortDescriptionOverrideIndo *string
}

type CafeTag struct {
	TagID   int
	Visible bool
}

// Cafe is the full editable representation of a cafe across every related
// table, used for both reads and create/update writes.
type Cafe struct {
	ID              string // location id (slug)
	CafeID          int
	Name            string
	Description     string // also holds the address (see 002_cafe_seeder.py)
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
	Price           CafePrice
	Review          *Review
	Ratings         []Rating
	Tags            []CafeTag
	Images          []Image
}

// CafeFormOptions is the reference data the cafe form needs.
type CafeFormOptions struct {
	Tags             []Tag
	RatingTypes      []RatingType
	RatingCategories []RatingCategory
}

// Weathers mirrors constants.Weather* in bandung-coffeeshop-be and the
// cafe_weather_valid CHECK in 005_cafe_weather.sql.
var Weathers = []string{"clear", "cloudy", "rain"}
