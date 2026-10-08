package models

type Image struct {
	URL             string
	Description     string
	DescriptionIndo string
}

type LocationListItem struct {
	ID     string
	Name   string
	Type   string
	Status string
}

// Location is a non-cafe location (poi / area / district) plus its images.
// Coordinates are exchanged as a GeoJSON geometry string.
type Location struct {
	ID              string
	Name            string
	Description     string
	DescriptionIndo string
	GmapsID         *string
	Type            string
	Status          string
	Geojson         *string
	Images          []Image
}

var (
	LocationStatuses = []string{"active", "closed", "deleted"}
	LocationTypes    = []string{"poi", "area", "district"}
)
