package http_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	http_internal "github.com/utilitywarehouse/castle-go/http"
)

func TestIPFromRequest(t *testing.T) {
	tests := map[string]struct {
		input    *http.Request
		expected string
	}{
		"empty": {
			input: httpRequest(map[string]string{
				"X-Real-Ip":        "",
				"X-Forwarded-For":  "",
				"Cf-Connecting-Ip": "",
			}, ""),
			expected: "",
		},
		"cf-connecting-ip": {
			input: httpRequest(map[string]string{
				"X-Real-Ip":        "foo",
				"X-Forwarded-For":  "bar",
				"Cf-Connecting-Ip": "cf-connecting-ip",
			}, "foobar"),
			expected: "cf-connecting-ip",
		},
		"x-forwarded-for": {
			input: httpRequest(map[string]string{
				"X-Real-Ip":        "foo",
				"X-Forwarded-For":  "127.0.0.1, 109.14.23.2",
				"Cf-Connecting-Ip": "",
			}, "foobar"),
			expected: "109.14.23.2",
		},
		"x-real-ip": {
			input: httpRequest(map[string]string{
				"X-Real-Ip":        "x-real-ip",
				"X-Forwarded-For":  "",
				"Cf-Connecting-Ip": "",
			}, "foobar"),
			expected: "x-real-ip",
		},
		"remote-addr": {
			input: httpRequest(map[string]string{
				"X-Real-Ip":        "",
				"X-Forwarded-For":  "",
				"Cf-Connecting-Ip": "",
			}, "remote-addr:8080"),
			expected: "remote-addr",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := http_internal.IPFromRequest(test.input)
			assert.Equal(t, test.expected, got)
		})
	}
}

func httpRequest(headers map[string]string, remoteAddr string) *http.Request {
	r := &http.Request{
		RemoteAddr: remoteAddr,
		Header:     make(http.Header),
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}
