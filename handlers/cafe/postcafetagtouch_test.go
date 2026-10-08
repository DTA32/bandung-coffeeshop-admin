package cafe

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
)

func TestTouchTagRejectsNonNumericTagID(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("POST /cafes/{id}/tags/{tagId}/touch", NewPostCafeTagTouchHandler(&handlers.Handler{}))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("POST", "/cafes/some-cafe/tags/abc/touch", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
