package controllers

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// NormalizeHours trims and validates cafe_hour rows (from the cafe form, an
// hours proposal or its edit) and returns them sorted: undated first, then
// dated by date; every day before mon→sun; ranges by open hour.
//
// Rules: day is a day_enum value or nil; hours are HH:MM; an open row needs
// both open and close; a closed row has no hours, and the same day / dates
// cannot be both closed and open; dates are YYYY-MM-DD with from <= until.
func NormalizeHours(rows []models.CafeHour) ([]models.CafeHour, error) {
	out := make([]models.CafeHour, 0, len(rows))
	closedKeys, openKeys := map[string]bool{}, map[string]bool{}
	for _, h := range rows {
		h.Day = NilIfBlank(h.Day)
		h.Open, h.Close = NilIfBlank(h.Open), NilIfBlank(h.Close)
		if h.Day != nil && !slices.Contains(models.Weekdays, *h.Day) {
			return nil, invalid("invalid day %q", *h.Day)
		}
		for _, d := range []**string{&h.ValidFrom, &h.ValidUntil} {
			if *d == nil {
				continue
			}
			s := strings.TrimSpace(**d)
			if s == "" {
				return nil, invalid("a special date needs a date")
			}
			if _, err := time.Parse(time.DateOnly, s); err != nil {
				return nil, invalid("invalid date %q: expected YYYY-MM-DD", s)
			}
			*d = &s
		}
		label := HourRowLabel(h)
		if h.ValidFrom != nil && h.ValidUntil != nil && *h.ValidFrom > *h.ValidUntil {
			return nil, invalid("%s: the start date is after the end date", label)
		}
		if h.Closed {
			if h.Open != nil || h.Close != nil {
				return nil, invalid("%s: a closed day cannot have hours", label)
			}
		} else {
			if h.Open == nil || h.Close == nil {
				return nil, invalid("%s: set both open and close (HH:MM), or mark it closed", label)
			}
			if err := validateHour(h.Open); err != nil {
				return nil, invalid("%s: %s", label, err.Error())
			}
			if err := validateHour(h.Close); err != nil {
				return nil, invalid("%s: %s", label, err.Error())
			}
		}
		key := hourGroupKey(h)
		if h.Closed {
			closedKeys[key] = true
		} else {
			openKeys[key] = true
		}
		if closedKeys[key] && openKeys[key] {
			return nil, invalid("%s is marked closed and also has hours", label)
		}
		out = append(out, h)
	}
	slices.SortStableFunc(out, compareHourRows)
	return out, nil
}

// HourRowLabel names a row's day / dates in messages ("Monday", "Every day",
// "2026-12-25", "2026-12-24 – 2026-12-26 (Friday)").
func HourRowLabel(h models.CafeHour) string {
	day := models.DayLabel(deref(h.Day))
	if !h.Dated() {
		return day
	}
	from, until := deref(h.ValidFrom), deref(h.ValidUntil)
	s := from
	switch {
	case from == "":
		s = "until " + until
	case until == "":
		s = "from " + from
	case from != until:
		s = from + " – " + until
	}
	if h.Day != nil {
		s += " (" + day + ")"
	}
	return s
}

// hourGroupKey identifies rows that apply to the same day / dates.
func hourGroupKey(h models.CafeHour) string {
	return deref(h.Day) + "|" + deref(h.ValidFrom) + "|" + deref(h.ValidUntil)
}

func compareHourRows(a, b models.CafeHour) int {
	if c := cmp.Compare(boolInt(a.Dated()), boolInt(b.Dated())); c != 0 {
		return c
	}
	if c := cmp.Compare(deref(a.ValidFrom), deref(b.ValidFrom)); c != 0 {
		return c
	}
	if c := cmp.Compare(deref(a.ValidUntil), deref(b.ValidUntil)); c != 0 {
		return c
	}
	if c := cmp.Compare(dayIndex(a.Day), dayIndex(b.Day)); c != 0 {
		return c
	}
	if c := cmp.Compare(boolInt(!a.Closed), boolInt(!b.Closed)); c != 0 {
		return c
	}
	return cmp.Compare(deref(a.Open), deref(b.Open))
}

// dayIndex orders nil (every day) before mon→sun.
func dayIndex(d *string) int {
	if d == nil {
		return -1
	}
	return slices.Index(models.Weekdays, *d)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
