package castle_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/utilitywarehouse/castle-go"
)

func configureHTTPRequest() *http.Request {
	req := httptest.NewRequest("GET", "/", nil)

	req.Header.Set("X-FORWARDED-FOR", "6.6.6.6, 3.3.3.3, 8.8.8.8")
	req.Header.Set("USER-AGENT", "some-agent")
	req.Header.Set("X-CASTLE-REQUEST-TOKEN", "request-token")

	return req
}

func configureRequest(httpReq *http.Request) *castle.Request {
	return &castle.Request{
		Context: castle.FromHTTPRequest(httpReq),
		Event: castle.Event{
			EventType:   castle.EventTypeLogin,
			EventStatus: castle.EventStatusSucceeded,
		},
		User: castle.User{
			ID:     "user-id",
			Email:  "user@test.com",
			Traits: map[string]string{"trait1": "traitValue1"},
		},
		Properties: map[string]string{"prop1": "propValue1"},
		CreatedAt:  time.Now(),
	}
}

func TestCastle_SendFilterCall(t *testing.T) {
	ctx := context.Background()
	req := configureRequest(configureHTTPRequest())

	cstl, err := castle.New("secret-string")
	require.NoError(t, err)

	t.Run("response error", func(t *testing.T) {
		fs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			_, err := w.Write([]byte(`{"error": "this is an error"}`))
			require.NoError(t, err)
		}))

		castle.FilterEndpoint = fs.URL

		res, err := cstl.Filter(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, castle.RecommendedActionNone, res)
	})

	t.Run("bad client request response", func(t *testing.T) {
		fs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(400)
			w.Write([]byte(`foo`)) // nolint: errcheck
		}))

		castle.FilterEndpoint = fs.URL

		res, err := cstl.Filter(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &castle.APIError{}, err)
		assert.Equal(t, &castle.APIError{StatusCode: 400, Message: "foo"}, err)
		assert.Equal(t, castle.RecommendedActionNone, res)
	})

	t.Run("allow action response", func(t *testing.T) {
		fs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(201)
			_, err := w.Write([]byte(`{"policy": {"action": "allow"}}`))
			require.NoError(t, err)
		}))

		castle.FilterEndpoint = fs.URL

		res, err := cstl.Filter(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, castle.RecommendedActionAllow, res)
	})
}

func TestCastle_Filter(t *testing.T) {
	ctx := context.Background()

	t.Run("validation error", func(t *testing.T) {
		cstl, err := castle.New("secret-string")
		require.NoError(t, err)

		tests := map[string]struct {
			input       *castle.Request
			expectedErr string
		}{
			"missing castle.Request": {
				input:       nil,
				expectedErr: "request cannot be nil",
			},
			"missing castle.Context": {
				input: &castle.Request{
					Context: nil,
					Event: castle.Event{
						EventType:   castle.EventTypeLogin,
						EventStatus: castle.EventStatusSucceeded,
					},
					User: castle.User{
						ID:     "user-id",
						Traits: map[string]string{"trait1": "traitValue1"},
					},
					Properties: map[string]string{"prop1": "propValue1"},
				},
				expectedErr: "request.Context cannot be nil",
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				gotRecommendation, gotErr := cstl.Filter(ctx, test.input)
				assert.Equal(t, castle.RecommendedActionNone, gotRecommendation)
				assert.ErrorContains(t, gotErr, test.expectedErr)
			})
		}
	})

	t.Run("executed", func(t *testing.T) {
		httpReq := configureHTTPRequest()
		req := configureRequest(httpReq)

		cstl, err := castle.New("secret-string")
		require.NoError(t, err)

		executed := false

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(201)
			_, err := w.Write([]byte(`{"policy": {"name": "name"}}`))
			require.NoError(t, err)

			type castleFilterRequest struct {
				Type         castle.EventType   `json:"type"`
				Status       castle.EventStatus `json:"status"`
				RequestToken string             `json:"request_token"`
				Params       castle.Params      `json:"params"`
				Context      *castle.Context    `json:"context"`
				Properties   map[string]string  `json:"properties"`
			}

			reqData := &castleFilterRequest{}

			username, password, ok := r.BasicAuth()

			assert.Empty(t, username)
			assert.Equal(t, password, "secret-string")
			assert.True(t, ok)

			err = json.NewDecoder(r.Body).Decode(reqData)
			require.NoError(t, err)

			assert.Equal(t, castle.EventTypeLogin, reqData.Type)
			assert.Equal(t, castle.EventStatusSucceeded, reqData.Status)
			assert.Equal(t, "user@test.com", reqData.Params.Email)
			assert.Equal(t, map[string]string{"prop1": "propValue1"}, reqData.Properties)
			assert.Equal(t, castle.FromHTTPRequest(httpReq), reqData.Context)

			executed = true
		}))

		castle.FilterEndpoint = ts.URL

		res, err := cstl.Filter(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, castle.RecommendedActionNone, res)

		assert.True(t, executed)
	})
}

