package controllers

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/repository"
)

// Repository sentinel errors, re-exported so handlers only import controllers.
var (
	ErrNotFound           = repository.ErrNotFound
	ErrIDExists           = repository.ErrIDExists
	ErrRatingTypeInUse    = repository.ErrRatingTypeInUse
	ErrProposalNotPending = repository.ErrProposalNotPending
	ErrReviewStateChanged = repository.ErrReviewStateChanged
	ErrUnknownCafe        = repository.ErrUnknownCafe
	ErrRatingTaken        = repository.ErrRatingTaken
)

// ValidationError is shown to the admin next to the form.
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

func invalid(format string, a ...any) error {
	return &ValidationError{Msg: fmt.Sprintf(format, a...)}
}

// IsValidation reports whether err should be shown to the admin as-is.
func IsValidation(err error) bool {
	var v *ValidationError
	return errors.As(err, &v) ||
		errors.Is(err, ErrIDExists) ||
		errors.Is(err, ErrRatingTypeInUse) ||
		errors.Is(err, ErrProposalNotPending) ||
		errors.Is(err, ErrReviewStateChanged) ||
		errors.Is(err, ErrUnknownCafe) ||
		errors.Is(err, ErrRatingTaken)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// Shared validation helpers.
var (
	slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	hourRe = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
)

func NilIfBlank(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

func validateHour(h *string) error {
	if h != nil && !hourRe.MatchString(*h) {
		return invalid("invalid hour %q: expected HH:MM", *h)
	}
	return nil
}

var menuGalleryRe = regexp.MustCompile(`^#image-[1-9][0-9]*$`)

// isMenuValue accepts what cafe.menu may hold when edited by hand: an http(s) URL,
// or a "#image-N" gallery hash (N >= 1, no leading zeros). Scraped menu proposals
// stay URL-only (see FieldUpdateFor).
func isMenuValue(s string) bool {
	return isHTTPURL(s) || menuGalleryRe.MatchString(s)
}

func isHTTPURL(s string) bool {
	return strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://")
}

func strPtr(s string) *string { return &s }
