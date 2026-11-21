package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dta32/bandung-coffeeshop-admin/models"
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

func TestCafeUpdateReplacesHours(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()
	var locID string
	if err := pool.QueryRow(ctx, `SELECT location_id FROM cafe WHERE id = $1`, cafeID).Scan(&locID); err != nil {
		t.Fatal(err)
	}
	c, err := repo.Cafe.Get(ctx, locID)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Hours) != 1 || *c.Hours[0].Open != "08:00" || c.Hours[0].Day != nil {
		t.Fatalf("created hours: %+v", c.Hours)
	}
	c.Hours = []models.CafeHour{
		{Day: strp("mon"), Closed: true},
		{Day: strp("sat"), Open: strp("08:00"), Close: strp("14:00")},
		{ValidFrom: strp("2026-12-25"), ValidUntil: strp("2026-12-25"), Closed: true},
	}
	if err := repo.Cafe.Update(ctx, locID, *c); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Cafe.Get(ctx, locID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Hours) != 3 || *got.Hours[0].Day != "mon" || !got.Hours[0].Closed || *got.Hours[2].ValidFrom != "2026-12-25" {
		t.Fatalf("hours after update: %+v", got.Hours)
	}
}

// cafeStamps is every updated_at of a cafe aggregate plus its cafe_hour ids
// (in id order), to tell exactly which rows a save wrote.
type cafeStamps struct {
	Cafe, Location, Price time.Time
	Tags, Images, Hours   string // "id/key@updated_at,..." in key order
}

func readStamps(t *testing.T, pool *pgxpool.Pool, cafeID int) cafeStamps {
	t.Helper()
	var s cafeStamps
	if err := pool.QueryRow(context.Background(), `
		SELECT c.updated_at, l.updated_at, p.updated_at,
		       COALESCE((SELECT string_agg(tag_id || '@' || updated_at, ',' ORDER BY tag_id) FROM cafe_tag WHERE cafe_id = c.id), ''),
		       COALESCE((SELECT string_agg(url || '@' || updated_at, ',' ORDER BY url) FROM location_image WHERE location_id = l.id), ''),
		       COALESCE((SELECT string_agg(id || '@' || updated_at, ',' ORDER BY id) FROM cafe_hour WHERE cafe_id = c.id), '')
		FROM cafe c JOIN location l ON l.id = c.location_id
		LEFT JOIN cafe_price p ON p.cafe_id = c.id
		WHERE c.id = $1`, cafeID).Scan(&s.Cafe, &s.Location, &s.Price, &s.Tags, &s.Images, &s.Hours); err != nil {
		t.Fatal(err)
	}
	return s
}

