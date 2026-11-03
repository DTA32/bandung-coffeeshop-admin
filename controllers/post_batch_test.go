package controllers

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestParsePostBatch(t *testing.T) {
	a, ids, err := ParsePostBatch("publish", []string{"3", " 1 ", "3"})
	if err != nil || a != BatchPublish || !reflect.DeepEqual(ids, []int64{3, 1}) {
		t.Fatalf("ok case: %v %v %v", a, ids, err)
	}
	for name, c := range map[string]struct {
		action string
		ids    []string
	}{
		"change action": {"approve", []string{"1"}},
		"bad action":    {"delete", []string{"1"}},
		"no ids":        {"ignore", nil},
		"non-numeric":   {"ignore", []string{"x"}},
	} {
		if _, _, err := ParsePostBatch(c.action, c.ids); !IsValidation(err) {
			t.Errorf("%s: err = %v, want validation error", name, err)
		}
	}
	if _, _, err := ParseBatch("publish", []string{"1"}); !IsValidation(err) {
		t.Errorf("changes must not accept publish: %v", err)
	}
}

func post(id int64, caption, status string) *models.Post {
	return &models.Post{
		Proposal: models.Proposal{ID: id, Field: models.FieldIGPost, Status: status},
		Value:    models.PostValue{Caption: caption},
	}
}

func TestRunPostBatchPublishPartialFailure(t *testing.T) {
	posts := map[int64]*models.Post{
		1: post(1, "new menu", models.StatusPending),
		2: post(2, "  ", models.StatusPending),       // no caption
		3: post(3, "old", models.StatusRejected),     // stale
		4: post(4, "db fails", models.StatusPending), // publish errors
		5: post(5, "promo", models.StatusPending),
	}
	var published []int64
	res := runPostBatch(context.Background(), BatchPublish, []int64{1, 2, 3, 4, 5, 6}, postBatchSteps{
		get: func(_ context.Context, id int64) (*models.Post, error) {
			if p, ok := posts[id]; ok {
				return p, nil
			}
			return nil, ErrNotFound
		},
		publish: func(_ context.Context, p models.Post) error {
			if p.ID == 4 {
				return errors.New("db exploded")
			}
			published = append(published, p.ID)
			return nil
		},
		ignore: func(context.Context, int64) error { t.Fatal("ignore called on publish"); return nil },
	})
	if !reflect.DeepEqual(res.Done, []int64{1, 5}) || !reflect.DeepEqual(published, []int64{1, 5}) {
		t.Fatalf("done = %v published = %v", res.Done, published)
	}
	reasons := map[int64]string{}
	for _, f := range res.Failed {
		reasons[f.ID] = f.Reason
	}
	want := map[int64]string{
		2: "no caption (use Create feed to write one)", 3: "already resolved", 4: "unexpected error", 6: "not found",
	}
	if !reflect.DeepEqual(reasons, want) {
		t.Fatalf("failed = %v", reasons)
	}
	if got := BatchSummary(res); got[:11] != "Published 2" {
		t.Fatalf("summary = %q", got)
	}
}

func TestRunPostBatchIgnore(t *testing.T) {
	var ignored []int64
	res := runPostBatch(context.Background(), BatchIgnore, []int64{7, 8}, postBatchSteps{
		get: func(_ context.Context, id int64) (*models.Post, error) {
			return post(id, "", models.StatusPending), nil
		},
		ignore: func(_ context.Context, id int64) error {
			ignored = append(ignored, id)
			return nil
		},
	})
	// an empty caption only blocks publish
	if !reflect.DeepEqual(res.Done, []int64{7, 8}) || !reflect.DeepEqual(ignored, []int64{7, 8}) {
		t.Fatalf("done = %v ignored = %v", res.Done, ignored)
	}
	if got := BatchSummary(res); got != "Ignored 2" {
		t.Fatalf("summary = %q", got)
	}
}

func TestDefaultFeedInput(t *testing.T) {
	jkt := time.FixedZone("WIB", 7*3600)
	now := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)
	p := models.Post{Value: models.PostValue{
		Caption:   "  hello  ",
		Permalink: "https://instagram.com/p/x",
		TakenAt:   time.Date(2026, 10, 12, 4, 45, 15, 0, time.UTC),
		Media:     []models.PostMedia{{R2Key: "/ig/a.jpg"}, {R2Key: ""}, {R2Key: "ig/b.jpg"}},
	}}
	got := DefaultFeedInput(p, FeedDefaults{ImageBase: "https://img.test", TZ: jkt, Now: now})
	want := FeedInput{
		Content: "hello", Author: "Admin", Link: "https://instagram.com/p/x", PublishedAt: "2026-10-12T11:45",
		Images: []string{"https://img.test/ig/a.jpg", "https://img.test/ig/b.jpg"}, Visible: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
	if _, err := BuildFeed(got, jkt, now); err != nil {
		t.Fatalf("defaults must build: %v", err)
	}
	p.Value.TakenAt = time.Time{} // scraper didn't record it -> now
	if got := DefaultFeedInput(p, FeedDefaults{TZ: jkt, Now: now}); got.PublishedAt != "2026-10-19T16:00" {
		t.Fatalf("fallback published_at = %q", got.PublishedAt)
	}
}
