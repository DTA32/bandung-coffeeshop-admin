package post

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

// Validation failures answer before any controller call, so a bare Handler
// (nil controllers) is enough.
func TestPostBatchRejectsBadInputWithFlash(t *testing.T) {
	for name, form := range map[string]url.Values{
		"change action": {"action": {"approve"}, "id": {"1"}},
		"no ids":        {"action": {"publish"}},
		"bad id":        {"action": {"ignore"}, "id": {"abc"}},
	} {
		for _, htmx := range []bool{true, false} {
			mux := http.NewServeMux()
			mux.Handle("POST /posts/batch", NewPostPostBatchHandler(&handlers.Handler{}))
			r := httptest.NewRequest("POST", "/posts/batch", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if htmx {
				r.Header.Set("HX-Request", "true")
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, r)
			body := rec.Body.String()
			switch {
			case htmx && (rec.Code != http.StatusOK || !strings.Contains(body, `id="batch-flash"`) || !strings.Contains(body, "hx-swap-oob")):
				t.Errorf("%s: htmx status=%d body=%q", name, rec.Code, body)
			case !htmx && (rec.Code != http.StatusSeeOther || !strings.HasPrefix(rec.Header().Get("Location"), "/posts?err=")):
				t.Errorf("%s: plain status=%d loc=%q", name, rec.Code, rec.Header().Get("Location"))
			}
		}
	}
}
