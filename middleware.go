package castle

import "net/http"

// Middleware is a function that wraps an http.Handler to inject the Castle context into the request.
// If ignoreEmpty is true, it will skip the middleware if the request token is empty.
func Middleware(ignoreEmpty bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			castleCtx := FromHTTPRequest(r)
			if ignoreEmpty && castleCtx.RequestToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(ToCtx(r.Context(), castleCtx))
			next.ServeHTTP(w, r)
		})
	}
}
