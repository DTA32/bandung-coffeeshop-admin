package controllers

import (
	"encoding/json"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestFieldUpdateFor(t *testing.T) {
	hours := models.Proposal{Field: models.FieldHours, NewValue: json.RawMessage(`{"open":"08:00","close":"00:00","days":[]}`)}
	u, resolved, err := FieldUpdateFor(hours, nil)
	if err != nil || *u.OpenHour != "08:00" || *u.CloseHour != "00:00" {
		t.Fatalf("hours as proposed: %+v %v", u, err)
	}
	if m := resolved.(map[string]string); m["open"] != "08:00" || m["close"] != "00:00" {
		t.Fatalf("resolved: %v", resolved)
	}
	u, _, err = FieldUpdateFor(hours, &Edit{Open: "09:00", Close: "21:30"})
	if err != nil || *u.OpenHour != "09:00" || *u.CloseHour != "21:30" {
		t.Fatalf("hours edited: %+v %v", u, err)
	}
	if _, _, err := FieldUpdateFor(hours, &Edit{Open: "09:00"}); err == nil {
		t.Fatal("missing close must fail")
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
