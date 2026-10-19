package review

import (
	"net/url"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestParseRatings(t *testing.T) {
	types := []models.RatingType{{Type: "wifi", Label: "Wifi"}, {Type: "noise", Label: "Noise"}}
	v := url.Values{
		"rating_enabled_wifi":  {"1"},
		"rating_score_wifi":    {"4"},
		"rating_override_wifi": {"Fast"},
		"rating_score_noise":   {"3"}, // not enabled -> ignored
	}
	got, err := ParseRatings(v, types)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].CategoryType != "wifi" || got[0].Score != 4 || *got[0].ShortDescriptionOverride != "Fast" {
		t.Fatalf("ratings: %+v", got)
	}

	got, err = ParseRatings(url.Values{"rating_enabled_wifi": {"1"}}, types)
	if err == nil || handlers.UserError(err) == "" || len(got) != 1 {
		t.Fatalf("enabled without score: %v %+v", err, got)
	}
}

func TestParsePartnershipForm(t *testing.T) {
	v := url.Values{
		"cafe_id":           {"7"},
		"reviewer_name":     {"Kopi Club"},
		"content":           {"Nice"},
		"overall_score":     {"4.5"},
		"visited_month":     {"2026-09"},
		"external_url_link": {"https://example.com"},
		"is_pinned":         {"1"},
	}
	rv, err := ParsePartnershipForm(v, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rv.CafeID != 7 || rv.UserID != nil || !rv.IsPinned || *rv.OverallScore != 4.5 || rv.WFCScore != nil || *rv.VisitedAt != "2026-09" {
		t.Fatalf("%+v", rv)
	}
	rv, err = ParsePartnershipForm(url.Values{"reviewer_name": {"x"}, "wfc_score": {"great"}}, nil)
	if err == nil || handlers.UserError(err) == "" || rv.ReviewerName != "x" {
		t.Fatalf("bad score: %v %+v", err, rv)
	}
}
