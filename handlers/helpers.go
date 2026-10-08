package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/templates"
	"github.com/dta32/bandung-coffeeshop-admin/utils"
	"github.com/jackc/pgx/v5/pgconn"
)

// FriendlyUnique turns a unique-constraint violation into a form error.
func FriendlyUnique(err error, what string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return &formError{what + " is already used"}
	}
	return err
}

// Render writes c as an HTML response with the given status.
func (h *Handler) Render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		slog.Error("render", "path", r.URL.Path, "err", err)
	}
}

// ServerError logs err and answers 500 (HTMX requests are told not to swap).
func (h *Handler) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	if utils.IsHTMX(r) {
		// HTMX does not swap non-2xx responses by default; surface the error.
		w.Header().Set("HX-Reswap", "none")
	}
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

// NotFound renders the 404 page for the named resource.
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request, what string) {
	h.Render(w, r, http.StatusNotFound, templates.NotFound(what))
}

// UserError returns the message to show for validation-type errors, or "".
func UserError(err error) string {
	var fe *formError
	if errors.As(err, &fe) || controllers.IsValidation(err) {
		return err.Error()
	}
	return ""
}

// SavedOr returns the user-facing error message, or "" on success.
func SavedOr(err error) string {
	if err == nil {
		return ""
	}
	return UserError(err)
}
