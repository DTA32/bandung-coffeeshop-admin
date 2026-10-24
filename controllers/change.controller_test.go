package controllers

import (
	"encoding/json"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestFieldUpdateFor(t *testing.T) {
	hours := models.Proposal{Field: models.FieldHours, NewValue: json.RawMessage(`{
		"window_from": "2026-10-24", "window_until": "2026-10-30",
		"rows": [
			{"day": "sat", "open": "17:00", "close": "22:00", "valid_from": null, "valid_until": null, "closed": false},
			{"day": null, "open": "08:00", "close": "22:00", "valid_from": null, "valid_until": null, "closed": false},
			{"day": null, "open": null, "close": null, "valid_from": "2026-10-28", "valid_until": "2026-10-28", "closed": true}
		],
		"days": [{"day": "Sabtu", "date": "2026-10-24", "hours": "17.00–22.00", "special": false}]}`)}
	u, resolved, err := FieldUpdateFor(hours, nil)
	if err != nil || u.Hours == nil || u.Hours.WindowFrom != "2026-10-24" || len(u.Hours.Rows) != 3 {
		t.Fatalf("hours as proposed: %+v %v", u, err)
	}
	if u.Hours.Rows[0].Day != nil || *u.Hours.Rows[2].ValidFrom != "2026-10-28" {
		t.Fatalf("rows not normalised / sorted: %+v", u.Hours.Rows)
	}
	if hp := resolved.(models.HoursProposal); hp.Days != nil || len(hp.Rows) != 3 {
		t.Fatalf("resolved: %+v", resolved)
	}
	u, _, err = FieldUpdateFor(hours, &Edit{Hours: []models.CafeHour{{Open: ptr("09:00"), Close: ptr("21:30")}}})
	if err != nil || len(u.Hours.Rows) != 1 || *u.Hours.Rows[0].Open != "09:00" || u.Hours.WindowUntil != "2026-10-30" {
		t.Fatalf("hours edited: %+v %v", u, err)
	}
	for name, edit := range map[string]*Edit{
		"half range":     {Hours: []models.CafeHour{{Open: ptr("09:00")}}},
		"outside window": {Hours: []models.CafeHour{{ValidFrom: ptr("2026-11-01"), ValidUntil: ptr("2026-11-01"), Closed: true}}},
	} {
		if _, _, err := FieldUpdateFor(hours, edit); err == nil || !IsValidation(err) {
			t.Fatalf("%s must fail validation, got %v", name, err)
		}
	}
	old := models.Proposal{Field: models.FieldHours, NewValue: json.RawMessage(`{"open":"08:00","close":"00:00","days":[]}`)}
	if _, _, err := FieldUpdateFor(old, nil); err == nil || !IsValidation(err) {
		t.Fatalf("old-shape hours must be a validation error, got %v", err)
	}

	phone := models.Proposal{Field: models.FieldPhone, NewValue: json.RawMessage(`"+62227777"`)}
	u, resolved, err = FieldUpdateFor(phone, nil)
	if err != nil || *u.Text != "+62227777" || resolved != "+62227777" {
		t.Fatalf("phone: %+v %v %v", u, resolved, err)
	}
	u, _, _ = FieldUpdateFor(phone, &Edit{Value: " +62 22 1 "})
	if *u.Text != "+62 22 1" {
		t.Fatalf("edited phone not trimmed: %q", *u.Text)
	}

	cases := []struct {
		p    models.Proposal
		edit *Edit
	}{
		{models.Proposal{Field: models.FieldMenu, NewValue: json.RawMessage(`"ftp://x"`)}, nil},
		{models.Proposal{Field: models.FieldMenu, NewValue: json.RawMessage(`"#image-1"`)}, nil},
		{models.Proposal{Field: models.FieldMenu}, &Edit{Value: "#image-2"}},
		{models.Proposal{Field: models.FieldStatus, NewValue: json.RawMessage(`"deleted"`)}, nil},
		{models.Proposal{Field: models.FieldAddress}, &Edit{Value: "  "}},
		{models.Proposal{Field: models.FieldIGBio, NewValue: json.RawMessage(`"bio"`)}, nil},
		{models.Proposal{Field: models.FieldIGPost, NewValue: json.RawMessage(`{}`)}, nil},
		{models.Proposal{Field: "mystery", NewValue: json.RawMessage(`"x"`)}, nil},
	}
	for _, tc := range cases {
		if _, _, err := FieldUpdateFor(tc.p, tc.edit); err == nil || !IsValidation(err) {
			t.Fatalf("%s: expected validation error, got %v", tc.p.Field, err)
		}
	}
}

func TestGroupByCafe(t *testing.T) {
	items := []models.Proposal{{ID: 1, CafeID: 1}, {ID: 2, CafeID: 1}, {ID: 3, CafeID: 2}}
	g := GroupByCafe(items)
	if len(g) != 2 || len(g[0].Items) != 2 || len(g[1].Items) != 1 {
		t.Fatalf("groups: %+v", g)
	}
	if GroupByCafe(nil) != nil {
		t.Fatal("empty input should give no groups")
	}
}
