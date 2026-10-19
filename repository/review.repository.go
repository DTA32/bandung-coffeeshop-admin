package repository

import (
	"context"
	"errors"
	"slices"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReviewRepository owns cafe_review and its cafe_rating rows (BE migrations
// 011 + 012). Moderation (status, pin) never touches updated_at: that column
// is the author's last edit and orders the public review list.
//
// Ratings are replaced, never upserted: DELETE the review's rows, then INSERT,
// in one transaction. Until 012 runs cafe_rating is UNIQUE (cafe_id,
// category_type), afterwards UNIQUE (review_id, category_type), so no single
// ON CONFLICT target works on both schemas.
type ReviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{db: db}
}

const reviewColumns = `
	rv.id, rv.cafe_id, l.id, l.name, rv.user_id, u.email, rv.reviewer_name, rv.status::text, rv.is_pinned,
	rv.overall_score::float8, rv.wfc_score::float8, rv.content, rv.visited_at::text,
	rv.external_url_title, rv.external_url_link, rv.created_at, rv.updated_at`

const reviewFrom = `
	FROM cafe_review rv
	JOIN cafe c ON c.id = rv.cafe_id
	JOIN location l ON l.id = c.location_id
	LEFT JOIN users u ON u.id = rv.user_id`

func scanReview(row pgx.CollectableRow) (models.Review, error) {
	var rv models.Review
	err := row.Scan(&rv.ID, &rv.CafeID, &rv.LocationID, &rv.CafeName, &rv.UserID, &rv.UserEmail,
		&rv.ReviewerName, &rv.Status, &rv.IsPinned, &rv.OverallScore, &rv.WFCScore, &rv.Content,
		&rv.VisitedAt, &rv.ExternalURLTitle, &rv.ExternalURLLink, &rv.CreatedAt, &rv.UpdatedAt)
	return rv, err
}

// List returns PageSize reviews with status p.Status after the review
// p.AfterID (0 = from the start). Pending is a FIFO queue (oldest edit first);
// the other tabs show the newest first. Cursor paging like /posts: moderated
// reviews leave the tab, which would make an offset skip unseen ones. Reviews
// of deleted users are left out (the public site hides them too). p must be
// normalised (controllers.NormalizeReviewListParams).
func (r *ReviewRepository) List(ctx context.Context, p models.ReviewListParams) (models.Page[models.Review], error) {
	order, cmp := "rv.updated_at DESC, rv.id DESC", "<"
	if p.Status == models.ReviewStatusPending {
		order, cmp = "rv.updated_at ASC, rv.id ASC", ">"
	}
	rows, err := r.db.Query(ctx, `
		SELECT `+reviewColumns+reviewFrom+`
		WHERE rv.status = $1::review_status_enum
		  AND (u.id IS NULL OR u.status <> 'deleted')
		  AND ($2 = 0 OR rv.cafe_id = $2)
		  AND (NOT $3::bool OR rv.user_id IS NULL)
		  AND ($4 = 0 OR (rv.updated_at, rv.id) `+cmp+`
		                 (SELECT a.updated_at, a.id FROM cafe_review a WHERE a.id = $4))
		ORDER BY `+order+`
		LIMIT $5
	`, p.Status, p.CafeID, p.PartnershipOnly, p.AfterID, PageSize+1)
	if err != nil {
		return models.Page[models.Review]{}, err
	}
	items, err := pgx.CollectRows(rows, scanReview)
	if err != nil {
		return models.Page[models.Review]{}, err
	}
	pg := toPage(items)
	return pg, r.attachRatings(ctx, pg.Items)
}

// Get returns one review with its ratings.
func (r *ReviewRepository) Get(ctx context.Context, id int) (*models.Review, error) {
	rows, err := r.db.Query(ctx, `SELECT `+reviewColumns+reviewFrom+` WHERE rv.id = $1`, id)
	if err != nil {
		return nil, err
	}
	rv, err := pgx.CollectExactlyOneRow(rows, scanReview)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	items := []models.Review{rv}
	if err := r.attachRatings(ctx, items); err != nil {
		return nil, err
	}
	return &items[0], nil
}

// attachRatings loads the ratings of items in one query, with each type's
// label and the bucket (rating_category) the score falls in.
func (r *ReviewRepository) attachRatings(ctx context.Context, items []models.Review) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]int, len(items))
	for i, it := range items {
		ids[i] = it.ID
	}
	rows, err := r.db.Query(ctx, `
		SELECT r.review_id, r.category_type, rtl.label, r.score::float8,
		       r.short_description_override, r.short_description_override_indo,
		       COALESCE((SELECT rc.name FROM rating_category rc
		                 WHERE rc.type = r.category_type AND r.score BETWEEN rc.lower_bound AND rc.upper_bound
		                 ORDER BY rc.lower_bound DESC, rc.id LIMIT 1), '')
		FROM cafe_rating r
		JOIN rating_type_label rtl ON rtl.type = r.category_type
		WHERE r.review_id = ANY($1::int[])
		ORDER BY r.review_id, rtl.type
	`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			reviewID int
			rt       models.Rating
		)
		if err := rows.Scan(&reviewID, &rt.CategoryType, &rt.Label, &rt.Score,
			&rt.ShortDescriptionOverride, &rt.ShortDescriptionOverrideIndo, &rt.Bucket); err != nil {
			return err
		}
		if i := slices.IndexFunc(items, func(it models.Review) bool { return it.ID == reviewID }); i >= 0 {
			items[i].Ratings = append(items[i].Ratings, rt)
		}
	}
	return rows.Err()
}

