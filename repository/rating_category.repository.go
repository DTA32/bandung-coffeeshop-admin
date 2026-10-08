package repository

import (
	"context"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RatingCategoryRepository struct {
	db *pgxpool.Pool
}

func NewRatingCategoryRepository(db *pgxpool.Pool) *RatingCategoryRepository {
	return &RatingCategoryRepository{db: db}
}

func (r *RatingCategoryRepository) List(ctx context.Context) ([]models.RatingCategory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, type, name, name_indo, slug,
		       short_description, short_description_indo,
		       long_description, long_description_indo,
		       lower_bound::float8, upper_bound::float8
		FROM rating_category
		ORDER BY type, lower_bound
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanRatingCategory)
}

func (r *RatingCategoryRepository) Get(ctx context.Context, id int) (*models.RatingCategory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, type, name, name_indo, slug,
		       short_description, short_description_indo,
		       long_description, long_description_indo,
		       lower_bound::float8, upper_bound::float8
		FROM rating_category WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	rc, err := pgx.CollectExactlyOneRow(rows, scanRatingCategory)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return &rc, err
}

func scanRatingCategory(row pgx.CollectableRow) (models.RatingCategory, error) {
	var rc models.RatingCategory
	err := row.Scan(&rc.ID, &rc.Type, &rc.Name, &rc.NameIndo, &rc.Slug,
		&rc.ShortDescription, &rc.ShortDescriptionIndo,
		&rc.LongDescription, &rc.LongDescriptionIndo,
		&rc.LowerBound, &rc.UpperBound)
	return rc, err
}

func (r *RatingCategoryRepository) Create(ctx context.Context, rc models.RatingCategory) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `
		INSERT INTO rating_category (type, name, name_indo, slug,
			short_description, short_description_indo,
			long_description, long_description_indo, lower_bound, upper_bound)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`, rc.Type, rc.Name, rc.NameIndo, rc.Slug,
		rc.ShortDescription, rc.ShortDescriptionIndo,
		rc.LongDescription, rc.LongDescriptionIndo, rc.LowerBound, rc.UpperBound).Scan(&id)
	return id, err
}

func (r *RatingCategoryRepository) Update(ctx context.Context, id int, rc models.RatingCategory) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE rating_category SET type = $2, name = $3,
			name_indo = $4, slug = $5, short_description = $6, short_description_indo = $7,
			long_description = $8, long_description_indo = $9,
			lower_bound = $10, upper_bound = $11, updated_at = NOW()
		WHERE id = $1
	`, id, rc.Type, rc.Name, rc.NameIndo, rc.Slug,
		rc.ShortDescription, rc.ShortDescriptionIndo,
		rc.LongDescription, rc.LongDescriptionIndo, rc.LowerBound, rc.UpperBound)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *RatingCategoryRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM rating_category WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
