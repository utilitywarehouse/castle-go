package castle_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/castle-go"
)

func TestToCtxFromRequest(t *testing.T) {
	tests := map[string]struct {
		input    http.Request
		expected castle.Context
	}{
		"castle token on header": {
			input: func() http.Request {
				req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
				req.Header.Set("X-Castle-Request-Token", "foo")
				req.RemoteAddr = "2.2.2.2"
				return *req
			}(),
			expected: castle.Context{
				IP:           "2.2.2.2",
				Headers:      map[string]string{"X-Castle-Request-Token": "foo"},
				RequestToken: "foo",
			},
		},
		"castle token in form": {
			input: func() http.Request {
				req := httptest.NewRequest(http.MethodPost, "http://example.com/bar?castle_request_token=bar", nil)
				req.RemoteAddr = "2.2.2.2"
				return *req
			}(),
			expected: castle.Context{
				IP:           "2.2.2.2",
				Headers:      map[string]string{},
				RequestToken: "bar",
			},
		},
		"no castle token": {
			input: func() http.Request {
				req := http.Request{}
				req.RemoteAddr = "2.2.2.2"

				return req
			}(),
			expected: castle.Context{
				IP:           "2.2.2.2",
				Headers:      map[string]string{},
				RequestToken: "",
			},
		},
	}
	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			gotCtx := castle.ToCtxFromHTTPRequest(ctx, &test.input)
			got := castle.FromCtx(gotCtx)
			assert.Equal(t, test.expected, *got)
		})
	}
}