// lockReview locks the review row for the rest of tx and returns its status
// and whether it is a partnership review (user_id IS NULL).
func lockReview(ctx context.Context, tx pgx.Tx, id int) (status string, partnership bool, cafeID int, err error) {
	err = tx.QueryRow(ctx, `
		SELECT status::text, user_id IS NULL, cafe_id FROM cafe_review WHERE id = $1 FOR UPDATE
	`, id).Scan(&status, &partnership, &cafeID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return status, partnership, cafeID, err
}

// SetStatus moves the review to status `to` when its current status is in
// from (ErrReviewStateChanged otherwise); partnershipOnly additionally
// requires user_id IS NULL (ErrNotFound otherwise). updated_at is untouched.
func (r *ReviewRepository) SetStatus(ctx context.Context, id int, to string, from []string, partnershipOnly bool) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	status, partnership, _, err := lockReview(ctx, tx, id)
	if err != nil {
		return err
	}
	if partnershipOnly && !partnership {
		return ErrNotFound
	}
	if !slices.Contains(from, status) {
		return ErrReviewStateChanged
	}
	if _, err := tx.Exec(ctx, `UPDATE cafe_review SET status = $2::review_status_enum WHERE id = $1`, id, to); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SetPinned sets is_pinned on a non-deleted review (ErrReviewStateChanged for
// a deleted one). updated_at is untouched.
func (r *ReviewRepository) SetPinned(ctx context.Context, id int, pinned bool) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	status, _, _, err := lockReview(ctx, tx, id)
	if err != nil {
		return err
	}
	if status == models.ReviewStatusDeleted {
		return ErrReviewStateChanged
	}
	if _, err := tx.Exec(ctx, `UPDATE cafe_review SET is_pinned = $2 WHERE id = $1`, id, pinned); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ReplaceRatings swaps a non-deleted review's ratings for ratings (status and
// updated_at untouched).
func (r *ReviewRepository) ReplaceRatings(ctx context.Context, id int, ratings []models.Rating) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	status, _, cafeID, err := lockReview(ctx, tx, id)
	if err != nil {
		return err
	}
	if status == models.ReviewStatusDeleted {
		return ErrReviewStateChanged
	}
	if err := writeRatings(ctx, tx, id, cafeID, ratings); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// CreatePartnership inserts an approved partnership review (user_id NULL) and
// its ratings in one transaction; returns the new id.
func (r *ReviewRepository) CreatePartnership(ctx context.Context, rv models.Review) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	if err := cafeExists(ctx, tx, rv.CafeID); err != nil {
		return 0, err
	}
	var id int
	if err := tx.QueryRow(ctx, `
		INSERT INTO cafe_review (cafe_id, user_id, reviewer_name, status, is_pinned, overall_score, wfc_score,
			content, visited_at, external_url_title, external_url_link)
		VALUES ($1, NULL, $2, 'approved', $3, $4, $5, $6, $7::date, $8, $9)
		RETURNING id
	`, rv.CafeID, rv.ReviewerName, rv.IsPinned, rv.OverallScore, rv.WFCScore, rv.Content, rv.VisitedAt,
		rv.ExternalURLTitle, rv.ExternalURLLink).Scan(&id); err != nil {
		return 0, err
	}
	if err := writeRatings(ctx, tx, id, rv.CafeID, rv.Ratings); err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}

// UpdatePartnership rewrites a non-deleted partnership review and its ratings
// in one transaction and sets updated_at = now() (status is kept). A user
// review or a deleted one is ErrNotFound.
func (r *ReviewRepository) UpdatePartnership(ctx context.Context, id int, rv models.Review) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	status, partnership, _, err := lockReview(ctx, tx, id)
	if err != nil {
		return err
	}
	if !partnership || status == models.ReviewStatusDeleted {
		return ErrNotFound
	}
	if err := cafeExists(ctx, tx, rv.CafeID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE cafe_review SET
			cafe_id = $2, reviewer_name = $3, is_pinned = $4, overall_score = $5, wfc_score = $6,
			content = $7, visited_at = $8::date, external_url_title = $9, external_url_link = $10,
			updated_at = NOW()
		WHERE id = $1
	`, id, rv.CafeID, rv.ReviewerName, rv.IsPinned, rv.OverallScore, rv.WFCScore, rv.Content, rv.VisitedAt,
		rv.ExternalURLTitle, rv.ExternalURLLink); err != nil {
		return err
	}
	if err := writeRatings(ctx, tx, id, rv.CafeID, rv.Ratings); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func cafeExists(ctx context.Context, tx pgx.Tx, cafeID int) error {
	var ok bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cafe WHERE id = $1)`, cafeID).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return ErrUnknownCafe
	}
	return nil
}

// writeRatings replaces the review's cafe_rating rows (DELETE + INSERT, no ON
// CONFLICT: see ReviewRepository). A unique violation means the pre-012
// UNIQUE (cafe_id, category_type) still holds and another review of the cafe
// already rates that type.
func writeRatings(ctx context.Context, tx pgx.Tx, reviewID, cafeID int, ratings []models.Rating) error {
	if _, err := tx.Exec(ctx, `DELETE FROM cafe_rating WHERE review_id = $1`, reviewID); err != nil {
		return err
	}
	for _, rt := range ratings {
		if _, err := tx.Exec(ctx, `
			INSERT INTO cafe_rating (cafe_id, category_type, score, short_description_override,
				short_description_override_indo, review_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, cafeID, rt.CategoryType, rt.Score, rt.ShortDescriptionOverride, rt.ShortDescriptionOverrideIndo,
			reviewID); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrRatingTaken
			}
			return err
		}
	}
	return nil
}
