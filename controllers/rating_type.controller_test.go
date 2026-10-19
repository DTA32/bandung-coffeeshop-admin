package controllers

import (
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestValidateRatings(t *testing.T) {
	bounds := map[string]models.RatingBounds{
		"wifi":       {Min: ptr(0.0), Max: ptr(5.0)},
		"price-rank": {Min: ptr(0.0), Max: ptr(100000.0)},
		"new":        {},
	}
	ok := []models.Rating{{CategoryType: "wifi", Score: 5}, {CategoryType: "price-rank", Score: 45000}}
	if err := ValidateRatings(ok, bounds); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRatings([]models.Rating{{CategoryType: "wifi", Score: 6}}, bounds); err == nil {
		t.Fatal("expected out of bounds")
	}
	if err := ValidateRatings([]models.Rating{{CategoryType: "nope", Score: 1}}, bounds); err == nil {
		t.Fatal("expected unknown type")
	}
	if err := ValidateRatings([]models.Rating{{CategoryType: "new", Score: 1}}, bounds); err == nil {
		t.Fatal("expected type without buckets")
	}
}
