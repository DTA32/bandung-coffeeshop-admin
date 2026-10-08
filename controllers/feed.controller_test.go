package controllers

import (
	"testing"
	"time"
)

func TestBuildFeed(t *testing.T) {
	jkt, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)

	f, err := BuildFeed(FeedInput{Content: " hi ", Images: []string{"https://i/1", "https://i/1", " "}}, jkt, now)
	if err != nil {
		t.Fatal(err)
	}
	if f.Content != "hi" || f.Author != "Admin" || !f.PublishedAt.Equal(now) || f.ExpiresAt != nil || len(f.Images) != 1 || f.Link != nil {
		t.Fatalf("defaults: %+v", f)
	}

	f, err = BuildFeed(FeedInput{Content: "x", PublishedAt: "2026-10-04T21:00", ExpiresAt: "2026-10-06T23:59", Link: "https://www.instagram.com/p/A/"}, jkt, now)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 10, 4, 14, 0, 0, 0, time.UTC); !f.PublishedAt.Equal(want) {
		t.Fatalf("published_at must be converted from WIB to UTC: %v", f.PublishedAt)
	}
	if want := time.Date(2026, 10, 6, 16, 59, 0, 0, time.UTC); !f.ExpiresAt.Equal(want) {
		t.Fatalf("expires_at: %v", f.ExpiresAt)
	}

	for _, in := range []FeedInput{
		{Content: " "},
		{Content: "x", Link: "javascript:alert(1)"},
		{Content: "x", Images: []string{"data:image/png;base64,AAA"}},
		{Content: "x", PublishedAt: "yesterday"},
		{Content: "x", PublishedAt: "2026-10-05T10:00", ExpiresAt: "2026-10-05T10:00"},
	} {
		if _, err := BuildFeed(in, jkt, now); err == nil || !IsValidation(err) {
			t.Fatalf("%+v: expected validation error, got %v", in, err)
		}
	}
}
