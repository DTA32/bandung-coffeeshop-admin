package templates

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func sp(s string) *string { return &s }

func TestNewHoursEditor(t *testing.T) {
	e := newHoursEditor([]models.CafeHour{
		{Open: sp("08:00"), Close: sp("22:00")},
		{Day: sp("mon"), Closed: true},
		{Day: sp("sat"), Open: sp("08:00"), Close: sp("14:00")},
		{Day: sp("sat"), Open: sp("17:00"), Close: sp("22:00")},
		{ValidFrom: sp("2026-12-25"), ValidUntil: sp("2026-12-25"), Closed: true},
		{ValidFrom: sp("2026-12-31"), ValidUntil: sp("2026-12-31"), Open: sp("08:00"), Close: sp("12:00")},
		{ValidFrom: sp("2026-12-31"), ValidUntil: sp("2026-12-31"), Open: sp("18:00"), Close: sp("23:00")},
	})
	if len(e.Regular) != 8 || e.Regular[0].Key != "all" || e.Regular[0].Ranges[0].Open != "08:00" {
		t.Fatalf("regular: %+v", e.Regular)
	}
	mon, tue, sat := e.Regular[1], e.Regular[2], e.Regular[6]
	if !mon.Closed || len(tue.Ranges) != 1 || tue.Ranges[0] != (models.HourRange{}) || len(sat.Ranges) != 2 {
		t.Fatalf("weekdays: %+v %+v %+v", mon, tue, sat)
	}
	if len(e.Specials) != 2 || !e.Specials[0].Closed || len(e.Specials[1].Ranges) != 2 || e.Specials[1].Key != "s1" {
		t.Fatalf("specials: %+v", e.Specials)
	}
}

func TestHoursEditorRenders(t *testing.T) {
	var buf bytes.Buffer
	rows := []models.CafeHour{{Day: sp("mon"), Closed: true}, {ValidFrom: sp("2026-12-25"), ValidUntil: sp("2026-12-25"), Closed: true}}
	if err := HoursEditor(rows).Render(context.Background(), &buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, want := range []string{`name="hours_closed" value="mon" checked`, `name="special_key" value="s0"`, `value="2026-12-25"`, `data-hours-range-template`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestHoursComparison(t *testing.T) {
	cur, _ := json.Marshal(models.CurrentHours{Rows: []models.CafeHour{{Open: sp("08:00"), Close: sp("22:00")}}})
	p := models.Proposal{Field: models.FieldHours, Current: cur, NewValue: json.RawMessage(`{
		"window_from": "2026-10-24", "window_until": "2026-10-30",
		"rows": [
			{"day": null, "open": "08:00", "close": "22:00", "valid_from": null, "valid_until": null, "closed": false},
			{"day": "mon", "open": null, "close": null, "valid_from": null, "valid_until": null, "closed": true},
			{"day": null, "open": "10:00", "close": "14:00", "valid_from": "2026-10-28", "valid_until": "2026-10-28", "closed": false}
		]}`)}
	rows, ok := hoursComparison(p)
	if !ok || len(rows) != 7 {
		t.Fatalf("%v %d", ok, len(rows))
	}
	// sat 24 same; mon 26 closed; wed 28 special.
	if rows[0].Differs || !rows[2].Differs || !rows[2].Proposed.Closed || !rows[4].Proposed.Special || !rows[4].Differs || rows[6].Differs {
		t.Fatalf("rows: %+v", rows)
	}
	p.NewValue = json.RawMessage(`{"open":"08:00","close":"22:00","days":[]}`)
	if _, ok := hoursComparison(p); ok || approvable(p) {
		t.Fatal("old shape must not compare or approve")
	}
}
