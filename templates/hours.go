package templates

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// hoursGroup is one block of the hours editor: the default row, a weekday or
// a special date. Key is its group key (handlers.ParseHours).
type hoursGroup struct {
	Key    string
	Label  string
	Closed bool
	Ranges []models.HourRange
}

// hoursSpecial is a special date (dated rows sharing day / from / until).
type hoursSpecial struct {
	hoursGroup
	From, Until, Day string
}

type hoursEditor struct {
	Regular  []hoursGroup // every day, then mon→sun
	Specials []hoursSpecial
}

// newHoursEditor groups cafe_hour rows into the editor's blocks. Regular
// blocks always show at least one (maybe empty) range row. It is the inverse
// of handlers.ParseHours.
func newHoursEditor(rows []models.CafeHour) hoursEditor {
	var e hoursEditor
	idx := map[string]int{}
	e.Regular = append(e.Regular, hoursGroup{Key: "all", Label: "Every day (default)"})
	idx["all"] = 0
	for i, d := range models.Weekdays {
		e.Regular = append(e.Regular, hoursGroup{Key: d, Label: models.DayLabel(d)})
		idx[d] = i + 1
	}
	specialIdx := map[string]int{}
	for _, h := range rows {
		var g *hoursGroup
		if !h.Dated() {
			key := "all"
			if h.Day != nil {
				key = *h.Day
			}
			i, ok := idx[key]
			if !ok {
				continue
			}
			g = &e.Regular[i]
		} else {
			sk := deref(h.Day) + "|" + deref(h.ValidFrom) + "|" + deref(h.ValidUntil)
			i, ok := specialIdx[sk]
			if !ok {
				i = len(e.Specials)
				specialIdx[sk] = i
				e.Specials = append(e.Specials, hoursSpecial{
					hoursGroup: hoursGroup{Key: "s" + strconv.Itoa(i)},
					From:       deref(h.ValidFrom), Until: deref(h.ValidUntil), Day: deref(h.Day),
				})
			}
			g = &e.Specials[i].hoursGroup
		}
		if h.Closed {
			g.Closed = true
		} else {
			g.Ranges = append(g.Ranges, models.HourRange{Open: deref(h.Open), Close: deref(h.Close)})
		}
	}
	for i := range e.Regular {
		if len(e.Regular[i].Ranges) == 0 {
			e.Regular[i].Ranges = []models.HourRange{{}}
		}
	}
	return e
}

// hoursPreview resolves rows for the 7 days from today (Jakarta).
func hoursPreview(rows []models.CafeHour) []models.DayHours {
	return models.ResolveDays(rows, models.TodayJakarta(time.Now()), 7)
}

// maxCompareDays caps the comparison table of an hours proposal.
const maxCompareDays = 31

type hoursCompareRow struct {
	Current, Proposed models.DayHours
	Differs           bool
}

// hoursComparison resolves the live rows (p.Current) and the proposed rows
// over the proposal's window. ok is false when new_value is not a per-day
// hours proposal (old shape): show the raw JSON instead.
func hoursComparison(p models.Proposal) (out []hoursCompareRow, ok bool) {
	hp, ok := models.ParseHoursProposal(p.NewValue)
	if !ok {
		return nil, false
	}
	var cur models.CurrentHours
	_ = json.Unmarshal(p.Current, &cur) // unparseable / null = no live rows
	from, _ := time.Parse(time.DateOnly, hp.WindowFrom)
	until, _ := time.Parse(time.DateOnly, hp.WindowUntil)
	n := min(int(until.Sub(from).Hours()/24)+1, maxCompareDays)
	c, pr := models.ResolveDays(cur.Rows, from, n), models.ResolveDays(hp.Rows, from, n)
	for i := range n {
		out = append(out, hoursCompareRow{Current: c[i], Proposed: pr[i], Differs: !c[i].SameHours(pr[i])})
	}
	return out, true
}

// hoursProposalRows returns the proposed rows of an hours proposal (nil if
// it does not parse).
func hoursProposalRows(raw []byte) []models.CafeHour {
	hp, _ := models.ParseHoursProposal(raw)
	return hp.Rows
}

func hoursWindow(raw []byte) (string, string) {
	hp, _ := models.ParseHoursProposal(raw)
	return hp.WindowFrom, hp.WindowUntil
}

// shortDay renders a YYYY-MM-DD date as "Sat 24 Oct".
func shortDay(date string) string {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return date
	}
	return t.Format("Mon 2 Jan")
}

func dayHoursCls(d models.DayHours) string {
	switch {
	case d.Closed:
		return "text-red-700"
	case d.Unknown():
		return "text-bark"
	}
	return "text-forest"
}
