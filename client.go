package castle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var castleReqsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "iam",
	Subsystem: "castle",
	Name:      "requests_total",
	Help:      "Number of requests made to castle",
}, []string{"endpoint", "status"})

var (
	FilterEndpoint = "https://api.castle.io/v1/filter"
	RiskEndpoint   = "https://api.castle.io/v1/risk"
)

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status code: %d, message: %s", e.StatusCode, e.Message)
}

// Castle encapsulates http client
type Castle struct {
	client    *http.Client
	apiSecret string

	metricsEnabled bool
}

// New creates a new castle client with default http client
func New(secret string, opts ...Opt) (*Castle, error) {
	return NewWithHTTPClient(secret, http.DefaultClient, opts...)
}

// NewWithHTTPClient same as New but allows passing of http.Client with custom config
func NewWithHTTPClient(secret string, client *http.Client, opts ...Opt) (*Castle, error) {
	os := &options{
		metricsEnabled: true,
	}
	for _, opt := range opts {
		opt(os)
	}
	return &Castle{
		client:         client,
		apiSecret:      secret,
		metricsEnabled: os.metricsEnabled,
	}, nil
}

// Filter sends a filter request to castle.io
// see https://reference.castle.io/#operation/filter for details
func (c *Castle) Filter(ctx context.Context, req *Request) (RecommendedAction, error) {
	if req == nil {
		return RecommendedActionNone, errors.New("request cannot be nil")
	}
	if req.Context == nil {
		return RecommendedActionNone, errors.New("request.Context cannot be nil")
	}
	params := Params{
		Email:    req.User.Email,
		Username: req.User.Name,
	}
	createdAt := req.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	r := &castleFilterAPIRequest{
		Type:         req.Event.EventType,
		Name:         req.Event.Name,
		Status:       req.Event.EventStatus,
		RequestToken: req.Context.RequestToken,
		Params:       params,
		Context:      req.Context,
		Properties:   req.Properties,
		CreatedAt:    createdAt,
	}
	return c.sendCall(ctx, r, FilterEndpoint)
}

// Risk sends a risk request to castle.io
// see https://reference.castle.io/#operation/risk for details
func (c *Castle) Risk(ctx context.Context, req *Request) (RecommendedAction, error) {
	if req == nil {
		return RecommendedActionNone, errors.New("request cannot be nil")
	}
	if req.Context == nil {
		return RecommendedActionNone, errors.New("request.Context cannot be nil")
	}
	createdAt := req.CreatedAt
	if req.CreatedAt.IsZero() {
		createdAt = time.Now()
	}
	r := &castleRiskAPIRequest{
		Type:         req.Event.EventType,
		Name:         req.Event.Name,
		Status:       req.Event.EventStatus,
		RequestToken: req.Context.RequestToken,
		User:         req.User,
		Context:      req.Context,
		Properties:   req.Properties,
		CreatedAt:    createdAt,
	}
	return c.sendCall(ctx, r, RiskEndpoint)
}

func (c *Castle) sendCall(ctx context.Context, r castleAPIRequest, url string) (_ RecommendedAction, err error) {
	defer func() {
		if !c.metricsEnabled {
			return
		}

		status := "ok"
		if err != nil {
			status = "error"
		}
		castleReqsCounter.WithLabelValues(url, status).Inc()
	}()

	b := new(bytes.Buffer)

	switch request := r.(type) {
	case *castleFilterAPIRequest, *castleRiskAPIRequest:
		err = json.NewEncoder(b).Encode(request)
	default:
		err = fmt.Errorf("incorrect request type passed as argument.")
	}
	if err != nil {
		return RecommendedActionNone, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, b)
	if err != nil {
		return RecommendedActionNone, err
	}

	req.SetBasicAuth("", c.apiSecret)
	req.Header.Set("content-type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return RecommendedActionNone, err
	}
	defer res.Body.Close() // nolint: gosec
	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body) // nolint: errcheck

		return RecommendedActionNone, &APIError{
			StatusCode: res.StatusCode,
			Message:    string(b),
		}
	}

	resp := &castleAPIResponse{}
	if err = json.NewDecoder(res.Body).Decode(resp); err != nil {
		return RecommendedActionNone, fmt.Errorf("unable to decode response body: %w", err)
	}

	return recommendedActionFromString(resp.Policy.Action), nil
}

func recommendedActionFromString(action string) RecommendedAction {
	switch action {
	case "allow":
		return RecommendedActionAllow
	case "deny":
		return RecommendedActionDeny
	case "challenge":
		return RecommendedActionChallenge
	default:
		return RecommendedActionNone
	}
}
