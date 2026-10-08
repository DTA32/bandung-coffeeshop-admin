package handlers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

// formError is a user-facing parse error (bad number etc.).
type formError struct{ msg string }

func (e *formError) Error() string { return e.msg }

// FormErrorf returns a user-facing parse error that handlers show next to the form.
func FormErrorf(format string, args ...any) error {
	return &formError{fmt.Sprintf(format, args...)}
}

// OptString returns the trimmed value of key, or nil if the key is absent or blank.
func OptString(v url.Values, key string) *string {
	if _, ok := v[key]; !ok {
		return nil
	}
	s := v.Get(key)
	return controllers.NilIfBlank(&s)
}

// OptInt parses key as a whole number; blank gives nil.
func OptInt(v url.Values, key string) (*int, error) {
	s := strings.TrimSpace(v.Get(key))
	if s == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, &formError{fmt.Sprintf("%s: %q is not a whole number", key, s)}
	}
	return &n, nil
}

// OptFloat parses key as a number; blank gives nil.
func OptFloat(v url.Values, key string) (*float64, error) {
	s := strings.TrimSpace(v.Get(key))
	if s == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, &formError{fmt.Sprintf("%s: %q is not a number", key, s)}
	}
	return &f, nil
}

// Checked reports whether the checkbox key was submitted.
func Checked(v url.Values, key string) bool {
	return v.Get(key) != ""
}

// ParseImages reads the parallel image_url / image_desc / image_desc_indo
// lists; their order is the display order.
func ParseImages(v url.Values) []models.Image {
	urls := v["image_url"]
	desc := v["image_desc"]
	descIndo := v["image_desc_indo"]
	out := make([]models.Image, 0, len(urls))
	for i, u := range urls {
		im := models.Image{URL: strings.TrimSpace(u)}
		if i < len(desc) {
			im.Description = strings.TrimSpace(desc[i])
		}
		if i < len(descIndo) {
			im.DescriptionIndo = strings.TrimSpace(descIndo[i])
		}
		out = append(out, im)
	}
	return out
}
