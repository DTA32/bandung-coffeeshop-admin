// Package controllers holds the business rules: validation, normalisation,
// proposal apply rules and feed rules, one file per entity
// (<entity>.controller.go). Controllers never write SQL; they call the
// matching repository, which owns the queries and the transactions.
package controllers

import (
	"context"

	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

// Controllers bundles one controller per entity.
type Controllers struct {
	repos *repository.Repositories

	Cafe           *CafeController
	Location       *LocationController
	Tag            *TagController
	RatingType     *RatingTypeController
	RatingCategory *RatingCategoryController
	Feed           *FeedController
	Change         *ChangeController
	Post           *PostController
	Source         *SourceController
	User           *UserController
	Dashboard      *DashboardController
}

func New(repos *repository.Repositories) *Controllers {
	return &Controllers{
		repos:          repos,
		Cafe:           NewCafeController(repos.Cafe),
		Location:       NewLocationController(repos.Location),
		Tag:            NewTagController(repos.Tag),
		RatingType:     NewRatingTypeController(repos.RatingType),
		RatingCategory: NewRatingCategoryController(repos.RatingCategory),
		Feed:           NewFeedController(repos.Feed),
		Change:         NewChangeController(repos.Change),
		Post:           NewPostController(repos.Post),
		Source:         NewSourceController(repos.Source),
		User:           NewUserController(repos.User),
		Dashboard:      NewDashboardController(repos.Dashboard, repos.Source),
	}
}

// Ping reports database reachability (used by /health).
func (c *Controllers) Ping(ctx context.Context) error {
	return c.repos.Ping(ctx)
}
