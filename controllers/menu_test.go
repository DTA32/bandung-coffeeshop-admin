package controllers

import "testing"

func TestIsMenuValue(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"https://example.com/menu.pdf", true},
		{"http://example.com", true},
		{"#image-1", true},
		{"#image-12", true},
		{"#image-0", false},
		{"#image-", false},
		{"#image-01", false},
		{"#img-1", false},
		{"#image-1x", false},
		{" #image-1", false},
		{"javascript:alert(1)", false},
		{"drive.google.com/x", false},
		{"plain text", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isMenuValue(tc.in); got != tc.want {
			t.Errorf("isMenuValue(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
