package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestApplyFieldProposal(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()

	// Live rows: the created every-day row, plus special dates before, inside
	// and after the window 2026-10-24 .. 2026-10-30.
	if _, err := pool.Exec(ctx, `
		INSERT INTO cafe_hour (cafe_id, valid_from, valid_until, close) VALUES
			($1, '2026-10-20', '2026-10-20', true),
			($1, '2026-10-25', '2026-10-25', true),
			($1, '2026-10-31', '2026-11-02', true)`, cafeID); err != nil {
		t.Fatal(err)
	}
	// Backdate updated_at so the bump is observable.
	_, _ = pool.Exec(ctx, `UPDATE cafe SET updated_at = now() - interval '1 day' WHERE id = $1`, cafeID)
	hours := models.HoursProposal{WindowFrom: "2026-10-24", WindowUntil: "2026-10-30", Rows: []models.CafeHour{
		{Day: strp("mon"), Closed: true},
		{Open: strp("09:00"), Close: strp("21:00")},
		{ValidFrom: strp("2026-10-28"), ValidUntil: strp("2026-10-28"), Open: strp("10:00"), Close: strp("12:00")},
	}}
	id := insertProposal(t, pool, cafeID, models.FieldHours, "", `{"window_from":"2026-10-24","window_until":"2026-10-30","rows":[]}`)
	if err := repo.Change.Apply(ctx, id, models.FieldUpdate{Hours: &hours}, hours); err != nil {
		t.Fatal(err)
	}
	live, err := listHours(ctx, pool, cafeID, false)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, h := range live {
		s := ""
		if h.Day != nil {
			s += *h.Day
		}
		if h.ValidFrom != nil {
			s += *h.ValidFrom
		}
		if h.Open != nil {
			s += " " + *h.Open
		}
		got = append(got, s)
	}
	// Kept: 10-20 and 10-31 (outside the window). Replaced: the every-day row
	// and 10-25 (inside).
	want := " 09:00|mon|2026-10-20|2026-10-28 10:00|2026-10-31"
	if strings.Join(got, "|") != want {
		t.Fatalf("rows after apply: %q, want %q", strings.Join(got, "|"), want)
	}
	var fresh bool
	if err := pool.QueryRow(ctx, `SELECT updated_at > now() - interval '1 minute' FROM cafe WHERE id = $1`, cafeID).Scan(&fresh); err != nil {
		t.Fatal(err)
	}
	if !fresh {
		t.Fatal("cafe.updated_at not bumped")
	}

	// Second approval of the same proposal must fail and change nothing.
	err = repo.Change.Apply(ctx, id, models.FieldUpdate{Hours: &models.HoursProposal{WindowFrom: "2026-10-24", WindowUntil: "2026-10-30"}}, nil)
	if !errors.Is(err, ErrProposalNotPending) {
		t.Fatalf("want ErrProposalNotPending, got %v", err)
	}

	// Address writes location.description and bumps both updated_at columns.
	addr := insertProposal(t, pool, cafeID, models.FieldAddress, "", `"Jl. Test 1"`)
	if err := repo.Change.Apply(ctx, addr, models.FieldUpdate{Text: strp("Jl. Test 1")}, "Jl. Test 1"); err != nil {
		t.Fatal(err)
	}
	var desc string
	if err := pool.QueryRow(ctx, `
		SELECT l.description FROM location l JOIN cafe c ON c.location_id = l.id WHERE c.id = $1`, cafeID).Scan(&desc); err != nil {
		t.Fatal(err)
	}
	if desc != "Jl. Test 1" {
		t.Fatalf("description = %q", desc)
	}

	// An informational field cannot be applied; the tx must roll back.
	bio := insertProposal(t, pool, cafeID, models.FieldIGBio, "", `"bio"`)
	if err := repo.Change.Apply(ctx, bio, models.FieldUpdate{Text: strp("bio")}, "bio"); err == nil {
		t.Fatal("ig_bio must not be applicable")
	}
	p, err := repo.Change.Get(ctx, bio)
	if err != nil || p.Status != models.StatusPending {
		t.Fatalf("ig_bio should still be pending: %+v %v", p, err)
	}
}

func TestPendingUniqueness(t *testing.T) {
	_, pool, cafeID := testRepo(t)
	insertProposal(t, pool, cafeID, models.FieldPhone, "", `"+621"`)
	_, err := pool.Exec(context.Background(), `
		INSERT INTO change_proposal (cafe_id, source, field, new_value) VALUES ($1, 'gmaps', 'phone', '"+622"')`, cafeID)
	if err == nil {
		t.Fatal("a second pending proposal for the same field must violate uq_change_proposal_pending")
	}
}
