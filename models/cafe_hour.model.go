package models

import (
	"encoding/json"
	"slices"
	"strings"
	"time"
)

// CafeHour is one cafe_hour row (BE migration 013) and also the row shape of
// an hours proposal (see HoursProposal). Day nil = every day; ValidFrom and
// ValidUntil nil on both = regular hours, otherwise a dated row (special
// hours, ValidUntil inclusive). Closed = the cafe_hour.close column.
type CafeHour struct {
	ID         int     `json:"-"`
	Day        *string `json:"day"`         // mon..sun or nil
	Open       *string `json:"open"`        // HH:MM
	Close      *string `json:"close"`       // HH:MM
	ValidFrom  *string `json:"valid_from"`  // YYYY-MM-DD
	ValidUntil *string `json:"valid_until"` // YYYY-MM-DD
	Closed     bool    `json:"closed"`
}

// Dated reports whether h is a dated (special hours) row. A row with only
// one bound is still dated.
func (h CafeHour) Dated() bool { return h.ValidFrom != nil || h.ValidUntil != nil }

// HoursProposal is new_value of an "hours" proposal (written by
// bandung-coffeeshop-scraper); Approve replaces the undated rows and the dated
// rows overlapping [WindowFrom, WindowUntil] with Rows. Days is the raw
// scraped week, for display only.
type HoursProposal struct {
	WindowFrom  string          `json:"window_from"`
	WindowUntil string          `json:"window_until"`
	Rows        []CafeHour      `json:"rows"`
	Days        json.RawMessage `json:"days,omitempty"`
}

// ParseHoursProposal decodes an hours proposal value. ok is false when raw is
// not in the per-day shape (e.g. the old {open, close, days} proposals): no
// valid window or no rows key.
func ParseHoursProposal(raw []byte) (HoursProposal, bool) {
	var hp HoursProposal
	if json.Unmarshal(raw, &hp) != nil || hp.Rows == nil {
		return HoursProposal{}, false
	}
	from, err1 := time.Parse(time.DateOnly, hp.WindowFrom)
	until, err2 := time.Parse(time.DateOnly, hp.WindowUntil)
	if err1 != nil || err2 != nil || until.Before(from) {
		return HoursProposal{}, false
	}
	return hp, true
}

// CurrentHours is the live value of the hours field as the changes queue
// reads it (repository currentValueSQL): the cafe's cafe_hour rows.
type CurrentHours struct {
	Rows []CafeHour `json:"rows"`
}

// ReplacedByWindow reports whether approving an hours proposal over the
// window [from, until] (YYYY-MM-DD, inclusive) replaces the live row h:
// every undated row, plus the dated rows overlapping the window (a nil bound
// is open). Dated rows wholly outside the window are kept.
func ReplacedByWindow(h CafeHour, from, until string) bool {
	if !h.Dated() {
		return true
	}
	if h.ValidFrom != nil && *h.ValidFrom > until {
		return false
	}
	if h.ValidUntil != nil && *h.ValidUntil < from {
		return false
	}
	return true
}

// Weekdays are the day_enum values, mon→sun.
var Weekdays = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

var dayNames = [...]string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}

var dayLabels = map[string]string{
	"mon": "Monday", "tue": "Tuesday", "wed": "Wednesday", "thu": "Thursday",
	"fri": "Friday", "sat": "Saturday", "sun": "Sunday",
}

// DayLabel is the English name of a day_enum value ("" = every day).
func DayLabel(day string) string {
	if l, ok := dayLabels[day]; ok {
		return l
	}
	return "Every day"
}

// DayOf returns the day_enum value of t's weekday.
func DayOf(t time.Time) string { return dayNames[t.Weekday()] }

// Jakarta is Asia/Jakarta as a fixed UTC+7 zone (no DST; the image may lack
// tzdata). "Today" for opening hours is always a Jakarta date.
var Jakarta = time.FixedZone("Asia/Jakarta", 7*60*60)

// HourRange is one open range of a day; Close < Open means overnight.
type HourRange struct {
	Open  string
	Close string
}

