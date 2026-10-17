package repository

import (
	"context"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository reads the users table (owned by BE migration 009) and flips
// users.status. It never touches sessions (Redis, BE) or reviews: the BE hides
// a deleted user's reviews at read time.
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// userListWhere matches $1 (an escaped ILIKE pattern, empty = any) against name
// and email, and $2 against status (empty = any).
const userListWhere = `
	WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
	  AND ($2 = '' OR status = $2::user_status_enum)`

// List returns one page of users, newest first, plus the total match count.
// p must already be normalised (controllers.NormalizeUserListParams).
func (r *UserRepository) List(ctx context.Context, p models.UserListParams) (*models.UserPage, error) {
	q := escapeLike(p.Q)
	status := p.Status
	if status == models.UserFilterAll {
		status = ""
	}
	page := &models.UserPage{Params: p, PerPage: models.UsersPerPage}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`+userListWhere, q, status).Scan(&page.Total); err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, email, name, status::text, created_at FROM users`+userListWhere+`
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`, q, status, models.UsersPerPage, (p.Page-1)*models.UsersPerPage)
	if err != nil {
		return nil, err
	}
	page.Items, err = pgx.CollectRows(rows, scanUser)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (r *UserRepository) Get(ctx context.Context, id int) (*models.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, email, name, status::text, created_at FROM users WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	u, err := pgx.CollectExactlyOneRow(rows, scanUser)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return &u, err
}

// SetStatus sets users.status (active | deleted) and bumps updated_at.
func (r *UserRepository) SetStatus(ctx context.Context, id int, status string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET status = $2::user_status_enum, updated_at = NOW() WHERE id = $1
	`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanUser(row pgx.CollectableRow) (models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Status, &u.CreatedAt)
	return u, err
}

// likeEscaper escapes the LIKE wildcards so a search for "a_b" or "50%" matches
// literally (backslash is Postgres' default LIKE escape character).
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func escapeLike(s string) string { return likeEscaper.Replace(s) }
