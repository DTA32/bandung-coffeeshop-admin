package change

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/controllers"
	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

// Validation failures answer before any controller call, so a bare Handler
// (nil controllers) is enough.
func postBatch(t *testing.T, form url.Values, htmx bool) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("POST /changes/batch", NewPostChangeBatchHandler(&handlers.Handler{}))
	r := httptest.NewRequest("POST", "/changes/batch", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if htmx {
		r.Header.Set("HX-Request", "true")
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, r)
	return rec
}

func TestBatchRejectsBadInputWithFlash(t *testing.T) {
	many := url.Values{"action": {"approve"}}
	for i := 1; i <= controllers.MaxBatch+1; i++ {
		many.Add("id", strconv.Itoa(i))
	}
	for name, form := range map[string]url.Values{
		"bad action": {"action": {"drop"}, "id": {"1"}},
		"no ids":     {"action": {"approve"}},
		"bad id":     {"action": {"reject"}, "id": {"abc"}},
		"too many":   many,
	} {
		rec := postBatch(t, form, true)
		body := rec.Body.String()
		if rec.Code != http.StatusOK || !strings.Contains(body, `id="batch-flash"`) || !strings.Contains(body, `hx-swap-oob`) {
			t.Errorf("%s: htmx status=%d body=%q", name, rec.Code, body)
		}
		rec = postBatch(t, form, false)
		if rec.Code != http.StatusSeeOther || !strings.HasPrefix(rec.Header().Get("Location"), "/changes?err=") {
			t.Errorf("%s: plain status=%d loc=%q", name, rec.Code, rec.Header().Get("Location"))
		}
	}
}
