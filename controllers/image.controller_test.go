package controllers

import (
	"context"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/storage"
)

func TestImageFileName(t *testing.T) {
	jpeg := imageTypes["image/jpeg"]
	cases := []struct {
		in   string
		exts []string
		want string
	}{
		{"IMG_20240513_174819.jpg", jpeg, "IMG_20240513_174819.jpg"},
		{"11DDC2A4_1_105_c.JPEG", jpeg, "11DDC2A4_1_105_c.jpeg"},
		{"Foto kopi (1).jpg", jpeg, "Foto-kopi-1.jpg"},
		{`C:\Users\me\menu.png`, imageTypes["image/png"], "menu.png"},
		{"looks-like.png", jpeg, "looks-like.jpg"},
		{"no-extension", imageTypes["image/webp"], "no-extension.webp"},
		{"???.jpg", jpeg, "image.jpg"},
		{"../../etc/passwd.jpg", jpeg, "passwd.jpg"},
	}
	for _, tc := range cases {
		if got := ImageFileName(tc.in, tc.exts); got != tc.want {
			t.Errorf("ImageFileName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWithSuffix(t *testing.T) {
	if got := withSuffix("map.png", 1); got != "map.png" {
		t.Errorf("got %q", got)
	}
	if got := withSuffix("map.png", 3); got != "map-3.png" {
		t.Errorf("got %q", got)
	}
}

func TestUploadValidation(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	ctx := context.Background()

	if _, err := NewImageController(nil, "").Upload(ctx, "cafe", "accio-coffee", "a.png", png); !IsValidation(err) {
		t.Fatalf("no store: got %v, want validation error", err)
	}

	// Validation runs before any request, so a store that can't reach R2 is fine.
	c := NewImageController(storage.NewR2("acct", "key", "secret", "bucket"), "https://image.bdgcafe.com")
	cases := []struct {
		name, folder, slug string
		body               []byte
	}{
		{"unknown folder", "ig", "accio-coffee", png},
		{"blank slug", "cafe", "", png},
		{"bad slug", "cafe", "../accio", png},
		{"empty file", "cafe", "accio-coffee", nil},
		{"too big", "cafe", "accio-coffee", make([]byte, MaxImageBytes+1)},
		{"not an image", "location", "dago", []byte("<html>hi</html>")},
	}
	for _, tc := range cases {
		if _, err := c.Upload(ctx, tc.folder, tc.slug, "a.png", tc.body); !IsValidation(err) {
			t.Errorf("%s: got %v, want validation error", tc.name, err)
		}
	}
}
