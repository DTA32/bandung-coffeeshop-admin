// Package repository holds every SQL statement (pgx, raw SQL) and every
// transaction, one file per entity (<entity>.repository.go). It knows nothing
// about validation or HTTP: controllers call it with already-normalised models.
//
// Multi-table writes run inside a single transaction owned by the repository
// method that performs them (cafe create/update, location create/update,
// tag delete, proposal approval, feed-from-post).
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrIDExists           = errors.New("id already exists")
	ErrRatingTypeInUse    = errors.New("rating type is in use by categories or cafe ratings")
	ErrProposalNotPending = errors.New("proposal is no longer pending")
)

// Repositories bundles one repository per entity over a shared pool.
type Repositories struct {
	db *pgxpool.Pool

	Cafe           *CafeRepository
	Location       *LocationRepository
	Tag            *TagRepository
	RatingType     *RatingTypeRepository
	RatingCategory *RatingCategoryRepository
	Feed           *FeedRepository
	Change         *ChangeRepository
	Post           *PostRepository
	Source         *SourceRepository
	User           *UserRepository
	Dashboard      *DashboardRepository
}

func New(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		db:             db,
		Cafe:           NewCafeRepository(db),
		Location:       NewLocationRepository(db),
		Tag:            NewTagRepository(db),
		RatingType:     NewRatingTypeRepository(db),
		RatingCategory: NewRatingCategoryRepository(db),
		Feed:           NewFeedRepository(db),
		Change:         NewChangeRepository(db),
		Post:           NewPostRepository(db),
		Source:         NewSourceRepository(db),
		User:           NewUserRepository(db),
		Dashboard:      NewDashboardRepository(db),
	}
}

func (r *Repositories) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
