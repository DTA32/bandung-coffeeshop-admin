package ratingcategory

import (
	"net/url"
	"testing"
)

func TestParseRatingCategoryForm(t *testing.T) {
	if _, err := ParseRatingCategoryForm(url.Values{"name": {"x"}, "lower_bound": {"1"}}); err == nil {
		t.Fatal("missing upper bound must fail")
	}
	rc, err := ParseRatingCategoryForm(url.Values{"type": {"wifi"}, "name": {"Fast"}, "lower_bound": {"3.5"}, "upper_bound": {"5"}})
	if err != nil || rc.LowerBound != 3.5 || rc.UpperBound != 5 {
		t.Fatalf("%+v %v", rc, err)
	}
}
