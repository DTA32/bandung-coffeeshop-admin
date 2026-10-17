package repository

import (
	"context"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardRepository holds the home-page count query. The last run per source
// comes from SourceRepository.LastRuns (see DashboardController).
type DashboardRepository struct {
	db *pgxpool.Pool
}

func NewDashboardRepository(db *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// Counts returns the dashboard numbers (LastRuns is left empty).
func (r *DashboardRepository) Counts(ctx context.Context) (*models.Dashboard, error) {
	var d models.Dashboard
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM cafe c JOIN location l ON l.id = c.location_id WHERE l.status <> 'deleted'),
			(SELECT COUNT(*) FROM change_proposal WHERE status = 'pending' AND field <> 'ig_post'),
			(SELECT COUNT(*) FROM change_proposal WHERE status = 'pending' AND field = 'ig_post'),
			(SELECT COUNT(*) FROM source_target WHERE enabled AND consecutive_failures >= 3),
			(SELECT COUNT(*) FROM users WHERE status = 'active'),
			(SELECT COUNT(*) FROM users WHERE status = 'deleted')
	`).Scan(&d.Cafes, &d.PendingChanges, &d.PendingPosts, &d.FailingTargets, &d.ActiveUsers, &d.DeletedUsers)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
