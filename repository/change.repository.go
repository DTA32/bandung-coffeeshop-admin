package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChangeRepository owns field change proposals (change_proposal rows other
// than ig_post, which PostRepository owns). The proposal helpers below are
// shared by both because they are the same table.
type ChangeRepository struct {
	db *pgxpool.Pool
}

func NewChangeRepository(db *pgxpool.Pool) *ChangeRepository {
	return &ChangeRepository{db: db}
}

// currentValueSQL renders the live value of a proposal's field as jsonb, so
// the queue can show "now" next to "proposed". For hours it is the cafe's
// cafe_hour rows in the proposal row shape: {"rows": [...]} (models.CurrentHours).
const currentValueSQL = `
	CASE p.field
		WHEN 'hours'   THEN jsonb_build_object('rows', (
			SELECT COALESCE(jsonb_agg(jsonb_build_object(
				'day', h.day::text,
				'open', TO_CHAR(h.open_hour, 'HH24:MI'), 'close', TO_CHAR(h.close_hour, 'HH24:MI'),
				'valid_from', TO_CHAR(h.valid_from, 'YYYY-MM-DD'), 'valid_until', TO_CHAR(h.valid_until, 'YYYY-MM-DD'),
				'closed', h.close)
				ORDER BY h.valid_from NULLS FIRST, h.valid_until NULLS FIRST, h.day NULLS FIRST, h.close DESC, h.open_hour, h.id),
				'[]'::jsonb)
			FROM cafe_hour h WHERE h.cafe_id = c.id))
		WHEN 'phone'   THEN to_jsonb(c.phone)
		WHEN 'menu'    THEN to_jsonb(c.menu)
		WHEN 'address' THEN to_jsonb(l.description)
		WHEN 'status'  THEN to_jsonb(l.status::text)
	END`

const proposalColumns = `
	p.id, p.cafe_id, l.id, l.name, p.source::text, p.field, p.item_key,
	p.old_value, p.new_value, p.evidence_url, p.snapshot_id, p.status::text, p.created_at,
	` + currentValueSQL

func scanProposal(row pgx.CollectableRow) (models.Proposal, error) {
	var p models.Proposal
	var oldV, newV, cur []byte
	err := row.Scan(&p.ID, &p.CafeID, &p.LocationID, &p.CafeName, &p.Source, &p.Field, &p.ItemKey,
		&oldV, &newV, &p.EvidenceURL, &p.SnapshotID, &p.Status, &p.CreatedAt, &cur)
	p.OldValue, p.NewValue, p.Current = oldV, newV, cur
	return p, err
}

// ListPending returns pending non-post proposals, optionally for one
// cafe (cafeID 0 = all), ordered by cafe name.
func (r *ChangeRepository) ListPending(ctx context.Context, cafeID int) ([]models.Proposal, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+proposalColumns+`
		FROM change_proposal p
		JOIN cafe c ON c.id = p.cafe_id
		JOIN location l ON l.id = c.location_id
		WHERE p.status = 'pending' AND p.field <> 'ig_post'
		  AND ($1 = 0 OR p.cafe_id = $1)
		ORDER BY l.name, p.cafe_id, p.field, p.created_at
	`, cafeID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanProposal)
}

