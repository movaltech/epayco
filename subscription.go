package epayco

import (
	"context"
	"encoding/json"
	"net/http"
)

// SubscriptionService manages recurring-billing subscriptions through
// ePayco's legacy core API (api.secure.payco.co/recurring/v1 and
// /payment/v1). See PlanService's doc comment for why this stays on the
// legacy API and why responses are returned as json.RawMessage.
type SubscriptionService struct {
	client *Client
}

// CreateSubscriptionParams are the fields accepted to subscribe a customer's
// card to a plan, confirmed against docs/reference/epayco-node/README.md.
type CreateSubscriptionParams struct {
	IDPlan             string `json:"id_plan"`
	Customer           string `json:"customer"`
	TokenCard          string `json:"token_card"`
	DocType            string `json:"doc_type"`
	DocNumber          string `json:"doc_number"`
	URLConfirmation    string `json:"url_confirmation,omitempty"`
	MethodConfirmation string `json:"method_confirmation,omitempty"`
}

// Create subscribes a customer's card to a plan via
// POST /recurring/v1/subscription/create.
func (s *SubscriptionService) Create(ctx context.Context, params CreateSubscriptionParams) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodPost, base: s.client.baseURL, path: "/recurring/v1/subscription/create",
		body: params, out: &result,
	})
	return result, err
}

// Get retrieves a subscription by id via
// GET /recurring/v1/subscription/{id}/{apiKey}.
func (s *SubscriptionService) Get(ctx context.Context, id string) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodGet, base: s.client.baseURL, path: "/recurring/v1/subscription/" + id + "/" + s.client.apiKey,
		out: &result,
	})
	return result, err
}

// List retrieves every subscription via GET /recurring/v1/subscriptions/{apiKey}.
func (s *SubscriptionService) List(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodGet, base: s.client.baseURL, path: "/recurring/v1/subscriptions/" + s.client.apiKey,
		out: &result,
	})
	return result, err
}

// Cancel cancels a subscription via POST /recurring/v1/subscription/cancel.
func (s *SubscriptionService) Cancel(ctx context.Context, id string) (json.RawMessage, error) {
	body := struct {
		ID        string `json:"id"`
		PublicKey string `json:"public_key"`
	}{ID: id, PublicKey: s.client.apiKey}

	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodPost, base: s.client.baseURL, path: "/recurring/v1/subscription/cancel",
		body: body, out: &result,
	})
	return result, err
}

// ChargeSubscriptionParams are the fields accepted to charge a subscribed
// customer's card for the current billing period.
type ChargeSubscriptionParams struct {
	IDPlan    string `json:"id_plan"`
	Customer  string `json:"customer"`
	TokenCard string `json:"token_card"`
	DocType   string `json:"doc_type"`
	DocNumber string `json:"doc_number"`
	IP        string `json:"ip"`
}

// Charge bills a subscribed customer's card via
// POST /payment/v1/charge/subscription/create.
func (s *SubscriptionService) Charge(ctx context.Context, params ChargeSubscriptionParams) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodPost, base: s.client.baseURL, path: "/payment/v1/charge/subscription/create",
		body: params, out: &result,
	})
	return result, err
}
