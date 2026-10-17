package repository

import "testing"

func TestEscapeLike(t *testing.T) {
	for in, want := range map[string]string{
		"":         "",
		"dago":     "dago",
		"100%":     `100\%`,
		"a_b":      `a\_b`,
		`c:\x`:     `c:\\x`,
		`%_\`:      `\%\_\\`,
		"kopi ‘é’": "kopi ‘é’",
	} {
		if got := EscapeLike(in); got != want {
			t.Errorf("EscapeLike(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPageWindowAndToPage(t *testing.T) {
	if l, o := pageWindow(1); l != PageSize+1 || o != 0 {
		t.Errorf("page 1: %d %d", l, o)
	}
	if _, o := pageWindow(3); o != 2*PageSize {
		t.Errorf("page 3 offset %d", o)
	}
	if _, o := pageWindow(-5); o != 0 {
		t.Errorf("negative page offset %d", o)
	}
	items := make([]int, PageSize+1)
	if p := toPage(items); len(p.Items) != PageSize || !p.HasMore {
		t.Errorf("full: %d %v", len(p.Items), p.HasMore)
	}
	if p := toPage(items[:PageSize]); len(p.Items) != PageSize || p.HasMore {
		t.Errorf("exact: %d %v", len(p.Items), p.HasMore)
	}
}
