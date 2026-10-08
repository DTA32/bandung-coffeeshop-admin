package repository

import (
	"context"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) List(ctx context.Context) ([]models.Tag, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, name_indo, slug, description, description_indo
		FROM tag ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanTag)
}

func (r *TagRepository) Get(ctx context.Context, id int) (*models.Tag, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, name_indo, slug, description, description_indo FROM tag WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	t, err := pgx.CollectExactlyOneRow(rows, scanTag)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return &t, err
}

func scanTag(row pgx.CollectableRow) (models.Tag, error) {
	var t models.Tag
	err := row.Scan(&t.ID, &t.Name, &t.NameIndo, &t.Slug, &t.Description, &t.DescriptionIndo)
	return t, err
}

func (r *TagRepository) Create(ctx context.Context, t models.Tag) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `
		INSERT INTO tag (name, name_indo, slug, description, description_indo)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, t.Name, t.NameIndo, t.Slug, t.Description, t.DescriptionIndo).Scan(&id)
	return id, err
}

func (r *TagRepository) Update(ctx context.Context, id int, t models.Tag) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE tag SET name = $2, name_indo = $3, slug = $4,
			description = $5, description_indo = $6, updated_at = NOW()
		WHERE id = $1
	`, id, t.Name, t.NameIndo, t.Slug, t.Description, t.DescriptionIndo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a tag and its cafe assignments in one transaction.
func (r *TagRepository) Delete(ctx context.Context, id int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM cafe_tag WHERE tag_id = $1`, id); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM tag WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}
