package repository

import (
	"context"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
)

// cafe_hour (BE migration 013) belongs to the cafe aggregate: the cafe form
// replaces every row (CafeRepository.Create/Update) and an approved hours
// proposal replaces the rows inside its window (ChangeRepository.Apply).

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

// replaceAllHours swaps every cafe_hour row of a cafe for rows (cafe form).
func replaceAllHours(ctx context.Context, tx pgx.Tx, cafeID int, rows []models.CafeHour) error {
	if _, err := tx.Exec(ctx, `DELETE FROM cafe_hour WHERE cafe_id = $1`, cafeID); err != nil {
		return err
	}
	return insertHours(ctx, tx, cafeID, rows)
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
