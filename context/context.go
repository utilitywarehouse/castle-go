package context

import (
	"context"
	"net/http"
	"strings"

	"github.com/utilitywarehouse/castle-go"
	http_internal "github.com/utilitywarehouse/castle-go/http"
)

type contextKey string

func (c contextKey) String() string {
	return "castle context key " + string(c)
}

var castleCtxKey = contextKey("castle_context")

// ToCtxFromRequest adds the token and other request information (i.e. castle context) to the context.
func ToCtxFromRequest(ctx context.Context, r *http.Request) context.Context {
	castleCtx := castle.Context{
		RequestToken: func() string {
			// grab the token from header if it exists
			if tkn := tokenFromHTTPHeader(r.Header); tkn != "" {
				return tkn
			}

			// otherwise, try grabbing it from form
			return tokenFromHTTPForm(r)
		}(),
		IP:      http_internal.IPFromRequest(r),
		Headers: FilterHeaders(r.Header), // pass in as much context as possible
	}
	return context.WithValue(ctx, castleCtxKey, castleCtx)
}

func FromCtx(ctx context.Context) *castle.Context {
	castleCtx, ok := ctx.Value(castleCtxKey).(castle.Context)
	if ok {
		return &castleCtx
	}
	return nil
}

func tokenFromHTTPHeader(header http.Header) string {
	// recommended header name
	if t := header.Get("X-Castle-Request-Token"); t != "" {
		return t
	}
	// header name used in the frontends
	if t := header.Get("Castle-Token"); t != "" {
		return t
	}
	return ""
}

func tokenFromHTTPForm(r *http.Request) string {
	// ParseForm is idempotent, so it's safe to call from anywhere
	if err := r.ParseForm(); err != nil {
		return ""
	}

	return r.Form.Get("castle_request_token")
}

func FilterHeaders(hs http.Header) map[string]string {
	castleHeaders := make(map[string]string)
	for key, value := range hs {
		// Ensure cookies or authorization are never sent along.
		// Everything else is fair game.
		if _, ok := disallowedHeaders[strings.ToLower(key)]; ok {
			continue
		}
		// View: https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html
		castleHeaders[key] = strings.Join(value, ", ")
	}
	return castleHeaders
}

var disallowedHeaders = map[string]struct{}{
	"cookie":        {},
	"authorization": {},
}
