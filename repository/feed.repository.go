package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FeedRepository owns cafe_feed.
type FeedRepository struct {
	db *pgxpool.Pool
}

func NewFeedRepository(db *pgxpool.Pool) *FeedRepository {
	return &FeedRepository{db: db}
}

type execQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func insertFeed(ctx context.Context, q execQuerier, f models.Feed) (int, error) {
	var id int
	err := q.QueryRow(ctx, `
		INSERT INTO cafe_feed (cafe_id, content, images, author, link, published_at, expires_at, visible)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, f.CafeID, f.Content, f.Images, f.Author, f.Link, f.PublishedAt.UTC(), utcPtr(f.ExpiresAt), f.Visible).Scan(&id)
	return id, err
}

func utcPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}

func (r *FeedRepository) Create(ctx context.Context, f models.Feed) (int, error) {
	return insertFeed(ctx, r.db, f)
}

const feedColumns = `id, cafe_id, content, images, author, link, published_at, expires_at, visible, created_at, updated_at`

func scanFeed(row pgx.CollectableRow) (models.Feed, error) {
	var f models.Feed
	err := row.Scan(&f.ID, &f.CafeID, &f.Content, &f.Images, &f.Author, &f.Link,
		&f.PublishedAt, &f.ExpiresAt, &f.Visible, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

// List returns one page (1-based) of a cafe's feed entries, newest first.
func (r *FeedRepository) List(ctx context.Context, cafeID, page int) (models.Page[models.Feed], error) {
	limit, offset := pageWindow(page)
	rows, err := r.db.Query(ctx, `
		SELECT `+feedColumns+` FROM cafe_feed WHERE cafe_id = $1
		ORDER BY published_at DESC, id DESC LIMIT $2 OFFSET $3
	`, cafeID, limit, offset)
	if err != nil {
		return models.Page[models.Feed]{}, err
	}
	items, err := pgx.CollectRows(rows, scanFeed)
	return toPage(items), err
}

func (r *FeedRepository) Get(ctx context.Context, cafeID, id int) (*models.Feed, error) {
	rows, err := r.db.Query(ctx, `SELECT `+feedColumns+` FROM cafe_feed WHERE id = $1 AND cafe_id = $2`, id, cafeID)
	if err != nil {
		return nil, err
	}
	f, err := pgx.CollectExactlyOneRow(rows, scanFeed)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &f, err
}

func (r *FeedRepository) Update(ctx context.Context, f models.Feed) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE cafe_feed SET content = $3, images = $4, author = $5, link = $6,
			published_at = $7, expires_at = $8, visible = $9, updated_at = NOW()
		WHERE id = $1 AND cafe_id = $2
	`, f.ID, f.CafeID, f.Content, f.Images, f.Author, f.Link, f.PublishedAt.UTC(), utcPtr(f.ExpiresAt), f.Visible)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *FeedRepository) SetVisible(ctx context.Context, cafeID, id int, visible bool) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE cafe_feed SET visible = $3, updated_at = NOW() WHERE id = $1 AND cafe_id = $2
	`, id, cafeID, visible)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
