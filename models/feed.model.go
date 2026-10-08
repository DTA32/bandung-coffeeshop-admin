package models

import "time"

type Feed struct {
	ID          int
	CafeID      int
	Content     string
	Images      []string
	Author      string
	Link        *string
	PublishedAt time.Time
	ExpiresAt   *time.Time
	Visible     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Active reports whether the entry is shown on the site right now.
func (f Feed) Active(now time.Time) bool {
	return f.Visible && (f.ExpiresAt == nil || f.ExpiresAt.After(now))
}
