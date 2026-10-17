package repository

import (
	"context"
	"errors"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SourceRepository owns source_target, source_snapshot and scrape_run.
type SourceRepository struct {
	db *pgxpool.Pool
}

func NewSourceRepository(db *pgxpool.Pool) *SourceRepository {
	return &SourceRepository{db: db}
}

// ListTargets returns every target matching f (used for single-target lookups).
func (r *SourceRepository) ListTargets(ctx context.Context, f models.TargetFilter) ([]models.SourceTarget, error) {
	return r.listTargets(ctx, f, nil, 0)
}

// ListTargetsPage returns one page (1-based) of targets matching f.
func (r *SourceRepository) ListTargetsPage(ctx context.Context, f models.TargetFilter, page int) (models.Page[models.SourceTarget], error) {
	limit, offset := pageWindow(page)
	items, err := r.listTargets(ctx, f, &limit, offset)
	return toPage(items), err
}

func (r *SourceRepository) listTargets(ctx context.Context, f models.TargetFilter, limit *int, offset int) ([]models.SourceTarget, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.cafe_id, l.id, l.name, t.source::text, t.external_id, t.enabled,
		       t.last_success_at, t.last_error, t.consecutive_failures, s.id, s.fetched_at
		FROM source_target t
		JOIN cafe c ON c.id = t.cafe_id
		JOIN location l ON l.id = c.location_id
		LEFT JOIN LATERAL (
			SELECT id, fetched_at FROM source_snapshot
			WHERE cafe_id = t.cafe_id AND source = t.source
			ORDER BY fetched_at DESC LIMIT 1
		) s ON true
		WHERE ($1 = '' OR t.source::text = $1)
		  AND (NOT $2 OR t.consecutive_failures > 0)
		  AND ($3 = 0 OR t.cafe_id = $3)
		ORDER BY t.consecutive_failures DESC, l.name, t.cafe_id, t.source
		LIMIT $4 OFFSET $5
	`, f.Source, f.FailingOnly, f.CafeID, limit, offset)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.SourceTarget, error) {
		var t models.SourceTarget
		err := row.Scan(&t.CafeID, &t.LocationID, &t.CafeName, &t.Source, &t.ExternalID, &t.Enabled,
			&t.LastSuccessAt, &t.LastError, &t.ConsecutiveFailures, &t.LatestSnapshotID, &t.LatestSnapshotAt)
		return t, err
	})
}

// UpsertTarget creates or overrides a target's external id (e.g. a Linktree
// URL the scraper could not discover). A nil externalID clears the cache so the
// scraper re-resolves / re-discovers it on the next run.
func (r *SourceRepository) UpsertTarget(ctx context.Context, cafeID int, source string, externalID *string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO source_target (cafe_id, source, external_id)
		VALUES ($1, $2::collector_source, $3)
		ON CONFLICT (cafe_id, source) DO UPDATE SET external_id = EXCLUDED.external_id, updated_at = NOW()
	`, cafeID, source, externalID)
	return err
}

func (r *SourceRepository) SetTargetEnabled(ctx context.Context, cafeID int, source string, enabled bool) error {
	return r.execTarget(ctx, `enabled = $3`, cafeID, source, enabled)
}

func (r *SourceRepository) ResetTargetFailures(ctx context.Context, cafeID int, source string) error {
	return r.execTarget(ctx, `consecutive_failures = 0, last_error = NULL`, cafeID, source)
}

func (r *SourceRepository) execTarget(ctx context.Context, set string, cafeID int, source string, args ...any) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE source_target SET `+set+`, updated_at = NOW()
		WHERE cafe_id = $1 AND source = $2::collector_source
	`, append([]any{cafeID, source}, args...)...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SourceRepository) GetSnapshot(ctx context.Context, id int64) (*models.Snapshot, error) {
	var s models.Snapshot
	var raw, parsed []byte
	err := r.db.QueryRow(ctx, `
		SELECT s.id, s.cafe_id, l.id, l.name, s.source::text, s.fetched_at, s.raw, s.parsed, s.hash
		FROM source_snapshot s
		JOIN cafe c ON c.id = s.cafe_id
		JOIN location l ON l.id = c.location_id
		WHERE s.id = $1
	`, id).Scan(&s.ID, &s.CafeID, &s.LocationID, &s.CafeName, &s.Source, &s.FetchedAt, &raw, &parsed, &s.Hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	s.Raw, s.Parsed = raw, parsed
	return &s, err
}

func scanRun(row pgx.CollectableRow) (models.ScrapeRun, error) {
	var s models.ScrapeRun
	err := row.Scan(&s.ID, &s.Source, &s.StartedAt, &s.FinishedAt, &s.OK, &s.Failed, &s.Proposals, &s.Error)
	return s, err
}

// LastRuns returns the latest run per source.
func (r *SourceRepository) LastRuns(ctx context.Context) ([]models.ScrapeRun, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (source) id, source::text, started_at, finished_at, ok, failed, proposals, error
		FROM scrape_run ORDER BY source, started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanRun)
}

func (r *SourceRepository) RecentRuns(ctx context.Context, limit int) ([]models.ScrapeRun, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, source::text, started_at, finished_at, ok, failed, proposals, error
		FROM scrape_run ORDER BY started_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanRun)
}
