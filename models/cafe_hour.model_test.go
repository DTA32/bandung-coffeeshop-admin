package models

import (
	"testing"
	"time"
)

// hr builds a row; "" leaves a field nil.
func hr(day, open, close, from, until string, closed bool) CafeHour {
	opt := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	return CafeHour{Day: opt(day), Open: opt(open), Close: opt(close), ValidFrom: opt(from), ValidUntil: opt(until), Closed: closed}
}

// Saturday 2026-10-24.
var sat = time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)

func TestResolveDay(t *testing.T) {
	tests := []struct {
		name string
		rows []CafeHour
		date time.Time
		want string
		spec bool
	}{
		{"no rows is unknown", nil, sat, "Unknown", false},
		{"every-day row", []CafeHour{hr("", "08:00", "22:00", "", "", false)}, sat, "08:00–22:00", false},
		{"weekday beats every day", []CafeHour{
			hr("", "08:00", "22:00", "", "", false), hr("sat", "10:00", "23:00", "", "", false),
		}, sat, "10:00–23:00", false},
		{"other weekday falls back to every day", []CafeHour{
			hr("", "08:00", "22:00", "", "", false), hr("mon", "", "", "", "", true),
		}, sat, "08:00–22:00", false},
		{"weekday closed", []CafeHour{
			hr("", "08:00", "22:00", "", "", false), hr("sat", "", "", "", "", true),
		}, sat, "Closed", false},
		{"split hours sorted", []CafeHour{
			hr("sat", "17:00", "22:00", "", "", false), hr("sat", "08:00", "14:00", "", "", false),
		}, sat, "08:00–14:00, 17:00–22:00", false},
		{"closed wins within a tier", []CafeHour{
			hr("sat", "08:00", "14:00", "", "", false), hr("sat", "", "", "", "", true),
		}, sat, "Closed", false},
		{"dated row beats regular", []CafeHour{
			hr("sat", "08:00", "22:00", "", "", false), hr("", "", "", "2026-10-24", "2026-10-24", true),
		}, sat, "Closed", true},
		{"dated row with day beats dated any day", []CafeHour{
			hr("", "09:00", "12:00", "2026-10-01", "2026-10-31", false), hr("sat", "10:00", "11:00", "2026-10-01", "2026-10-31", false),
		}, sat, "10:00–11:00", true},
		{"dated row outside its range is ignored", []CafeHour{
			hr("", "08:00", "22:00", "", "", false), hr("", "", "", "2026-10-25", "2026-10-25", true),
		}, sat, "08:00–22:00", false},
		{"one-bound dated row is still dated", []CafeHour{
			hr("", "08:00", "22:00", "", "", false), hr("", "07:00", "12:00", "2026-10-01", "", false),
		}, sat, "07:00–12:00", true},
		{"null hours are unknown", []CafeHour{hr("", "08:00", "", "", "", false)}, sat, "Unknown", false},
		{"24 hours", []CafeHour{hr("", "00:00", "23:59", "", "", false)}, sat, "Open 24 hours", false},
	}
	for _, tc := range tests {
		got := ResolveDay(tc.rows, tc.date)
		if got.String() != tc.want || got.Special != tc.spec || got.Day != "sat" || got.Date != "2026-10-24" {
			t.Errorf("%s: got %+v (%s)", tc.name, got, got)
		}
	}
}

func TestResolveDaysAndSameHours(t *testing.T) {
	// A null-day row equals seven identical weekday rows.
	every := []CafeHour{hr("", "08:00", "22:00", "", "", false)}
	var week []CafeHour
	for _, d := range Weekdays {
		week = append(week, hr(d, "08:00", "22:00", "", "", false))
	}
	a, b := ResolveDays(every, sat, 7), ResolveDays(week, sat, 7)
	if len(a) != 7 || a[6].Date != "2026-10-30" || a[6].Day != "fri" {
		t.Fatalf("days: %+v", a)
	}
	for i := range a {
		if !a[i].SameHours(b[i]) {
			t.Fatalf("day %d differs: %v vs %v", i, a[i], b[i])
		}
	}
	if (DayHours{Closed: true}).SameHours(DayHours{}) {
		t.Fatal("closed must differ from unknown")
	}
}

func TestTodayJakarta(t *testing.T) {
	// 18:30 UTC on the 23rd is 01:30 on the 24th in Jakarta.
	got := TodayJakarta(time.Date(2026, 10, 23, 18, 30, 0, 0, time.UTC))
	if got.Format(time.DateOnly) != "2026-10-24" {
		t.Fatal(got)
	}
}

func TestReplacedByWindow(t *testing.T) {
	const from, until = "2026-10-24", "2026-10-30"
	cases := []struct {
		row  CafeHour
		want bool
	}{
		{hr("", "08:00", "22:00", "", "", false), true},                        // undated every day
		{hr("mon", "", "", "", "", true), true},                                // undated weekday
		{hr("", "", "", "2026-10-25", "2026-10-25", true), true},               // inside
		{hr("", "", "", "2026-10-20", "2026-10-24", true), true},               // touches the start
		{hr("", "", "", "2026-10-30", "2026-11-02", true), true},               // touches the end
		{hr("", "", "", "2026-10-01", "2026-12-31", true), true},               // spans the window
		{hr("", "", "", "2026-10-23", "2026-10-23", true), false},              // before
		{hr("", "", "", "2026-10-31", "2026-10-31", true), false},              // after
		{hr("", "", "", "2026-11-01", "", true), false},                        // open-ended, starts after
		{hr("", "", "", "2026-10-01", "", true), true},                         // open-ended, covers
		{hr("", "", "", "", "2026-10-23", true), false},                        // open start, ends before
		{hr("sat", "10:00", "12:00", "2026-10-24", "2026-10-24", false), true}, // dated weekday inside
	}
	for i, tc := range cases {
		if got := ReplacedByWindow(tc.row, from, until); got != tc.want {
			t.Errorf("case %d: got %v want %v", i, got, tc.want)
		}
	}
}

func TestParseHoursProposal(t *testing.T) {
	if _, ok := ParseHoursProposal([]byte(`{"open":"08:00","close":"22:00","days":[]}`)); ok {
		t.Fatal("old shape must not parse")
	}
	if _, ok := ParseHoursProposal([]byte(`{"window_from":"2026-10-30","window_until":"2026-10-24","rows":[]}`)); ok {
		t.Fatal("inverted window must not parse")
	}
	hp, ok := ParseHoursProposal([]byte(`{"window_from":"2026-10-24","window_until":"2026-10-30",
		"rows":[{"day":"mon","open":"08:00","close":"22:00","valid_from":null,"valid_until":null,"closed":false}],
		"days":[{"day":"Senin","date":"2026-10-26","hours":"08.00–22.00","special":false}]}`))
	if !ok || len(hp.Rows) != 1 || *hp.Rows[0].Day != "mon" || hp.Rows[0].Dated() {
		t.Fatalf("%+v %v", hp, ok)
	}
}
