package controllers

import (
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestNormalizeHours(t *testing.T) {
	rows, err := NormalizeHours([]models.CafeHour{
		{ValidFrom: ptr("2026-12-25"), ValidUntil: ptr("2026-12-25"), Closed: true},
		{Day: ptr("sat"), Open: ptr("17:00"), Close: ptr("23:00")},
		{Day: ptr("sat"), Open: ptr("08:00"), Close: ptr("14:00")},
		{Day: ptr(""), Open: ptr("08:00"), Close: ptr("22:00")},
		{Day: ptr("mon"), Closed: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, h := range rows {
		got = append(got, HourRowLabel(h)+" "+deref(h.Open))
	}
	want := "Every day 08:00|Monday |Saturday 08:00|Saturday 17:00|2026-12-25 "
	if strings.Join(got, "|") != want {
		t.Fatalf("order: %q", strings.Join(got, "|"))
	}
	if rows[0].Day != nil {
		t.Fatal("blank day must become NULL (every day)")
	}

	bad := []struct {
		name string
		rows []models.CafeHour
		want string
	}{
		{"bad hour", []models.CafeHour{{Open: ptr("8:00"), Close: ptr("22:00")}}, "invalid hour"},
		{"half range", []models.CafeHour{{Day: ptr("mon"), Open: ptr("08:00")}}, "Monday: set both"},
		{"no hours", []models.CafeHour{{}}, "set both"},
		{"closed with hours", []models.CafeHour{{Closed: true, Open: ptr("08:00"), Close: ptr("10:00")}}, "closed day cannot have hours"},
		{"closed and open", []models.CafeHour{{Day: ptr("sun"), Closed: true}, {Day: ptr("sun"), Open: ptr("08:00"), Close: ptr("10:00")}}, "Sunday is marked closed and also has hours"},
		{"bad day", []models.CafeHour{{Day: ptr("monday"), Closed: true}}, "invalid day"},
		{"bad date", []models.CafeHour{{ValidFrom: ptr("25/12/2026"), Closed: true}}, "invalid date"},
		{"blank date", []models.CafeHour{{ValidFrom: ptr(""), ValidUntil: ptr(""), Closed: true}}, "needs a date"},
		{"from after until", []models.CafeHour{{ValidFrom: ptr("2026-12-26"), ValidUntil: ptr("2026-12-25"), Closed: true}}, "start date is after"},
	}
	for _, tc := range bad {
		if _, err := NormalizeHours(tc.rows); err == nil || !IsValidation(err) || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want %q", tc.name, err, tc.want)
		}
	}
	// Split hours and the same weekday on a special date are fine.
	if _, err := NormalizeHours([]models.CafeHour{
		{Day: ptr("sun"), Closed: true},
		{Day: ptr("sun"), ValidFrom: ptr("2026-12-27"), ValidUntil: ptr("2026-12-27"), Open: ptr("10:00"), Close: ptr("12:00")},
	}); err != nil {
		t.Fatal(err)
	}
}
