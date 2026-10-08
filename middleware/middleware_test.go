package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSameOrigin(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	h := SameOrigin(ok)

	cases := []struct {
		name    string
		method  string
		headers map[string]string
		want    int
	}{
		{"GET cross-site allowed", http.MethodGet, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusNoContent},
		{"POST same-origin", http.MethodPost, map[string]string{"Sec-Fetch-Site": "same-origin"}, http.StatusNoContent},
		{"POST user-initiated (none)", http.MethodPost, map[string]string{"Sec-Fetch-Site": "none"}, http.StatusNoContent},
		{"POST cross-site", http.MethodPost, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"POST same-site subdomain", http.MethodPost, map[string]string{"Sec-Fetch-Site": "same-site"}, http.StatusForbidden},
		{"POST origin match", http.MethodPost, map[string]string{"Origin": "https://mission-control.bdgcafe.com"}, http.StatusNoContent},
		{"POST origin mismatch", http.MethodPost, map[string]string{"Origin": "https://evil.example"}, http.StatusForbidden},
		{"POST origin null", http.MethodPost, map[string]string{"Origin": "null"}, http.StatusForbidden},
		{"POST no headers (curl)", http.MethodPost, nil, http.StatusNoContent},
		{"DELETE cross-site", http.MethodDelete, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "https://mission-control.bdgcafe.com/cafes", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("got %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
