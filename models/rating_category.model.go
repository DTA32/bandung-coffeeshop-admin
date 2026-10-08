package models

type RatingCategory struct {
	ID                   int
	Type                 string
	Name                 string
	NameIndo             string
	Slug                 *string
	ShortDescription     string
	ShortDescriptionIndo string
	LongDescription      string
	LongDescriptionIndo  string
	LowerBound           float64
	UpperBound           float64
}
