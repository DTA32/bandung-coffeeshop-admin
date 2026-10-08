package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// Integration tests run against a real, migrated database (001–008) and only
// when ADMIN_TEST_DSN is set, e.g.
//
//	ADMIN_TEST_DSN='postgres://postgres:postgres@localhost:55432/bandung_coffeeshop?sslmode=disable' go test ./repository
func testRepo(t *testing.T) (*Repositories, *pgxpool.Pool, int) {
	dsn := os.Getenv("ADMIN_TEST_DSN")
	if dsn == "" {
		t.Skip("ADMIN_TEST_DSN not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	repo := New(pool)
	locID := "zz-test-" + time.Now().Format("150405.000000")
	locID = sanitize(locID)
	if err := repo.Cafe.Create(ctx, models.Cafe{ID: locID, Name: "Test", Status: "active", Weather: []string{},
		OpenHour: strp("08:00"), CloseHour: strp("22:00")}); err != nil {
		t.Fatal(err)
	}
	var cafeID int
	if err := pool.QueryRow(ctx, `SELECT id FROM cafe WHERE location_id = $1`, locID).Scan(&cafeID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, q := range []string{
			`DELETE FROM cafe_feed WHERE cafe_id = $1`,
			`DELETE FROM change_proposal WHERE cafe_id = $1`,
			`DELETE FROM cafe_price WHERE cafe_id = $1`,
			`DELETE FROM cafe WHERE id = $1`,
		} {
			_, _ = pool.Exec(ctx, q, cafeID)
		}
		_, _ = pool.Exec(ctx, `DELETE FROM location WHERE id = $1`, locID)
		pool.Close()
	})
	return repo, pool, cafeID
}

func sanitize(s string) string {
	out := []rune{}
	for _, r := range s {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'z' {
			out = append(out, r)
		} else {
			out = append(out, '-')
		}
	}
	return string(out)
}

func strp(s string) *string { return &s }

func insertProposal(t *testing.T, pool *pgxpool.Pool, cafeID int, field, itemKey, newValue string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO change_proposal (cafe_id, source, field, item_key, new_value)
		VALUES ($1, 'gmaps', $2, $3, $4::jsonb) RETURNING id
	`, cafeID, field, itemKey, newValue).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
