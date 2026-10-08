package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestCreateFeedFromPost(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()
	id := insertProposal(t, pool, cafeID, models.FieldIGPost, "DdTEST", `{"caption":"hi","media":[]}`)

	exp := time.Now().Add(48 * time.Hour)
	feedID, err := repo.Post.CreateFeed(ctx, id, models.Feed{
		Content: "hi", Images: []string{"https://image.bdgcafe.com/ig/x/0.jpg"}, Author: "Admin",
		Link: strp("https://www.instagram.com/p/DdTEST/"), PublishedAt: time.Now(), ExpiresAt: &exp, Visible: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	f, err := repo.Feed.Get(ctx, cafeID, feedID)
	if err != nil || f.Content != "hi" || len(f.Images) != 1 || f.ExpiresAt == nil {
		t.Fatalf("%+v %v", f, err)
	}
	var status, resolved string
	if err := pool.QueryRow(ctx, `SELECT status::text, resolved_value->>'cafe_feed_id' FROM change_proposal WHERE id = $1`, id).
		Scan(&status, &resolved); err != nil {
		t.Fatal(err)
	}
	if status != "approved" || resolved == "" {
		t.Fatalf("status=%s resolved=%s", status, resolved)
	}
	if _, err := repo.Post.CreateFeed(ctx, id, models.Feed{Content: "again", Images: []string{}, Author: "Admin", PublishedAt: time.Now()}); !errors.Is(err, ErrProposalNotPending) {
		t.Fatalf("want ErrProposalNotPending, got %v", err)
	}
	// The same shortcode can never be proposed twice.
	if _, err := pool.Exec(ctx, `
		INSERT INTO change_proposal (cafe_id, source, field, item_key, new_value) VALUES ($1, 'instagram', 'ig_post', 'DdTEST', '{}')`, cafeID); err == nil {
		t.Fatal("uq_change_proposal_post must reject a repeated shortcode")
	}
}
