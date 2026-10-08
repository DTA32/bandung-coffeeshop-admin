package models

type RatingType struct {
	Type      string
	Label     string
	LabelIndo string
}

// RatingBounds is a dimension's valid score range, derived from its buckets.
// Min/Max are nil for a type without buckets.
type RatingBounds struct {
	Min, Max *float64
}
