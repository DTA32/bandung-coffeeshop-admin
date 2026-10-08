package controllers

import (
	"encoding/json"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestPostImageURLsAndDecode(t *testing.T) {
	p := models.Proposal{ID: 7, Field: models.FieldIGPost, EvidenceURL: ptr("https://www.instagram.com/p/A/"),
		NewValue: json.RawMessage(`{"caption":"c","taken_at":"2026-10-01T04:45:15Z","media":[{"type":"image","r2_key":"ig/x/A/0.jpg"},{"type":"video","r2_key":"/ig/x/A/1.jpg"},{"type":"image","r2_key":""}]}`)}
	post, err := DecodePost(p)
	if err != nil {
		t.Fatal(err)
	}
	if post.Value.Permalink != "https://www.instagram.com/p/A/" {
		t.Fatalf("permalink should fall back to evidence_url")
	}
	got := PostImageURLs(post, "https://image.bdgcafe.com")
	if len(got) != 2 || got[0] != "https://image.bdgcafe.com/ig/x/A/0.jpg" || got[1] != "https://image.bdgcafe.com/ig/x/A/1.jpg" {
		t.Fatalf("urls: %v", got)
	}
}
