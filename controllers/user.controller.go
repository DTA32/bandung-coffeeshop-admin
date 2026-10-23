package controllers

import (
	"context"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

// UserController lists users and deletes / revives them. A delete only sets
// users.status = 'deleted': the BE checks the status on every request (so the
// user is logged out at once) and hides their reviews at read time; a revive
// restores both. Nothing here touches sessions or reviews.
type UserController struct {
	repo *repository.UserRepository
}

func NewUserController(repo *repository.UserRepository) *UserController {
	return &UserController{repo: repo}
}

func (c *UserController) List(ctx context.Context, p models.UserListParams) (models.Page[models.User], error) {
	return c.repo.List(ctx, NormalizeUserListParams(p))
}

// NormalizeUserListParams trims the search, maps unknown status filters to
// "all" and clamps the page to >= 1.
func NormalizeUserListParams(p models.UserListParams) models.UserListParams {
	p.Q = strings.TrimSpace(p.Q)
	switch p.Status {
	case models.UserStatusActive, models.UserStatusDeleted:
	default:
		p.Status = models.UserFilterAll
	}
	if p.Page < 1 {
		p.Page = 1
	}
	return p
}

// Delete marks the user deleted; already deleted is a no-op.
func (c *UserController) Delete(ctx context.Context, id int) (*models.User, error) {
	return c.SetStatus(ctx, id, models.UserStatusDeleted)
}

// Revive marks the user active again; already active is a no-op.
func (c *UserController) Revive(ctx context.Context, id int) (*models.User, error) {
	return c.SetStatus(ctx, id, models.UserStatusActive)
}

// SetStatus moves the user to status and returns the updated user. Setting
// the status the user already has changes nothing (updated_at included).
func (c *UserController) SetStatus(ctx context.Context, id int, status string) (*models.User, error) {
	if status != models.UserStatusActive && status != models.UserStatusDeleted {
		return nil, invalid("invalid user status %q", status)
	}
	u, err := c.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.Status == status {
		return u, nil
	}
	if err := c.repo.SetStatus(ctx, id, status); err != nil {
		return nil, err
	}
	u.Status = status
	return u, nil
}
