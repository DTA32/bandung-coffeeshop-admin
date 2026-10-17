// Package user holds the user route handlers.
package user

import (
	"context"
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

// setStatus runs a delete / revive from the {id} path and answers with the
// updated row for the HTMX outerHTML swap.
func setStatus(h *handlers.Handler, w http.ResponseWriter, r *http.Request,
	apply func(context.Context, int) (*models.User, error)) {
	id, ok := utils.PathInt(r, "id")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u, err := apply(r.Context(), id)
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "User")
		return
	}
	if err != nil {
		h.ServerError(w, r, err)
		return
	}
	h.Render(w, r, http.StatusOK, templates.UserRow(*u))
}
