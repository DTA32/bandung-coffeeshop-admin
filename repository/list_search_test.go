package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestCafeListSearchAndPaging(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()
	var locID string
	if err := pool.QueryRow(ctx, `SELECT location_id FROM cafe WHERE id = $1`, cafeID).Scan(&locID); err != nil {
		t.Fatal(err)
	}

	// by slug, case-insensitive
	pg, err := repo.Cafe.List(ctx, strings.ToUpper(locID), models.CafeSortUpdatedAsc, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(pg.Items) != 1 || pg.Items[0].ID != locID || pg.HasMore {
		t.Fatalf("slug search: %+v", pg)
	}
	// by name ("Test") finds it as well
	pg, err = repo.Cafe.List(ctx, "tEsT", models.CafeSortUpdatedDesc, 1)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, it := range pg.Items {
		found = found || it.ID == locID
	}
	if !found && !pg.HasMore {
		t.Fatalf("name search missed the test cafe: %+v", pg)
	}
	// wildcards are literal: "%" / "_" only match names or ids that contain them
	for _, q := range []string{"%", "_", `\`} {
		pg, err = repo.Cafe.List(ctx, q, models.CafeSortUpdatedAsc, 1)
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range pg.Items {
			if !strings.Contains(it.Name+it.ID, q) {
				t.Fatalf("q=%q matched %q/%q", q, it.Name, it.ID)
			}
		}
	}
	// paging: pages are disjoint and HasMore agrees with the next page
	p1, err := repo.Cafe.List(ctx, "", models.CafeSortUpdatedAsc, 1)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := repo.Cafe.List(ctx, "", models.CafeSortUpdatedAsc, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(p1.Items) > PageSize || p1.HasMore != (len(p2.Items) > 0) {
		t.Fatalf("paging: p1=%d hasMore=%v p2=%d", len(p1.Items), p1.HasMore, len(p2.Items))
	}
	seen := map[string]bool{}
	for _, it := range p1.Items {
		seen[it.ID] = true
	}
	for _, it := range p2.Items {
		if seen[it.ID] {
			t.Fatalf("%s on both pages", it.ID)
		}
	}
	opts, err := repo.Cafe.Options(ctx)
	if err != nil || len(opts) < len(p1.Items) {
		t.Fatalf("options: %d, %v", len(opts), err)
	}
}

func TestLocationListSearchAndPaging(t *testing.T) {
	repo, _, _ := testRepo(t)
	ctx := context.Background()
	p1, err := repo.Location.List(ctx, "", "", 1)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := repo.Location.List(ctx, "", "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(p1.Items) > PageSize || p1.HasMore != (len(p2.Items) > 0) {
		t.Fatalf("paging: p1=%d hasMore=%v p2=%d", len(p1.Items), p1.HasMore, len(p2.Items))
	}
	for _, q := range []string{"%", "_"} {
		pg, err := repo.Location.List(ctx, "", q, 1)
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range pg.Items {
			if !strings.Contains(it.Name+it.ID, q) {
				t.Fatalf("q=%q matched %q/%q", q, it.Name, it.ID)
			}
		}
	}
	if len(p1.Items) > 0 {
		first := p1.Items[0]
		pg, err := repo.Location.List(ctx, first.Type, strings.ToUpper(first.ID), 1)
		if err != nil || len(pg.Items) == 0 {
			t.Fatalf("id search: %+v %v", pg, err)
		}
	}
}
