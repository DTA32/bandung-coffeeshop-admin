package utils

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParseListParams(t *testing.T) {
	cases := []struct {
		name, method, url string
		hdr               map[string]string
		want              ListParams
	}{
		{"plain", "GET", "/cafes", nil, ListParams{Page: 1}},
		{"q trimmed", "GET", "/cafes?q=%20dago%20", nil, ListParams{Page: 1, Q: "dago"}},
		{"bad page", "GET", "/cafes?page=abc", nil, ListParams{Page: 1}},
		{"zero page", "GET", "/cafes?page=0", nil, ListParams{Page: 1}},
		{"huge page", "GET", "/cafes?page=999999999", nil, ListParams{Page: MaxPage}},
		{"full page with page>=2 but not htmx", "GET", "/cafes?page=2", nil, ListParams{Page: 2}},
		{"sentinel", "GET", "/cafes?page=2", map[string]string{"HX-Request": "true"}, ListParams{Page: 2, Fragment: true}},
		{"typeahead", "GET", "/cafes?q=x", map[string]string{"HX-Request": "true", "HX-Target": "cafe-list"}, ListParams{Page: 1, Q: "x", Fragment: true}},
		{"other target", "GET", "/cafes?q=x", map[string]string{"HX-Request": "true", "HX-Target": "body"}, ListParams{Page: 1, Q: "x"}},
		{"history restore", "GET", "/cafes?page=2", map[string]string{"HX-Request": "true", "HX-History-Restore-Request": "true"}, ListParams{Page: 2}},
		{"post is never a fragment", "POST", "/cafes?page=2", map[string]string{"HX-Request": "true"}, ListParams{Page: 2}},
	}
	for _, c := range cases {
		r := httptest.NewRequest(c.method, c.url, nil)
		for k, v := range c.hdr {
			r.Header.Set(k, v)
		}
		if got := ParseListParams(r, "cafe-list"); got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestCleanQuery(t *testing.T) {
	if got := CleanQuery("a\x00b\n c"); got != "ab c" {
		t.Errorf("control chars: %q", got)
	}
	if got := CleanQuery(strings.Repeat("é", 300)); len([]rune(got)) != MaxQueryRunes {
		t.Errorf("cap: %d runes", len([]rune(got)))
	}
	if got := CleanQuery("a\xffb"); got != "ab" {
		t.Errorf("invalid utf8: %q", got)
	}
}

func TestNextURL(t *testing.T) {
	v := url.Values{"q": {"a&b c"}, "sort": {"updated_desc"}, "type": {""}, "page": {"9"}}
	got := NextURL("/cafes", v, 3)
	want := "/cafes?page=3&q=a%26b+c&sort=updated_desc"
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
	p := ListParams{Page: 2}
	if p.Next("/x", nil, false) != "" {
		t.Error("no more -> empty")
	}
	if p.Next("/x", nil, true) != "/x?page=3" {
		t.Error("next page")
	}
}
