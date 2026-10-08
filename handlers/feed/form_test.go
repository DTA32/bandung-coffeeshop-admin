package feed

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseFeedForm(t *testing.T) {
	in := ParseFeedForm(url.Values{
		"content":     {"Promo"},
		"image":       {"https://i/1.jpg"},
		"images_text": {"https://i/2.jpg\r\n\n  https://i/3.jpg  "},
		"visible":     {"1"},
	})
	if in.Content != "Promo" || !in.Visible || strings.Join(in.Images, ",") != "https://i/1.jpg,https://i/2.jpg,https://i/3.jpg" {
		t.Fatalf("%+v", in)
	}
	if ParseFeedForm(url.Values{"content": {"x"}}).Visible {
		t.Fatal("unchecked visible must be false")
	}
}
