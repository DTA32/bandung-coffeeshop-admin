// Package utils holds small request and value helpers shared by handlers.
package utils

import (
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func RedirectOK(w http.ResponseWriter, r *http.Request, path, msg string) {
	if msg != "" {
		path += "?ok=" + url.QueryEscape(msg)
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

func PathInt(r *http.Request, name string) (int, bool) {
	n, err := strconv.Atoi(r.PathValue(name))
	return n, err == nil
}

func PathInt64(r *http.Request, name string) (int64, bool) {
	n, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	return n, err == nil
}

func QueryInt(r *http.Request, name string) int {
	n, _ := strconv.Atoi(r.URL.Query().Get(name))
	return n
}

func Contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func DerefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func FormatLocal(t *time.Time, tz *time.Location) string {
	if t == nil {
		return ""
	}
	return t.In(tz).Format("2006-01-02T15:04")
}
