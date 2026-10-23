package routes

import (
	"net/http"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers/user"
)

// UserRoutes registers the user endpoints.
func UserRoutes(mux *http.ServeMux, h *handlers.Handler) {
	mux.Handle("GET /users", user.NewGetUsersHandler(h))
	mux.Handle("POST /users/{id}/delete", user.NewPostUserDeleteHandler(h))
	mux.Handle("POST /users/{id}/revive", user.NewPostUserReviveHandler(h))
}
