package castle

import "net/http"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		castleCtx := FromHTTPRequest(r)
		r = r.WithContext(ToCtx(r.Context(), castleCtx))

		next.ServeHTTP(w, r)
	})
}
