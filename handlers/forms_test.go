package handlers

import (
	"net/url"
	"testing"
)

func TestParseHours(t *testing.T) {
	v := url.Values{
		// every day 08-22; mon closed (its blank range is ignored); sat split;
		// special s0 = one closed day; n1 = a range with "until" left blank;
		// n2 = a date with nothing set (kept for validation); a stray group.
		"hours_group":   {"all", "mon", "sat", "sat", "s0", "n1", "n2", "zzz"},
		"hours_open":    {"08:00", "", "08:00", "17:00", "", " 10:00 ", "", "01:00"},
		"hours_close":   {"22:00", "", "14:00", "22:00", "", "12:00", "", "02:00"},
		"hours_closed":  {"mon", "s0"},
		"special_key":   {"s0", "n1", "n2"},
		"special_from":  {"2026-12-25", "2026-12-31", ""},
		"special_until": {"2026-12-26", "", ""},
		"special_day":   {"", "thu", ""},
	}
	rows := ParseHours(v)
	str := func(s *string) string {
		if s == nil {
			return "-"
		}
		return *s
	}
	var got []string
	for _, h := range rows {
		c := ""
		if h.Closed {
			c = " closed"
		}
		got = append(got, str(h.Day)+" "+str(h.Open)+"-"+str(h.Close)+" "+str(h.ValidFrom)+".."+str(h.ValidUntil)+c)
	}
	want := []string{
		"- 08:00-22:00 -..-",
		"mon --- -..- closed",
		"sat 08:00-14:00 -..-",
		"sat 17:00-22:00 -..-",
		"- --- 2026-12-25..2026-12-26 closed",
		"thu 10:00-12:00 2026-12-31..2026-12-31",
		"- --- ..",
	}
	if len(got) != len(want) {
		t.Fatalf("got %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %q want %q", i, got[i], want[i])
		}
	}
	if len(ParseHours(url.Values{})) != 0 {
		t.Fatal("empty editor must give no rows")
	}
}
