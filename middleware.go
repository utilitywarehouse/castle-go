package castle

import "net/http"

type middlewareOpts struct {
	methodFilter string
	pathFilter   string
	ignoreEmpty  bool
}

type MiddlewareOpt func(*middlewareOpts)

func WithMethodFilter(method string) MiddlewareOpt {
	return func(o *middlewareOpts) {
		o.methodFilter = method
	}
}

func WithPathFilter(path string) MiddlewareOpt {
	return func(o *middlewareOpts) {
		o.pathFilter = path
	}
}

func WithIgnoreEmpty(ignore bool) MiddlewareOpt {
	return func(o *middlewareOpts) {
		o.ignoreEmpty = ignore
	}
}

// Middleware is a function that wraps an http.Handler to inject the Castle context into the request.
// If ignoreEmpty is true, it will skip the middleware if the request token is empty.
func Middleware(opts ...MiddlewareOpt) func(next http.Handler) http.Handler {
	options := &middlewareOpts{
		methodFilter: "POST",
		ignoreEmpty:  true,
	}

	for _, opt := range opts {
		opt(options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.methodFilter != "" && r.Method != options.methodFilter {
				next.ServeHTTP(w, r)
				return
			}
			if options.pathFilter != "" && r.URL.Path != options.pathFilter {
				next.ServeHTTP(w, r)
				return
			}
			castleCtx := FromHTTPRequest(r)
			if options.ignoreEmpty && castleCtx.RequestToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(ToCtx(r.Context(), castleCtx))
			next.ServeHTTP(w, r)
		})
	}
}
