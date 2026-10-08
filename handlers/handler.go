// Package handlers holds the code shared by every route handler: the Handler
// dependency bundle (handler.go), render/error helpers (helpers.go), form parsing
// primitives (forms.go) and the health endpoint. Each entity has its own
// subpackage (handlers/cafe, handlers/tag, ...) with one small struct per route,
// one file per route named <get|post><entity><action>.go (getcafes.go,
// postcafeedit.go), plus handler.go (helpers used by several routes of the entity)
// and form.go (form parsing). Handlers parse forms, call the controllers and
// render templ templates. Full forms use POST/redirect/GET; queue actions and
// inline rows are HTMX requests that return the swapped fragment.
package handlers

import (
	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/initializers"
)

// Handler carries the dependencies shared by every route handler.
type Handler struct {
	Ctl *controllers.Controllers
	Cfg *initializers.Config
}

func New(ctl *controllers.Controllers, cfg *initializers.Config) *Handler {
	return &Handler{Ctl: ctl, Cfg: cfg}
}
