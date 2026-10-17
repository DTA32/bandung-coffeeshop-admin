package models

// Page is one page of a paginated list; HasMore reports whether a following
// page exists (repositories fetch one extra row to know).
type Page[T any] struct {
	Items   []T
	HasMore bool
}
