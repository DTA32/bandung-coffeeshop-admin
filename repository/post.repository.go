package repository

import (
	"context"
	"fmt"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostRepository owns Instagram post proposals (change_proposal rows with
// field ig_post).
type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) ListPending(ctx context.Context, cafeID int) ([]models.Proposal, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+proposalColumns+`
		FROM change_proposal p
		JOIN cafe c ON c.id = p.cafe_id
		JOIN location l ON l.id = c.location_id
		WHERE p.status = 'pending' AND p.field = 'ig_post'
		  AND ($1 = 0 OR p.cafe_id = $1)
		ORDER BY p.new_value->>'taken_at' DESC NULLS LAST, p.id DESC
	`, cafeID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanProposal)
}

// Get returns one proposal by id (the controller checks that it is an ig_post).
func (r *PostRepository) Get(ctx context.Context, id int64) (*models.Proposal, error) {
	return getProposal(ctx, r.db, id)
}

// Ignore rejects a pending post proposal; the scraper never re-proposes it.
func (r *PostRepository) Ignore(ctx context.Context, id int64) error {
	return rejectProposal(ctx, r.db, id)
}

// CreateFeed inserts the feed entry and marks the ig_post proposal
// approved with {"cafe_feed_id": n}, in one transaction.
func (r *PostRepository) CreateFeed(ctx context.Context, proposalID int64, f models.Feed) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	cafeID, field, err := lockPending(ctx, tx, proposalID)
	if err != nil {
		return 0, err
	}
	if field != models.FieldIGPost {
		return 0, fmt.Errorf("proposal %d is not an instagram post", proposalID)
	}
	f.CafeID = cafeID
	id, err := insertFeed(ctx, tx, f)
	if err != nil {
		return 0, err
	}
	if err := resolve(ctx, tx, proposalID, models.StatusApproved, map[string]int{"cafe_feed_id": id}); err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}
