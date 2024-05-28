package castle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	FilterEndpoint = "https://api.castle.io/v1/filter"
	RiskEndpoint   = "https://api.castle.io/v1/risk"
)

// Castle encapsulates http client
type Castle struct {
	client    *http.Client
	apiSecret string
}

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status code: %d, message: %s", e.StatusCode, e.Message)
}

// New creates a new castle client with default http client
func New(secret string) (*Castle, error) {
	return NewWithHTTPClient(secret, http.DefaultClient)
}

// NewWithHTTPClient same as New but allows passing of http.Client with custom config
func NewWithHTTPClient(secret string, client *http.Client) (*Castle, error) {
	return &Castle{client: client, apiSecret: secret}, nil
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
	r := &castleAPIRequest{
		Type:         req.Event.EventType,
		Name:         req.Event.Name,
		Status:       req.Event.EventStatus,
		RequestToken: req.Context.RequestToken,
		User:         req.User,
		Context:      req.Context,
		Properties:   req.Properties,
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
	r := &castleAPIRequest{
		Type:         req.Event.EventType,
		Name:         req.Event.Name,
		Status:       req.Event.EventStatus,
		RequestToken: req.Context.RequestToken,
		User:         req.User,
		Context:      req.Context,
		Properties:   req.Properties,
	}
	return c.sendCall(ctx, r, RiskEndpoint)
}

func (c *Castle) sendCall(ctx context.Context, r *castleAPIRequest, url string) (RecommendedAction, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(r)
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
