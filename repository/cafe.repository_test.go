package repository

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTouchTagBumpsOnlyCafeTag(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()

	var locID string
	if err := pool.QueryRow(ctx, `SELECT location_id FROM cafe WHERE id = $1`, cafeID).Scan(&locID); err != nil {
		t.Fatal(err)
	}
	var tagID, otherTagID int
	for _, dst := range []*int{&tagID, &otherTagID} {
		if err := pool.QueryRow(ctx, `INSERT INTO tag (name) VALUES ('zz touch test') RETURNING id`).Scan(dst); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM cafe_tag WHERE cafe_id = $1`, cafeID)
		_, _ = pool.Exec(ctx, `DELETE FROM tag WHERE id IN ($1, $2)`, tagID, otherTagID)
	})
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, id := range []int{tagID, otherTagID} {
		if _, err := pool.Exec(ctx, `INSERT INTO cafe_tag (cafe_id, tag_id, visible, updated_at) VALUES ($1, $2, true, $3)`,
			cafeID, id, old); err != nil {
			t.Fatal(err)
		}
	}
	var cafeBefore time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM cafe WHERE id = $1`, cafeID).Scan(&cafeBefore); err != nil {
		t.Fatal(err)
	}

	got, err := repo.Cafe.TouchTag(ctx, locID, tagID)
	if err != nil {
		t.Fatal(err)
	}
	var touched, other, cafeAfter time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM cafe_tag WHERE cafe_id = $1 AND tag_id = $2`, cafeID, tagID).Scan(&touched); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM cafe_tag WHERE cafe_id = $1 AND tag_id = $2`, cafeID, otherTagID).Scan(&other); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM cafe WHERE id = $1`, cafeID).Scan(&cafeAfter); err != nil {
		t.Fatal(err)
	}
	if !touched.After(old) || !touched.Equal(got) {
		t.Fatalf("touched tag updated_at = %v, returned %v", touched, got)
	}
	if !other.Equal(old) {
		t.Fatalf("other tag was touched: %v", other)
	}
	if !cafeAfter.Equal(cafeBefore) {
		t.Fatalf("cafe.updated_at changed: %v -> %v", cafeBefore, cafeAfter)
	}

	if _, err := repo.Cafe.TouchTag(ctx, locID, 999999999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown tag: err = %v, want ErrNotFound", err)
	}
	if _, err := repo.Cafe.TouchTag(ctx, "zz-no-such-cafe", tagID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown cafe: err = %v, want ErrNotFound", err)
	}
}
