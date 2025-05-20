package castle

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Middleware(t *testing.T) {
	t.Run("ctx is set", func(t *testing.T) {
		middleware := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			castleCtx := FromCtx(r.Context())
			if castleCtx == nil {
				t.Error("Expected castle context to be present in request context")
			}
			w.WriteHeader(http.StatusOK)
		}))

		req, err := http.NewRequest("GET", "https://example.com", nil)
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
