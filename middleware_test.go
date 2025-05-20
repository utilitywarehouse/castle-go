package castle

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Middleware(t *testing.T) {
	t.Run("ctx is set", func(t *testing.T) {
		middleware := Middleware(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			castleCtx := FromCtx(r.Context())
			if castleCtx == nil {
				t.Error("Expected castle context to be present in request context")
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusOK)
		}))

		req, err := http.NewRequestWithContext(t.Context(), "GET", "https://example.com", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
		}
	})
	t.Run("ctx is not set", func(t *testing.T) {
		middleware := Middleware(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			castleCtx := FromCtx(r.Context())
			if castleCtx != nil {
				t.Error("Expected castle context to not be present in request context")
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusOK)
		}))

		req, err := http.NewRequestWithContext(t.Context(), "GET", "https://example.com", nil)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
