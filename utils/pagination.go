package utils

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pagination / typeahead convention shared by every list page.
//
//   - ?page=N (1-based, default 1) selects the page; ?q= is the search text.
//   - A request is a *fragment* request (rows + next sentinel only, no layout)
//     when it is a GET with HX-Request set, is not an htmx history restore, and
//     either page >= 2 (infinite-scroll sentinel) or HX-Target is the list
//     container id (typeahead swaps the container's innerHTML). Everything else,
//     including a reload of ?q=..., gets the full page.
const (
	MaxPage       = 10000
	MaxQueryRunes = 100
)

// ListParams are the parsed paging / search parameters of a list request.
type ListParams struct {
	Page     int    // 1-based
	Q        string // trimmed, sanitised search text
	Fragment bool   // render only the list body
}

// ParseListParams reads page / q from the query string. listID is the id of the
// list container that typeahead requests target ("" when the page has no search).
func ParseListParams(r *http.Request, listID string) ListParams {
	qv := r.URL.Query()
	p := ListParams{Page: 1, Q: CleanQuery(qv.Get("q"))}
	if n, err := strconv.Atoi(qv.Get("page")); err == nil && n >= 1 {
		p.Page = min(n, MaxPage)
	}
	if r.Method == http.MethodGet && IsHTMX(r) && r.Header.Get("HX-History-Restore-Request") != "true" {
		p.Fragment = p.Page >= 2 || (listID != "" && r.Header.Get("HX-Target") == listID)
	}
	return p
}

// CleanQuery trims s, drops invalid UTF-8 and control characters (Postgres
// rejects NUL) and caps it at MaxQueryRunes.
func CleanQuery(s string) string {
	s = strings.ToValidUTF8(s, "")
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > MaxQueryRunes {
		s = string([]rune(s)[:MaxQueryRunes])
	}
	return s
}

// NextURL builds the URL of page `page` of path with the given params
// (empty values and any "page" key are dropped; keys are sorted).
func NextURL(path string, params url.Values, page int) string {
	v := url.Values{}
	for k, vals := range params {
		if k == "page" {
			continue
		}
		for _, x := range vals {
			if x != "" {
				v.Add(k, x)
			}
		}
	}
	v.Set("page", strconv.Itoa(page))
	return path + "?" + v.Encode()
}

// Next returns the URL of the following page, or "" when there is none.
func (p ListParams) Next(path string, params url.Values, hasMore bool) string {
	if !hasMore {
		return ""
	}
	return NextURL(path, params, p.Page+1)
}
