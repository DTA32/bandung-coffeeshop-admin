package models

import "time"

// User is a public-site account (users table, BE migration 009). Mission
// control only lists users and flips their status; sessions live in the BE's
// Redis and the BE re-checks users.status on every request, so a delete here
// locks the user out immediately.
type User struct {
	ID        int
	Email     string
	Name      string
	Status    string
	CreatedAt time.Time
}

// User statuses (user_status_enum).
const (
	UserStatusActive  = "active"
	UserStatusDeleted = "deleted"
)

// UserFilterAll is the status filter that shows every user.
const UserFilterAll = "all"

// UsersPerPage is the /users page size.
const UsersPerPage = 50

// UserListParams is the /users query: search text, status filter and 1-based page.
type UserListParams struct {
	Q      string
	Status string // all | active | deleted
	Page   int
}

// UserPage is one page of the user list plus the total match count.
type UserPage struct {
	Params  UserListParams
	Items   []User
	Total   int
	PerPage int
}

func (p UserPage) TotalPages() int {
	if p.Total == 0 || p.PerPage <= 0 {
		return 1
	}
	return (p.Total + p.PerPage - 1) / p.PerPage
}

func (p UserPage) HasPrev() bool { return p.Params.Page > 1 }
func (p UserPage) HasNext() bool { return p.Params.Page < p.TotalPages() }
