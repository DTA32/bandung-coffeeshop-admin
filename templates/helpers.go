package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config is set once at startup.
var (
	TZ        = time.UTC
	ImageBase = "https://image.bdgcafe.com"
	// ImageUploads shows the R2 upload buttons in ImagesEditor.
	ImageUploads = false
)

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intStr(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}

func floatStr(p *float64) string {
	if p == nil {
		return ""
	}
	return strconv.FormatFloat(*p, 'f', -1, 64)
}

func fmtFloat(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(TZ).Format("2006-01-02 15:04")
}

func fmtTimePtr(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return fmtTime(*t)
}

// inputTime renders a time for <input type="datetime-local">.
func inputTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.In(TZ).Format("2006-01-02T15:04")
}

func ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func agoPtr(t *time.Time) string {
	if t == nil {
		return "never"
	}
	return ago(*t)
}

// prettyJSON indents a jsonb value for display.
func prettyJSON(raw []byte) string {
	if len(raw) == 0 {
		return "null"
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}
	return buf.String()
}

// displayValue renders a proposal value (old/new/current) as one line.
func displayValue(field string, raw []byte) string {
	if len(raw) == 0 || string(raw) == "null" {
		return "—"
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var b bool
	if json.Unmarshal(raw, &b) == nil {
		if b {
			return "yes"
		}
		return "no"
	}
	var m map[string]any
	if json.Unmarshal(raw, &m) == nil {
		if u, ok := m["url"].(string); ok {
			return u
		}
	}
	return string(raw)
}

// hoursDays extracts the raw per-day hours carried in an hours proposal.
func hoursDays(raw []byte) []string {
	var h struct {
		Days []struct {
			Day     string `json:"day"`
			Date    string `json:"date"`
			Hours   string `json:"hours"`
			Special bool   `json:"special"`
		} `json:"days"`
	}
	if json.Unmarshal(raw, &h) != nil {
		return nil
	}
	out := make([]string, 0, len(h.Days))
	for _, d := range h.Days {
		line := d.Day
		if d.Date != "" {
			line += " " + d.Date
		}
		line += ": " + d.Hours
		if d.Special {
			line += " (special hours)"
		}
		out = append(out, line)
	}
	return out
}

func jsonString(raw []byte) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return ""
}

var fieldLabels = map[string]string{
	"hours":        "Opening hours",
	"phone":        "Phone",
	"address":      "Address (description)",
	"status":       "Status",
	"temp_closed":  "Temporarily closed",
	"menu":         "Menu link",
	"menu_content": "Menu file changed",
	"ig_bio":       "Instagram bio",
	"ig_post":      "Instagram post",
}

func fieldLabel(f string) string {
	if l, ok := fieldLabels[f]; ok {
		return l
	}
	return f
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n])) + "…"
}

func imageURL(key string) string {
	return ImageBase + "/" + strings.TrimLeft(key, "/")
}

// thumb asks the image service for a small variant.
func thumb(u string, w int) string {
	if strings.HasPrefix(u, ImageBase+"/") {
		sep := "?"
		if strings.Contains(u, "?") {
			sep = "&"
		}
		return fmt.Sprintf("%s%sw=%d", u, sep, w)
	}
	return u
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func itoa(i int) string  { return strconv.Itoa(i) }
func i64(i int64) string { return strconv.FormatInt(i, 10) }
func i64p(i *int64) string {
	if i == nil {
		return ""
	}
	return strconv.FormatInt(*i, 10)
}

func lower(s string) string { return strings.ToLower(s) }

// emptyListText is the empty-state message of a searchable list.
func emptyListText(what, q string) string {
	if q != "" {
		return "No " + what + " match \"" + q + "\"."
	}
	return "No " + what + " found."
}

func boolParam(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
