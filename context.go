package castle

import (
	"context"
	"net/http"
	"strings"

	http_internal "github.com/utilitywarehouse/castle-go/http"
)

var (
	// X-Castle-Request-Token is the recommended header name, while Castle-Token is the header name used in the UW frontends.
	validTokenHeaderNames = []string{"X-Castle-Request-Token", "Castle-Token"}
	validTokenFormNames   = []string{"castle_request_token", "castle-request-token"}
)

type contextKey string

func (c contextKey) String() string {
	return "castle context key " + string(c)
}

var castleCtxKey = contextKey("castle_context")

// ToCtx adds the Castle context to the context.Context.
func ToCtx(ctx context.Context, castleCtx *Context) context.Context {
	return context.WithValue(ctx, castleCtxKey, castleCtx)
}

// FromCtx returns the Castle context from the context.Context.
func FromCtx(ctx context.Context) *Context {
	castleCtx, ok := ctx.Value(castleCtxKey).(*Context)
	if ok {
		return castleCtx
	}
	return nil
}

func FromHTTPRequest(r *http.Request) *Context {
	return &Context{
		RequestToken: func() string {
			// grab the token from header if it exists
			if tkn := tokenFromHTTPHeader(r.Header); tkn != "" {
				return tkn
			}

			// otherwise, try grabbing it from form
			return tokenFromHTTPForm(r)
		}(),
		IP:      http_internal.IPFromRequest(r),
		Headers: filterHeader(r.Header), // pass in as much context as possible
	}
}

// ToCtxFromHTTPRequest adds the token and other request information (i.e. castle context) to the context.
func ToCtxFromHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return ToCtx(ctx, FromHTTPRequest(r))
}

func tokenFromHTTPHeader(header http.Header) string {
	for _, name := range validTokenHeaderNames {
		if t := header.Get(name); t != "" {
			return t
		}
	}
	return ""
}

func tokenFromHTTPForm(r *http.Request) string {
	// ParseForm is idempotent, so it's safe to call from anywhere
	if err := r.ParseForm(); err != nil {
		return ""
	}

	for _, name := range validTokenFormNames {
		if t := r.Form.Get(name); t != "" {
			return t
		}
	}
	return ""
}

func filterHeader(hs http.Header) map[string]string {
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
