// Package post holds the post route handlers.
package post

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
)

func loadPost(h *handlers.Handler, w http.ResponseWriter, r *http.Request) (*models.Post, bool) {
	id, ok := utils.PathInt64(r, "id")
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return nil, false
	}
	post, err := h.Ctl.Post.Get(r.Context(), id)
	if controllers.IsNotFound(err) {
		h.NotFound(w, r, "Post")
		return nil, false
	}
	if err != nil {
		h.ServerError(w, r, err)
		return nil, false
	}
	return post, true
}
