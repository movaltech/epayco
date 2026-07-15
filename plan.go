package epayco

import (
	"context"
	"encoding/json"
	"net/http"
)

// PlanService manages recurring-billing plans through ePayco's legacy core
// API (api.secure.payco.co/recurring/v1). This is the only sanctioned way to
// manage plans: docs.epayco.com/docs/planes publishes no REST contract for
// them and directs integrators to the official SDKs instead, which all
// target this host — see docs/epayco-go-architecture.md section 3.
//
// Unlike the apify resources, no source available (official docs, the
// production Postman collection, or the reference SDKs themselves) documents
// a response schema for this API — the reference SDKs just forward whatever
// ePayco returns. Every method here does the same and returns the decoded
// JSON as json.RawMessage instead of a fabricated struct; unmarshal it into
// your own type once you've confirmed the shape against your account.
type PlanService struct {
	client *Client
}

// CreatePlanParams are the fields accepted to create a recurring-billing
// plan, confirmed against docs/reference/epayco-node/README.md (the official
// SDKs are the only documented contract for this API).
type CreatePlanParams struct {
	IDPlan                     string  `json:"id_plan"`
	Name                       string  `json:"name"`
	Description                string  `json:"description,omitempty"`
	Amount                     float64 `json:"amount"`
	Currency                   string  `json:"currency"`
	Interval                   string  `json:"interval"`
	IntervalCount              int     `json:"interval_count"`
	TrialDays                  int     `json:"trial_days,omitempty"`
	Tax                        float64 `json:"iva,omitempty"`
	Ico                        float64 `json:"ico,omitempty"`
	PlanLink                   string  `json:"planLink,omitempty"`
	GreetMessage               string  `json:"greetMessage,omitempty"`
	LinkExpirationDate         string  `json:"linkExpirationDate,omitempty"`
	SubscriptionLimit          int     `json:"subscriptionLimit,omitempty"`
	ImgURL                     string  `json:"imgUrl,omitempty"`
	DiscountValue              float64 `json:"discountValue,omitempty"`
	DiscountPercentage         float64 `json:"discountPercentage,omitempty"`
	TransactionalLimit         int     `json:"transactionalLimit,omitempty"`
	AdditionalChargePercentage float64 `json:"additionalChargePercentage,omitempty"`
	FirstPaymentAdditionalCost float64 `json:"firstPaymentAdditionalCost,omitempty"`
}

// UpdatePlanParams are the fields accepted to edit an existing plan.
type UpdatePlanParams struct {
	Name                       string  `json:"name,omitempty"`
	Description                string  `json:"description,omitempty"`
	Amount                     float64 `json:"amount,omitempty"`
	Currency                   string  `json:"currency,omitempty"`
	Interval                   string  `json:"interval,omitempty"`
	IntervalCount              int     `json:"interval_count,omitempty"`
	TrialDays                  int     `json:"trial_days,omitempty"`
	IP                         string  `json:"ip,omitempty"`
	Tax                        float64 `json:"iva,omitempty"`
	Ico                        float64 `json:"ico,omitempty"`
	TransactionalLimit         int     `json:"transactionalLimit,omitempty"`
	AdditionalChargePercentage float64 `json:"additionalChargePercentage,omitempty"`
	AfterPayment               string  `json:"afterPayment,omitempty"`
}

// Create registers a recurring-billing plan via POST /recurring/v1/plan/create.
func (s *PlanService) Create(ctx context.Context, params CreatePlanParams) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodPost, base: s.client.baseURL, path: "/recurring/v1/plan/create",
		body: params, out: &result,
	})
	return result, err
}

// Get retrieves a plan by id via GET /recurring/v1/plan/{apiKey}/{id}.
func (s *PlanService) Get(ctx context.Context, id string) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodGet, base: s.client.baseURL, path: "/recurring/v1/plan/" + s.client.apiKey + "/" + id,
		out: &result,
	})
	return result, err
}

// List retrieves every plan via GET /recurring/v1/plans/{apiKey}.
func (s *PlanService) List(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodGet, base: s.client.baseURL, path: "/recurring/v1/plans/" + s.client.apiKey,
		out: &result,
	})
	return result, err
}

// Update edits a plan via POST /recurring/v1/plan/edit/{id}.
func (s *PlanService) Update(ctx context.Context, id string, params UpdatePlanParams) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodPost, base: s.client.baseURL, path: "/recurring/v1/plan/edit/" + id,
		body: params, out: &result,
	})
	return result, err
}

// Delete removes a plan via POST /recurring/v1/plan/remove/{apiKey}/{id}.
func (s *PlanService) Delete(ctx context.Context, id string) (json.RawMessage, error) {
	var result json.RawMessage
	err := s.client.do(ctx, requestOptions{
		method: http.MethodPost, base: s.client.baseURL, path: "/recurring/v1/plan/remove/" + s.client.apiKey + "/" + id,
		out: &result,
	})
	return result, err
}
