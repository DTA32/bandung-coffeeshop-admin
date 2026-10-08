package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestApplyFieldProposal(t *testing.T) {
	repo, pool, cafeID := testRepo(t)
	ctx := context.Background()

	// Backdate updated_at so the bump is observable.
	_, _ = pool.Exec(ctx, `UPDATE cafe SET updated_at = now() - interval '1 day' WHERE id = $1`, cafeID)
	id := insertProposal(t, pool, cafeID, models.FieldHours, "", `{"open":"09:00","close":"21:00"}`)
	if err := repo.Change.Apply(ctx, id, models.FieldUpdate{OpenHour: strp("09:00"), CloseHour: strp("21:00")},
		map[string]string{"open": "09:00", "close": "21:00"}); err != nil {
		t.Fatal(err)
	}
	var open, close string
	var fresh bool
	if err := pool.QueryRow(ctx, `
		SELECT TO_CHAR(open_hour,'HH24:MI'), TO_CHAR(close_hour,'HH24:MI'), updated_at > now() - interval '1 minute'
		FROM cafe WHERE id = $1`, cafeID).Scan(&open, &close, &fresh); err != nil {
		t.Fatal(err)
	}
	if open != "09:00" || close != "21:00" || !fresh {
		t.Fatalf("got %s-%s fresh=%v", open, close, fresh)
	}

	// Second approval of the same proposal must fail and change nothing.
	err := repo.Change.Apply(ctx, id, models.FieldUpdate{OpenHour: strp("01:00"), CloseHour: strp("02:00")}, nil)
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
