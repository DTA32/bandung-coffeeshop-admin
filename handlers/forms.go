package handlers

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// formError is a user-facing parse error (bad number etc.).
type formError struct{ msg string }

func (e *formError) Error() string { return e.msg }

// FormErrorf returns a user-facing parse error that handlers show next to the form.
func FormErrorf(format string, args ...any) error {
	return &formError{fmt.Sprintf(format, args...)}
}

// OptString returns the trimmed value of key, or nil if the key is absent or blank.
func OptString(v url.Values, key string) *string {
	if _, ok := v[key]; !ok {
		return nil
	}
	s := v.Get(key)
	return controllers.NilIfBlank(&s)
}

// OptInt parses key as a whole number; blank gives nil.
func OptInt(v url.Values, key string) (*int, error) {
	s := strings.TrimSpace(v.Get(key))
	if s == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, &formError{fmt.Sprintf("%s: %q is not a whole number", key, s)}
	}
	return &n, nil
}

// OptFloat parses key as a number; blank gives nil.
func OptFloat(v url.Values, key string) (*float64, error) {
	s := strings.TrimSpace(v.Get(key))
	if s == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, &formError{fmt.Sprintf("%s: %q is not a number", key, s)}
	}
	return &f, nil
}

// Checked reports whether the checkbox key was submitted.
func Checked(v url.Values, key string) bool {
	return v.Get(key) != ""
}

// ParseImages reads the parallel image_url / image_desc / image_desc_indo
// lists; their order is the display order.
func ParseImages(v url.Values) []models.Image {
	urls := v["image_url"]
	desc := v["image_desc"]
	descIndo := v["image_desc_indo"]
	out := make([]models.Image, 0, len(urls))
	for i, u := range urls {
		im := models.Image{URL: strings.TrimSpace(u)}
		if i < len(desc) {
			im.Description = strings.TrimSpace(desc[i])
		}
		if i < len(descIndo) {
			im.DescriptionIndo = strings.TrimSpace(descIndo[i])
		}
		out = append(out, im)
	}
	return out
}

// HoursAllDays is the hours editor group key of the default (every day) row;
// the weekday groups use the day_enum values.
const HoursAllDays = "all"

// ParseHours reads the per-day hours editor (templates.HoursEditor) into
// cafe_hour rows. The editor posts, per time range, the parallel lists
// hours_group / hours_open / hours_close; hours_closed lists the closed
// group keys; and per special date the parallel lists special_key /
// special_from / special_until / special_day. Group keys are "all" (every
// day), mon..sun, or a special_key. Blank ranges are skipped; a blank special
// "until" means the same day as "from". A special date with neither hours nor
// "Closed" still yields a row, so validation (controllers.NormalizeHours)
// reports it instead of dropping it. Never fails: validation owns errors.
func ParseHours(v url.Values) []models.CafeHour {
	type group struct {
		base   models.CafeHour
		closed bool
		ranges [][2]string
	}
	groups := map[string]*group{HoursAllDays: {}}
	order := []string{HoursAllDays}
	for _, d := range models.Weekdays {
		day := d
		groups[d] = &group{base: models.CafeHour{Day: &day}}
		order = append(order, d)
	}
	from, until, days := v["special_from"], v["special_until"], v["special_day"]
	for i, key := range v["special_key"] {
		if _, dup := groups[key]; dup || key == "" {
			continue
		}
		f := strings.TrimSpace(at(from, i))
		u := strings.TrimSpace(at(until, i))
		if u == "" {
			u = f
		}
		g := &group{base: models.CafeHour{ValidFrom: &f, ValidUntil: &u}}
		if d := strings.TrimSpace(at(days, i)); slices.Contains(models.Weekdays, d) {
			g.base.Day = &d
		}
		groups[key] = g
		order = append(order, key)
	}
	for _, key := range v["hours_closed"] {
		if g, ok := groups[key]; ok {
			g.closed = true
		}
	}
	opens, closes := v["hours_open"], v["hours_close"]
	for i, key := range v["hours_group"] {
		g, ok := groups[key]
		o, c := strings.TrimSpace(at(opens, i)), strings.TrimSpace(at(closes, i))
		if !ok || (o == "" && c == "") {
			continue
		}
		g.ranges = append(g.ranges, [2]string{o, c})
	}

	var out []models.CafeHour
	for _, key := range order {
		g := groups[key]
		if g.closed {
			h := g.base
			h.Closed = true
			out = append(out, h)
		}
		for _, r := range g.ranges {
			h := g.base
			o, c := r[0], r[1]
			h.Open, h.Close = &o, &c
			out = append(out, h)
		}
		if !g.closed && len(g.ranges) == 0 && g.base.Dated() {
			out = append(out, g.base)
		}
	}
	return out
}

func at(list []string, i int) string {
	if i < len(list) {
		return list[i]
	}
	return ""
}
