package repository

import (
	"context"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
)

// cafe_hour (BE migration 013) belongs to the cafe aggregate: the cafe form
// syncs every row (CafeRepository.Create/Update, syncHours) and an approved
// hours proposal replaces the rows inside its window (ChangeRepository.Apply).

type rowsQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// hourRowOrder is the display order: undated first, every day before
// mon→sun, then by open hour.
const hourRowOrder = `valid_from NULLS FIRST, valid_until NULLS FIRST, day NULLS FIRST, close DESC, open_hour, id`

// listHours returns a cafe's cafe_hour rows (times as HH:MM, dates as
// YYYY-MM-DD).
func listHours(ctx context.Context, q rowsQuerier, cafeID int, forUpdate bool) ([]models.CafeHour, error) {
	lock := ""
	if forUpdate {
		lock = " FOR UPDATE"
	}
	rows, err := q.Query(ctx, `
		SELECT id, day::text, TO_CHAR(open_hour, 'HH24:MI'), TO_CHAR(close_hour, 'HH24:MI'),
		       TO_CHAR(valid_from, 'YYYY-MM-DD'), TO_CHAR(valid_until, 'YYYY-MM-DD'), close
		FROM cafe_hour WHERE cafe_id = $1
		ORDER BY `+hourRowOrder+lock, cafeID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.CafeHour, error) {
		var h models.CafeHour
		err := row.Scan(&h.ID, &h.Day, &h.Open, &h.Close, &h.ValidFrom, &h.ValidUntil, &h.Closed)
		return h, err
	})
}

// insertHours inserts rows for a cafe (ids are ignored).
func insertHours(ctx context.Context, tx pgx.Tx, cafeID int, rows []models.CafeHour) error {
	for _, h := range rows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO cafe_hour (cafe_id, day, open_hour, close_hour, valid_from, valid_until, close)
			VALUES ($1, $2::day_enum, $3::time, $4::time, $5::date, $6::date, $7)
		`, cafeID, h.Day, h.Open, h.Close, h.ValidFrom, h.ValidUntil, h.Closed); err != nil {
			return err
		}
	}
	return nil
}

// hourKey is a row's content (everything but the id), for diffing.
func hourKey(h models.CafeHour) [6]string {
	v := func(p *string) string {
		if p == nil {
			return "\x00"
		}
		return *p
	}
	closed := "f"
	if h.Closed {
		closed = "t"
	}
	return [6]string{v(h.Day), v(h.Open), v(h.Close), v(h.ValidFrom), v(h.ValidUntil), closed}
}

// syncHours makes a cafe's cafe_hour rows equal rows (cafe form) while writing
// as little as possible: live rows identical to a submitted row stay as they
// are (same id, same timestamps); the remaining live and submitted rows are
// paired in display order and updated in place (bumping updated_at); surplus
// live rows are deleted and surplus submitted rows inserted. The form does not
// post row ids, so which live row a changed row reuses is arbitrary.
func syncHours(ctx context.Context, tx pgx.Tx, cafeID int, rows []models.CafeHour) error {
	live, err := listHours(ctx, tx, cafeID, true)
	if err != nil {
		return err
	}
	unmatched := map[[6]string][]int{} // content -> indexes into live
	for i, h := range live {
		k := hourKey(h)
		unmatched[k] = append(unmatched[k], i)
	}
	kept := make([]bool, len(live))
	var added []models.CafeHour
	for _, h := range rows {
		k := hourKey(h)
		if idx := unmatched[k]; len(idx) > 0 {
			kept[idx[0]], unmatched[k] = true, idx[1:]
			continue
		}
		added = append(added, h)
	}
	var stale []int // ids of live rows nothing matched
	for i, h := range live {
		if !kept[i] {
			stale = append(stale, h.ID)
		}
	}
	for len(stale) > 0 && len(added) > 0 {
		h := added[0]
		if _, err := tx.Exec(ctx, `
			UPDATE cafe_hour SET day = $2::day_enum, open_hour = $3::time, close_hour = $4::time,
				valid_from = $5::date, valid_until = $6::date, close = $7, updated_at = NOW()
			WHERE id = $1
		`, stale[0], h.Day, h.Open, h.Close, h.ValidFrom, h.ValidUntil, h.Closed); err != nil {
			return err
		}
		stale, added = stale[1:], added[1:]
	}
	if len(stale) > 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM cafe_hour WHERE cafe_id = $1 AND id = ANY($2::int[])`, cafeID, stale); err != nil {
			return err
		}
	}
	return insertHours(ctx, tx, cafeID, added)
}

// replaceWindowHours applies an approved hours proposal: it deletes the
// cafe's undated rows and the dated rows overlapping [from, until]
// (models.ReplacedByWindow), keeps dated rows outside the window, and inserts
// rows. The caller has locked the cafe row.
func replaceWindowHours(ctx context.Context, tx pgx.Tx, cafeID int, from, until string, rows []models.CafeHour) error {
	live, err := listHours(ctx, tx, cafeID, true)
	if err != nil {
		return err
	}
	ids := []int{}
	for _, h := range live {
		if models.ReplacedByWindow(h, from, until) {
			ids = append(ids, h.ID)
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM cafe_hour WHERE cafe_id = $1 AND id = ANY($2::int[])`, cafeID, ids); err != nil {
		return err
	}
	return insertHours(ctx, tx, cafeID, rows)
}
