package castle

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// FilterEndpoint defines the filter URL castle.io side
var FilterEndpoint = "https://api.castle.io/v1/filter"

// RiskEndpoint defines the risk URL castle.io side
var RiskEndpoint = "https://api.castle.io/v1/risk"

// New creates a new castle client
func New(secret string) (*Castle, error) {
	client := &http.Client{}

	return NewWithHTTPClient(secret, client)
}

// NewWithHTTPClient same as New but allows passing of http.Client with custom config
func NewWithHTTPClient(secret string, client *http.Client) (*Castle, error) {
	return &Castle{client: client, apiSecret: secret}, nil
}

// Castle encapsulates http client
type Castle struct {
	client    *http.Client
	apiSecret string
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
	e := &castleAPIRequest{
		Type:         req.Event.EventType,
		Status:       req.Event.EventStatus,
		RequestToken: req.Context.RequestToken,
		User:         req.User,
		Context:      req.Context,
		Properties:   req.Properties,
	}
	return c.sendFilterCall(ctx, e)
}

// sendFilterCall is a plumbing method constructing the HTTP req/res and interpreting results
func (c *Castle) sendFilterCall(ctx context.Context, e *castleAPIRequest) (RecommendedAction, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(e)
	if err != nil {
		return RecommendedActionNone, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, FilterEndpoint, b)
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
	if expected, got := http.StatusCreated, res.StatusCode; expected != got {
		b, _ := io.ReadAll(res.Body) // nolint: errcheck
		return RecommendedActionNone, errors.Errorf("expected %d status but got %d: %s", expected, got, string(b))
	}

	resp := &castleAPIResponse{}
	if err = json.NewDecoder(res.Body).Decode(resp); err != nil {
		return RecommendedActionNone, err
	}

	if resp.Type != "" {
		// we have an api error
		return RecommendedActionNone, errors.New(resp.Type)
	}

	if resp.Message != "" {
		// we have an api error
		return RecommendedActionNone, errors.Errorf("%s: %s", resp.Type, resp.Message)
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

// Risk sends a risk request to castle.io
// see https://reference.castle.io/#operation/risk for details
func (c *Castle) Risk(ctx context.Context, req *Request) (RecommendedAction, error) {
	if req == nil {
		return RecommendedActionNone, errors.New("request cannot be nil")
	}
	if req.Context == nil {
		return RecommendedActionNone, errors.New("request.Context cannot be nil")
	}
	e := &castleAPIRequest{
		Type:         req.Event.EventType,
		Status:       req.Event.EventStatus,
		RequestToken: req.Context.RequestToken,
		User:         req.User,
		Context:      req.Context,
		Properties:   req.Properties,
	}
	return c.sendRiskCall(ctx, e)
}

// sendRiskCall is a plumbing method constructing the HTTP req/res and interpreting results
func (c *Castle) sendRiskCall(ctx context.Context, e *castleAPIRequest) (RecommendedAction, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(e)
	if err != nil {
		return RecommendedActionNone, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RiskEndpoint, b)
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
		return RecommendedActionNone, errors.Errorf("expected 201 status but got %d: %s", res.StatusCode, string(b))
	}

	resp := &castleAPIResponse{}
	if err = json.NewDecoder(res.Body).Decode(resp); err != nil {
		return RecommendedActionNone, errors.Errorf("unable to decode response body: %v", err)
	}

	if resp.Type != "" {
		// we have an api error
		return RecommendedActionNone, errors.New(resp.Type)
	}

	if resp.Message != "" {
		// we have an api error
		return RecommendedActionNone, errors.Errorf("%s: %s", resp.Type, resp.Message)
	}

	return recommendedActionFromString(resp.Policy.Action), nil
}
