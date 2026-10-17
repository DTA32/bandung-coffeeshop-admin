package repository

import (
	"context"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestFeedListPaging(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()
	for i := 0; i < PageSize+3; i++ {
		if _, err := pool.Exec(ctx, `
			INSERT INTO cafe_feed (cafe_id, content, images, author, published_at, visible)
			VALUES ($1, 'zz paging', '{}', 'test', NOW() - make_interval(secs => $2), true)`, cafeID, i); err != nil {
			t.Fatal(err)
		}
	}
	p1, err := repo.Feed.List(ctx, cafeID, 1)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := repo.Feed.List(ctx, cafeID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(p1.Items) != PageSize || !p1.HasMore || len(p2.Items) != 3 || p2.HasMore {
		t.Fatalf("pages: %d/%v, %d/%v", len(p1.Items), p1.HasMore, len(p2.Items), p2.HasMore)
	}
	if p1.Items[PageSize-1].ID == p2.Items[0].ID {
		t.Fatal("pages overlap")
	}
}

func TestPendingPagesGroupWholeCafes(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()
	insertProposal(t, pool, cafeID, "phone", "", `"+62811"`)
	insertProposal(t, pool, cafeID, "menu", "", `"https://example.com/m"`)

	items, more, err := repo.Change.ListPendingPage(ctx, cafeID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || more {
		t.Fatalf("items=%d more=%v", len(items), more)
	}
	items, more, err = repo.Change.ListPendingPage(ctx, cafeID, 2)
	if err != nil || len(items) != 0 || more {
		t.Fatalf("page 2: %d %v %v", len(items), more, err)
	}
	if _, err := repo.Post.ListPendingPage(ctx, cafeID, 1); err != nil {
		t.Fatal(err)
	}
}

func TestSimpleListPages(t *testing.T) {
	repo, _, cafeID := testRepo(t)
	ctx := context.Background()
	if pg, err := repo.Tag.ListPage(ctx, 1); err != nil || len(pg.Items) > PageSize {
		t.Fatalf("tags: %v", err)
	}
	if _, err := repo.RatingType.ListPage(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.RatingCategory.ListPage(ctx, 1); err != nil {
		t.Fatal(err)
	}
	f := models.TargetFilter{CafeID: cafeID}
	pg, err := repo.Source.ListTargetsPage(ctx, f, 1)
	if err != nil || len(pg.Items) > PageSize {
		t.Fatalf("targets: %v", err)
	}
	all, err := repo.Source.ListTargets(ctx, f)
	if err != nil || len(all) != len(pg.Items) {
		t.Fatalf("unpaged targets: %d vs %d, %v", len(all), len(pg.Items), err)
	}
}
