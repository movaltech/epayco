package epayco

import (
	"context"
	"net/http"
)

// CustomerService manages ePayco customers and the card tokens attached to
// them through the apify API (apify.epayco.co).
type CustomerService struct {
	client *Client
}

// CreateParams are the fields accepted by Create to register a customer
// and, optionally, attach a previously created card token (see TokenService).
type CreateParams struct {
	DocType          string `json:"docType"`
	DocNumber        string `json:"docNumber"`
	Name             string `json:"name"`
	LastName         string `json:"lastName"`
	Email            string `json:"email"`
	CellPhone        string `json:"cellPhone"`
	Phone            string `json:"phone"`
	RequireCardToken bool   `json:"requireCardToken"`
	CardTokenID      string `json:"cardTokenId,omitempty"`
	Address          string `json:"address,omitempty"`
	City             string `json:"city,omitempty"`
}

// Customer is the customer record returned by Create.
type Customer struct {
	Status  bool   `json:"status"`
	Success bool   `json:"success"`
	Type    string `json:"type"`
	Data    struct {
		Status      string `json:"status"`
		Description string `json:"description"`
		CustomerID  string `json:"customerId"`
		Name        string `json:"name"`
		Email       string `json:"email"`
	} `json:"data"`
	Object string `json:"object"`
}

// Create registers a customer and/or attaches a card token to it via
// POST /token/customer.
func (s *CustomerService) Create(ctx context.Context, params CreateParams) (*Customer, error) {
	var result Customer
	if err := s.client.doApify(ctx, http.MethodPost, "/token/customer", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateParams are the fields accepted by Update.
type UpdateParams struct {
	CustomerID string `json:"customerId"`
	Name       string `json:"name"`
}

// Update edits a customer's name via POST /subscriptions/customer/update.
func (s *CustomerService) Update(ctx context.Context, params UpdateParams) error {
	return s.client.doApify(ctx, http.MethodPost, "/subscriptions/customer/update", params, nil)
}

// CustomerCard is a card token attached to a customer.
type CustomerCard struct {
	Token     string `json:"token"`
	Franchise string `json:"franchise"`
	Mask      string `json:"mask"`
	Created   string `json:"created"`
	Default   bool   `json:"default"`
}

// CustomerDetail is the full record returned by Get, including attached cards.
type CustomerDetail struct {
	ID        string         `json:"id_customer"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	DocType   string         `json:"doc_type"`
	DocNumber string         `json:"doc_number"`
	Created   string         `json:"created"`
	Cards     []CustomerCard `json:"cards"`
}

// Get retrieves a customer and its attached cards via POST /subscriptions/customer.
func (s *CustomerService) Get(ctx context.Context, customerID string) (*CustomerDetail, error) {
	var wrapped struct {
		Data CustomerDetail `json:"data"`
	}
	body := struct {
		CustomerID string `json:"customerId"`
	}{CustomerID: customerID}
	if err := s.client.doApify(ctx, http.MethodPost, "/subscriptions/customer", body, &wrapped); err != nil {
		return nil, err
	}
	return &wrapped.Data, nil
}

// CustomerSummary is the summary form of a customer returned by List.
type CustomerSummary struct {
	ID      string `json:"id_customer"`
	Object  string `json:"object"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Created string `json:"created"`
}

// List retrieves every customer registered on the account via
// GET /subscriptions/customers.
func (s *CustomerService) List(ctx context.Context) ([]CustomerSummary, error) {
	var wrapped struct {
		Data []CustomerSummary `json:"data"`
	}
	if err := s.client.doApify(ctx, http.MethodGet, "/subscriptions/customers", nil, &wrapped); err != nil {
		return nil, err
	}
	return wrapped.Data, nil
}

// DeleteTokenParams identify the card token to remove from a customer.
type DeleteTokenParams struct {
	Franchise  string `json:"franchise"`
	Mask       string `json:"mask"`
	CustomerID string `json:"customerId"`
}

// DeleteToken removes a card token from a customer via
// POST /subscription/token/card/delete.
func (s *CustomerService) DeleteToken(ctx context.Context, params DeleteTokenParams) error {
	return s.client.doApify(ctx, http.MethodPost, "/subscription/token/card/delete", params, nil)
}

// AddTokenParams attach an existing card token to a customer.
type AddTokenParams struct {
	CardToken  string `json:"cardToken"`
	CustomerID string `json:"customerId"`
}

// AddToken attaches a card token to an existing customer via
// POST /subscriptions/customer/add/new/token.
func (s *CustomerService) AddToken(ctx context.Context, params AddTokenParams) error {
	return s.client.doApify(ctx, http.MethodPost, "/subscriptions/customer/add/new/token", params, nil)
}

// SetDefaultTokenParams identify the card token to promote to default.
type SetDefaultTokenParams struct {
	CardToken  string `json:"cardToken"`
	CustomerID string `json:"customerId"`
	Franchise  string `json:"franchise"`
	Mask       string `json:"mask"`
}

// SetDefaultToken makes an existing card the default one for a customer via
// POST /subscriptions/customer/add/new/token/default.
func (s *CustomerService) SetDefaultToken(ctx context.Context, params SetDefaultTokenParams) error {
	return s.client.doApify(ctx, http.MethodPost, "/subscriptions/customer/add/new/token/default", params, nil)
}