// DayHours is the resolved schedule of one date. Special = it came from a
// dated row. Not closed with no ranges = unknown hours.
type DayHours struct {
	Date    string // YYYY-MM-DD
	Day     string // mon..sun
	Closed  bool
	Special bool
	Ranges  []HourRange
}

// Unknown reports whether no row gave hours for the day.
func (d DayHours) Unknown() bool { return !d.Closed && len(d.Ranges) == 0 }

// SameHours compares the schedules (closed + ranges), not the special flag.
func (d DayHours) SameHours(o DayHours) bool {
	return d.Closed == o.Closed && slices.Equal(d.Ranges, o.Ranges)
}

// String renders the schedule: "Closed", "Unknown", "Open 24 hours" or
// "08:00–14:00, 17:00–22:00".
func (d DayHours) String() string {
	switch {
	case d.Closed:
		return "Closed"
	case len(d.Ranges) == 0:
		return "Unknown"
	case len(d.Ranges) == 1 && d.Ranges[0] == HourRange{Open: "00:00", Close: "23:59"}:
		return "Open 24 hours"
	}
	parts := make([]string, len(d.Ranges))
	for i, r := range d.Ranges {
		parts[i] = r.Open + "–" + r.Close
	}
	return strings.Join(parts, ", ")
}

// Tiers, most specific first (bandung-coffeeshop-be CLAUDE.md "Opening
// hours"). Dated rows are one tier where a matching day beats a null day.
const (
	tierNone = iota
	tierDatedDay
	tierDatedAny
	tierDay
	tierAny
)

// hourTier returns the tier of row h for date (YYYY-MM-DD) on weekday day, or
// tierNone when h does not apply to it.
func hourTier(h CafeHour, date, day string) int {
	if h.Day != nil && *h.Day != day {
		return tierNone
	}
	if !h.Dated() {
		if h.Day != nil {
			return tierDay
		}
		return tierAny
	}
	// Dates are YYYY-MM-DD, so they compare as strings. A null bound is open.
	if (h.ValidFrom != nil && *h.ValidFrom > date) || (h.ValidUntil != nil && *h.ValidUntil < date) {
		return tierNone
	}
	if h.Day != nil {
		return tierDatedDay
	}
	return tierDatedAny
}

// ResolveDay applies the rows of the most specific tier present for date (a
// Jakarta date; only its calendar date is used). Within the tier a closed row
// wins over open rows; rows with a null open or close are ignored. Ported
// from bandung-coffeeshop-be service/cafe.go — keep the two in sync.
func ResolveDay(rows []CafeHour, date time.Time) DayHours {
	d := DayHours{Date: date.Format(time.DateOnly), Day: DayOf(date), Ranges: []HourRange{}}
	best := tierNone
	for _, h := range rows {
		if t := hourTier(h, d.Date, d.Day); t != tierNone && (best == tierNone || t < best) {
			best = t
		}
	}
	if best == tierNone {
		return d
	}
	d.Special = best == tierDatedDay || best == tierDatedAny
	for _, h := range rows {
		if hourTier(h, d.Date, d.Day) != best {
			continue
		}
		if h.Closed {
			d.Closed, d.Ranges = true, []HourRange{}
			return d
		}
		if h.Open != nil && h.Close != nil {
			d.Ranges = append(d.Ranges, HourRange{Open: *h.Open, Close: *h.Close})
		}
	}
	slices.SortFunc(d.Ranges, func(a, b HourRange) int {
		if c := strings.Compare(a.Open, b.Open); c != 0 {
			return c
		}
		return strings.Compare(a.Close, b.Close)
	})
	return d
}

// ResolveDays resolves n consecutive dates starting at from.
func ResolveDays(rows []CafeHour, from time.Time, n int) []DayHours {
	out := make([]DayHours, 0, n)
	for i := range n {
		out = append(out, ResolveDay(rows, from.AddDate(0, 0, i)))
	}
	return out
}

// TodayJakarta is now's calendar date in Jakarta, at midnight UTC (so
// AddDate never crosses a zone transition).
func TodayJakarta(now time.Time) time.Time {
	y, m, d := now.In(Jakarta).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