// TestCafeUpdateWritesOnlyChanges: a no-op save writes nothing, and an edit
// bumps only the table that holds it (BCF-45).
func TestCafeUpdateWritesOnlyChanges(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()
	var locID string
	if err := pool.QueryRow(ctx, `SELECT location_id FROM cafe WHERE id = $1`, cafeID).Scan(&locID); err != nil {
		t.Fatal(err)
	}
	var tagID int
	if err := pool.QueryRow(ctx, `INSERT INTO tag (name) VALUES ('zz noop test') RETURNING id`).Scan(&tagID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM cafe_tag WHERE cafe_id = $1`, cafeID)
		_, _ = pool.Exec(ctx, `DELETE FROM tag WHERE id = $1`, tagID)
		_, _ = pool.Exec(ctx, `DELETE FROM location_image WHERE location_id = $1`, locID)
	})

	// Fill every child table, then backdate every timestamp.
	c, err := repo.Cafe.Get(ctx, locID)
	if err != nil {
		t.Fatal(err)
	}
	ten, twenty := 10, 20
	c.Price = models.CafePrice{PriceRangeMin: &ten, PriceRangeMax: &twenty}
	c.Phone = strp("+62811")
	c.Coordinates = &models.Coordinates{Lat: -6.9, Lng: 107.6}
	c.Tags = []models.CafeTag{{TagID: tagID, Visible: true}}
	c.Images = []models.Image{{URL: "https://example.invalid/a.jpg", Description: "a"}, {URL: "https://example.invalid/b.jpg"}}
	c.Hours = []models.CafeHour{
		{Open: strp("08:00"), Close: strp("22:00")},
		{Day: strp("mon"), Closed: true},
		{ValidFrom: strp("2026-12-25"), ValidUntil: strp("2026-12-25"), Closed: true},
	}
	if err := repo.Cafe.Update(ctx, locID, *c); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`UPDATE cafe SET updated_at = '2020-01-01' WHERE id = $1`,
		`UPDATE location SET updated_at = '2020-01-01' WHERE id = (SELECT location_id FROM cafe WHERE id = $1)`,
		`UPDATE cafe_price SET updated_at = '2020-01-01' WHERE cafe_id = $1`,
		`UPDATE cafe_tag SET updated_at = '2020-01-01' WHERE cafe_id = $1`,
		`UPDATE cafe_hour SET updated_at = '2020-01-01' WHERE cafe_id = $1`,
		`UPDATE location_image SET updated_at = '2020-01-01' WHERE location_id = (SELECT location_id FROM cafe WHERE id = $1)`,
	} {
		if _, err := pool.Exec(ctx, q, cafeID); err != nil {
			t.Fatal(err)
		}
	}
	before := readStamps(t, pool, cafeID)

	// 1. Saving exactly what Get returns (hours in a different order) writes nothing.
	c, err = repo.Cafe.Get(ctx, locID)
	if err != nil {
		t.Fatal(err)
	}
	c.Hours[0], c.Hours[2] = c.Hours[2], c.Hours[0]
	if err := repo.Cafe.Update(ctx, locID, *c); err != nil {
		t.Fatal(err)
	}
	if got := readStamps(t, pool, cafeID); got != before {
		t.Fatalf("no-op save changed rows:\nbefore %+v\nafter  %+v", before, got)
	}

	// 2. Each edit bumps only its own table.
	edits := []struct {
		name  string
		edit  func(c *models.Cafe)
		check func(b, a cafeStamps) bool // true = only the expected part moved
	}{
		{"phone", func(c *models.Cafe) { c.Phone = strp("+62812") }, func(b, a cafeStamps) bool {
			return a.Cafe.After(b.Cafe) && a.Location == b.Location && a.Price == b.Price && a.Hours == b.Hours && a.Images == b.Images && a.Tags == b.Tags
		}},
		{"name", func(c *models.Cafe) { c.Name = "Test renamed" }, func(b, a cafeStamps) bool {
			return a.Location.After(b.Location) && a.Cafe == b.Cafe && a.Price == b.Price && a.Hours == b.Hours && a.Images == b.Images
		}},
		{"price", func(c *models.Cafe) { c.Price.PriceRangeMax = &ten }, func(b, a cafeStamps) bool {
			return a.Price.After(b.Price) && a.Cafe == b.Cafe && a.Location == b.Location && a.Hours == b.Hours && a.Images == b.Images
		}},
		{"image description", func(c *models.Cafe) { c.Images[1].Description = "b" }, func(b, a cafeStamps) bool {
			return a.Images != b.Images && a.Cafe == b.Cafe && a.Location == b.Location && a.Price == b.Price && a.Hours == b.Hours
		}},
		{"one hour row", func(c *models.Cafe) {
			for i := range c.Hours {
				if c.Hours[i].Day != nil {
					c.Hours[i] = models.CafeHour{Day: strp("mon"), Open: strp("10:00"), Close: strp("14:00")}
				}
			}
		}, func(b, a cafeStamps) bool {
			return a.Hours != b.Hours && a.Cafe == b.Cafe && a.Location == b.Location && a.Price == b.Price && a.Images == b.Images
		}},
	}
	for _, e := range edits {
		b := readStamps(t, pool, cafeID)
		c, err := repo.Cafe.Get(ctx, locID)
		if err != nil {
			t.Fatal(err)
		}
		e.edit(c)
		if err := repo.Cafe.Update(ctx, locID, *c); err != nil {
			t.Fatal(err)
		}
		if a := readStamps(t, pool, cafeID); !e.check(b, a) {
			t.Fatalf("%s edit:\nbefore %+v\nafter  %+v", e.name, b, a)
		}
	}

	// The hour edit reused the live row: same three ids, only one stamp moved.
	after := readStamps(t, pool, cafeID)
	ids := func(s string) (out []string) {
		for _, p := range strings.Split(s, ",") {
			out = append(out, strings.SplitN(p, "@", 2)[0])
		}
		return out
	}
	if strings.Join(ids(after.Hours), ",") != strings.Join(ids(before.Hours), ",") {
		t.Fatalf("hour ids changed: %s -> %s", before.Hours, after.Hours)
	}
	if n := strings.Count(after.Hours, "@2020-01-01"); n != 2 {
		t.Fatalf("want 2 untouched hour rows, got %d: %s", n, after.Hours)
	}

	// 3. Removing a row deletes it; adding one inserts it.
	c, err = repo.Cafe.Get(ctx, locID)
	if err != nil {
		t.Fatal(err)
	}
	c.Hours = c.Hours[:2]
	if err := repo.Cafe.Update(ctx, locID, *c); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Cafe.Get(ctx, locID); len(got.Hours) != 2 {
		t.Fatalf("hours after removing one: %+v", got.Hours)
	}
}

func TestSoftDeleteIsIdempotent(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()
	var locID string
	if err := pool.QueryRow(ctx, `SELECT location_id FROM cafe WHERE id = $1`, cafeID).Scan(&locID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Cafe.SoftDelete(ctx, locID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE location SET updated_at = '2020-01-01' WHERE id = $1`, locID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Cafe.SoftDelete(ctx, locID); err != nil {
		t.Fatalf("second delete: %v", err)
	}
	var at time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM location WHERE id = $1`, locID).Scan(&at); err != nil {
		t.Fatal(err)
	}
	if !at.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("second delete bumped location.updated_at to %v", at)
	}
	if err := repo.Cafe.SoftDelete(ctx, "zz-no-such-cafe"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown cafe: err = %v, want ErrNotFound", err)
	}
}
