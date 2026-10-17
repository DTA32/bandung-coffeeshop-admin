package repository

import (
	"context"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RatingTypeRepository owns rating_type_label (the rating dimensions).
type RatingTypeRepository struct {
	db *pgxpool.Pool
}

func NewRatingTypeRepository(db *pgxpool.Pool) *RatingTypeRepository {
	return &RatingTypeRepository{db: db}
}

// ListPage returns one page (1-based) of rating types ordered by type.
func (r *RatingTypeRepository) ListPage(ctx context.Context, page int) (models.Page[models.RatingType], error) {
	limit, offset := pageWindow(page)
	rows, err := r.db.Query(ctx, `SELECT type, label, label_indo FROM rating_type_label ORDER BY type ASC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return models.Page[models.RatingType]{}, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.RatingType, error) {
		var rt models.RatingType
		err := row.Scan(&rt.Type, &rt.Label, &rt.LabelIndo)
		return rt, err
	})
	return toPage(items), err
}

func (r *RatingTypeRepository) List(ctx context.Context) ([]models.RatingType, error) {
	rows, err := r.db.Query(ctx, `SELECT type, label, label_indo FROM rating_type_label ORDER BY type ASC`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.RatingType, error) {
		var rt models.RatingType
		err := row.Scan(&rt.Type, &rt.Label, &rt.LabelIndo)
		return rt, err
	})
}

func (r *RatingTypeRepository) Create(ctx context.Context, rt models.RatingType) error {
	tag, err := r.db.Exec(ctx, `
		INSERT INTO rating_type_label (type, label, label_indo) VALUES ($1, $2, $3)
		ON CONFLICT (type) DO NOTHING
	`, rt.Type, rt.Label, rt.LabelIndo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrIDExists
	}
	return nil
}

// Update edits the labels; the type key is immutable (referenced).
func (r *RatingTypeRepository) Update(ctx context.Context, typeKey string, rt models.RatingType) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE rating_type_label SET label = $2, label_indo = $3 WHERE type = $1
	`, typeKey, rt.Label, rt.LabelIndo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete refuses while any category or cafe rating references it.
func (r *RatingTypeRepository) Delete(ctx context.Context, typeKey string) error {
	var refs int
	if err := r.db.QueryRow(ctx, `
		SELECT (SELECT COUNT(*) FROM rating_category WHERE type = $1)
		     + (SELECT COUNT(*) FROM cafe_rating WHERE category_type = $1)
	`, typeKey).Scan(&refs); err != nil {
		return err
	}
	if refs > 0 {
		return ErrRatingTypeInUse
	}
	tag, err := r.db.Exec(ctx, `DELETE FROM rating_type_label WHERE type = $1`, typeKey)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *RatingTypeRepository) Bounds(ctx context.Context) (map[string]models.RatingBounds, error) {
	rows, err := r.db.Query(ctx, `
		SELECT rtl.type, MIN(rc.lower_bound)::float8, MAX(rc.upper_bound)::float8
		FROM rating_type_label rtl
		LEFT JOIN rating_category rc ON rc.type = rtl.type
		GROUP BY rtl.type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]models.RatingBounds)
	for rows.Next() {
		var t string
		var b models.RatingBounds
		if err := rows.Scan(&t, &b.Min, &b.Max); err != nil {
			return nil, err
		}
		out[t] = b
	}
	return out, rows.Err()
}