// ListPendingPage returns the pending non-post proposals of PageSize *cafes*
// (ordered by name, id) after the cafe afterCafeID (0 = from the start), so a
// cafe's group is never split across pages. It pages by cursor, not offset:
// approving rows already on screen removes cafes from the queue and would make
// an offset skip unseen ones. hasMore reports whether more cafes follow.
func (r *ChangeRepository) ListPendingPage(ctx context.Context, cafeID, afterCafeID int) (items []models.Proposal, hasMore bool, err error) {
	idRows, err := r.db.Query(ctx, `
		SELECT c.id
		FROM cafe c
		JOIN location l ON l.id = c.location_id
		WHERE ($1 = 0 OR c.id = $1)
		  AND EXISTS (SELECT 1 FROM change_proposal p
		              WHERE p.cafe_id = c.id AND p.status = 'pending' AND p.field <> 'ig_post')
		  AND ($3 = 0 OR (l.name, c.id) > (SELECT l2.name, c2.id FROM cafe c2
		                                   JOIN location l2 ON l2.id = c2.location_id WHERE c2.id = $3))
		ORDER BY l.name, c.id
		LIMIT $2
	`, cafeID, PageSize+1, afterCafeID)
	if err != nil {
		return nil, false, err
	}
	ids, err := pgx.CollectRows(idRows, pgx.RowTo[int])
	if err != nil {
		return nil, false, err
	}
	if len(ids) > PageSize {
		ids, hasMore = ids[:PageSize], true
	}
	if len(ids) == 0 {
		return nil, false, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT `+proposalColumns+`
		FROM change_proposal p
		JOIN cafe c ON c.id = p.cafe_id
		JOIN location l ON l.id = c.location_id
		WHERE p.status = 'pending' AND p.field <> 'ig_post' AND p.cafe_id = ANY($1)
		ORDER BY l.name, p.cafe_id, p.field, p.created_at, p.id
	`, ids)
	if err != nil {
		return nil, false, err
	}
	items, err = pgx.CollectRows(rows, scanProposal)
	return items, hasMore, err
}

func (r *ChangeRepository) Get(ctx context.Context, id int64) (*models.Proposal, error) {
	return getProposal(ctx, r.db, id)
}

// getProposal is shared with PostRepository (same table).
func getProposal(ctx context.Context, db *pgxpool.Pool, id int64) (*models.Proposal, error) {
	rows, err := db.Query(ctx, `
		SELECT `+proposalColumns+`
		FROM change_proposal p
		JOIN cafe c ON c.id = p.cafe_id
		JOIN location l ON l.id = c.location_id
		WHERE p.id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	p, err := pgx.CollectExactlyOneRow(rows, scanProposal)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

// lockPending locks a pending proposal row inside tx and returns its cafe id
// and field.
func lockPending(ctx context.Context, tx pgx.Tx, id int64) (cafeID int, field string, err error) {
	var status string
	err = tx.QueryRow(ctx, `
		SELECT cafe_id, field, status::text FROM change_proposal WHERE id = $1 FOR UPDATE
	`, id).Scan(&cafeID, &field, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", ErrNotFound
	}
	if err != nil {
		return 0, "", err
	}
	if status != models.StatusPending {
		return 0, "", ErrProposalNotPending
	}
	return cafeID, field, nil
}

func resolve(ctx context.Context, tx pgx.Tx, id int64, status string, resolved any) error {
	var val []byte
	if resolved != nil {
		b, err := json.Marshal(resolved)
		if err != nil {
			return err
		}
		val = b
	}
	_, err := tx.Exec(ctx, `
		UPDATE change_proposal
		SET status = $2::proposal_status, resolved_value = $3, resolved_at = NOW()
		WHERE id = $1
	`, id, status, val)
	return err
}

// Apply writes the value to the live tables, marks the proposal approved and
// bumps cafe.updated_at (and location.updated_at when a location column
// changed), all in one transaction. Hours replace the undated cafe_hour rows
// and the dated rows overlapping the proposal's window (replaceWindowHours).
func (r *ChangeRepository) Apply(ctx context.Context, id int64, u models.FieldUpdate, resolved any) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cafeID, field, err := lockPending(ctx, tx, id)
	if err != nil {
		return err
	}

	switch field {
	case models.FieldHours:
		if u.Hours == nil {
			return errors.New("hours update without rows")
		}
		// Bumping updated_at first also locks the cafe row against a
		// concurrent cafe form save.
		if _, err = tx.Exec(ctx, `UPDATE cafe SET updated_at = NOW() WHERE id = $1`, cafeID); err == nil {
			err = replaceWindowHours(ctx, tx, cafeID, u.Hours.WindowFrom, u.Hours.WindowUntil, u.Hours.Rows)
		}
	case models.FieldPhone:
		_, err = tx.Exec(ctx, `UPDATE cafe SET phone = $2, updated_at = NOW() WHERE id = $1`, cafeID, u.Text)
	case models.FieldMenu:
		_, err = tx.Exec(ctx, `UPDATE cafe SET menu = $2, updated_at = NOW() WHERE id = $1`, cafeID, u.Text)
	case models.FieldAddress, models.FieldStatus:
		column := "description = $2"
		if field == models.FieldStatus {
			column = "status = $2::location_status_enum"
		}
		if _, err = tx.Exec(ctx, `
			UPDATE location SET `+column+`, updated_at = NOW()
			WHERE id = (SELECT location_id FROM cafe WHERE id = $1)
		`, cafeID, u.Text); err == nil {
			_, err = tx.Exec(ctx, `UPDATE cafe SET updated_at = NOW() WHERE id = $1`, cafeID)
		}
	default:
		return fmt.Errorf("field %q cannot be applied", field)
	}
	if err != nil {
		return err
	}
	if err := resolve(ctx, tx, id, models.StatusApproved, resolved); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Reject rejects (or, for informational fields, dismisses) a pending proposal.
// The scraper never re-proposes a rejected value.
func (r *ChangeRepository) Reject(ctx context.Context, id int64) error {
	return rejectProposal(ctx, r.db, id)
}

// rejectProposal is shared with PostRepository.Ignore (same table).
func rejectProposal(ctx context.Context, db *pgxpool.Pool, id int64) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, _, err := lockPending(ctx, tx, id); err != nil {
		return err
	}
	if err := resolve(ctx, tx, id, models.StatusRejected, nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
