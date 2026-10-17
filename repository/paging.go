package repository

import (
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// PageSize is the number of rows per list page (infinite scroll). Paged
// queries fetch PageSize+1 rows to learn whether a next page exists.
const PageSize = 50

// pageWindow returns LIMIT / OFFSET for a 1-based page (limit includes the
// extra look-ahead row).
func pageWindow(page int) (limit, offset int) {
	if page < 1 {
		page = 1
	}
	return PageSize + 1, (page - 1) * PageSize
}

// toPage trims the look-ahead row off items.
func toPage[T any](items []T) models.Page[T] {
	if len(items) > PageSize {
		return models.Page[T]{Items: items[:PageSize], HasMore: true}
	}
	return models.Page[T]{Items: items}
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// EscapeLike escapes LIKE / ILIKE wildcards (and the escape char itself) so
// user input matches literally; pass the result as a bind parameter inside
// '%' || $n || '%'.
func EscapeLike(s string) string { return likeEscaper.Replace(s) }