func TestCastle_Risk(t *testing.T) {
	ctx := context.Background()

	t.Run("validation error", func(t *testing.T) {
		cstl, err := castle.New("secret-string")
		require.NoError(t, err)

		tests := map[string]struct {
			input       *castle.Request
			expectedErr string
		}{
			"missing castle.Request": {
				input:       nil,
				expectedErr: "request cannot be nil",
			},
			"missing castle.Context": {
				input: &castle.Request{
					Context: nil,
					Event: castle.Event{
						EventType:   castle.EventTypeLogin,
						EventStatus: castle.EventStatusSucceeded,
					},
					User: castle.User{
						ID:     "user-id",
						Traits: map[string]string{"trait1": "traitValue1"},
					},
					Properties: map[string]string{"prop1": "propValue1"},
				},
				expectedErr: "request.Context cannot be nil",
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				gotRecommendation, gotErr := cstl.Risk(ctx, test.input)
				assert.Equal(t, castle.RecommendedActionNone, gotRecommendation)
				assert.ErrorContains(t, gotErr, test.expectedErr)
			})
		}
	})

	t.Run("executed", func(t *testing.T) {
		httpReq := configureHTTPRequest()
		req := configureRequest(httpReq)

		cstl, err := castle.New("secret-string")
		require.NoError(t, err)

		executed := false

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(201)
			_, err := w.Write([]byte(`{"policy": {"name": "name"}}`))
			require.NoError(t, err)

			type castleRiskRequest struct {
				Type         castle.EventType   `json:"type"`
				Status       castle.EventStatus `json:"status"`
				RequestToken string             `json:"request_token"`
				User         castle.User        `json:"user"`
				Context      *castle.Context    `json:"context"`
				Properties   map[string]string  `json:"properties"`
			}

			reqData := &castleRiskRequest{}

			username, password, ok := r.BasicAuth()

			assert.Empty(t, username)
			assert.Equal(t, password, "secret-string")
			assert.True(t, ok)

			err = json.NewDecoder(r.Body).Decode(reqData)
			require.NoError(t, err)

			assert.Equal(t, castle.EventTypeLogin, reqData.Type)
			assert.Equal(t, castle.EventStatusSucceeded, reqData.Status)
			assert.Equal(t, "user-id", reqData.User.ID)
			assert.Equal(t, map[string]string{"prop1": "propValue1"}, reqData.Properties)
			assert.Equal(t, map[string]string{"trait1": "traitValue1"}, reqData.User.Traits)
			assert.Equal(t, castle.FromHTTPRequest(httpReq), reqData.Context)

			executed = true
		}))

		castle.RiskEndpoint = ts.URL

		_, err = cstl.Risk(ctx, req)
		require.NoError(t, err)

		assert.True(t, executed)
	})
}

func TestCastle_SendRiskCall(t *testing.T) {
	ctx := context.Background()
	req := configureRequest(configureHTTPRequest())

	cstl, err := castle.New("secret-string")
	require.NoError(t, err)

	t.Run("response error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			_, err := w.Write([]byte(`{"error": "this is an error"}`))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)

		castle.RiskEndpoint = ts.URL

		res, err := cstl.Risk(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, castle.RecommendedActionNone, res)
	})

	t.Run("bad client request response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(400)
			w.Write([]byte(`foo`)) // nolint: errcheck
		}))
		t.Cleanup(ts.Close)

		castle.RiskEndpoint = ts.URL

		res, err := cstl.Risk(ctx, req)
		assert.Error(t, err)
		assert.IsType(t, &castle.APIError{}, err)
		assert.Equal(t, &castle.APIError{StatusCode: 400, Message: "foo"}, err)
		assert.Equal(t, castle.RecommendedActionNone, res)
	})

	t.Run("invalid parameter response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			_, err := w.Write([]byte(`{"type": "invalid_parameter", "message": "error message"}`))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)

		castle.RiskEndpoint = ts.URL

		res, err := cstl.Risk(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, castle.RecommendedActionNone, res)
	})

	t.Run("allow action response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, err := w.Write([]byte(`{"policy": { "action": "allow"}}`))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)

		castle.RiskEndpoint = ts.URL

		res, err := cstl.Risk(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, castle.RecommendedActionAllow, res)
	})

	t.Run("challenge action response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, err := w.Write([]byte(`{"policy": { "action": "challenge"}}`))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)

		castle.RiskEndpoint = ts.URL

		res, err := cstl.Risk(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, castle.RecommendedActionChallenge, res)
	})

	t.Run("deny action response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, err := w.Write([]byte(`{"policy": { "action": "deny"}}`))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)

		castle.RiskEndpoint = ts.URL

		res, err := cstl.Risk(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, castle.RecommendedActionDeny, res)
	})
}
