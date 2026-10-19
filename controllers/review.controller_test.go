package controllers

import (
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestNormalizePartnershipReview(t *testing.T) {
	rv := models.Review{
		CafeID:           3,
		ReviewerName:     "  Kopi Club  ",
		Content:          ptr(" Great beans "),
		OverallScore:     ptr(4.5),
		VisitedAt:        ptr("2026-09"),
		ExternalURLTitle: ptr(" "),
		ExternalURLLink:  ptr("https://example.com/post"),
	}
	if err := NormalizePartnershipReview(&rv); err != nil {
		t.Fatal(err)
	}
	if rv.ReviewerName != "Kopi Club" || *rv.Content != "Great beans" || *rv.VisitedAt != "2026-09-01" || rv.ExternalURLTitle != nil {
		t.Fatalf("unexpected normalisation: %+v", rv)
	}

	bad := []struct {
		name string
		mut  func(*models.Review)
		want string
	}{
		{"no cafe", func(r *models.Review) { r.CafeID = 0 }, "cafe is required"},
		{"no name", func(r *models.Review) { r.ReviewerName = " " }, "reviewer name"},
		{"no content", func(r *models.Review) { r.Content = ptr("") }, "content is required"},
		{"score low", func(r *models.Review) { r.WFCScore = ptr(0.5) }, "between 1 and 5"},
		{"score step", func(r *models.Review) { r.OverallScore = ptr(4.25) }, "one decimal"},
		{"bad month", func(r *models.Review) { r.VisitedAt = ptr("09/2026") }, "visit month"},
		{"bad link", func(r *models.Review) { r.ExternalURLLink = ptr("javascript:alert(1)") }, "http(s)"},
		{"title no link", func(r *models.Review) { r.ExternalURLTitle = ptr("IG") }, "needs a link"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			r := models.Review{CafeID: 1, ReviewerName: "x", Content: ptr("y")}
			tc.mut(&r)
			err := NormalizePartnershipReview(&r)
			if err == nil || !strings.Contains(err.Error(), tc.want) || !IsValidation(err) {
				t.Fatalf("got %v, want validation error containing %q", err, tc.want)
			}
		})
	}
}

func TestNormalizeReviewListParams(t *testing.T) {
	p := NormalizeReviewListParams(models.ReviewListParams{Status: "nope", CafeID: -1, AfterID: -5})
	if p.Status != models.ReviewStatusPending || p.CafeID != 0 || p.AfterID != 0 {
		t.Fatalf("%+v", p)
	}
}
