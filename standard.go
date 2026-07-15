package epayco

import (
	"context"
	"net/http"
)

// StandardService creates and confirms "standard checkout" transactions
// (redirect-based checkout used by ePayco's own hosted-payment products)
// through ePayco's apify API. Included in v1 at the user's explicit request,
// even though it has no precedent in the reference SDKs.
type StandardService struct {
	client *Client
}

// CreateStandardParams are the fields accepted to start a standard checkout
// transaction.
type CreateStandardParams struct {
	Channel         string `json:"channel"`
	Value           string `json:"value"`
	DocType         string `json:"docType"`
	DocNumber       string `json:"docNumber"`
	Name            string `json:"name"`
	LastName        string `json:"lastName"`
	Email           string `json:"email"`
	CellPhone       string `json:"cellPhone"`
	Phone           string `json:"phone,omitempty"`
	IP              string `json:"ip,omitempty"`
	City            string `json:"city,omitempty"`
	Address         string `json:"address,omitempty"`
	Description     string `json:"description,omitempty"`
	Invoice         string `json:"_invoice,omitempty"`
	URLResponse     string `json:"urlResponse,omitempty"`
	URLConfirmation string `json:"urlConfirmation,omitempty"`
	// MethodConfimation matches ePayco's own (misspelled) field name exactly
	// as observed in the production Postman collection.
	MethodConfimation string `json:"methodConfimation,omitempty"`
	Extra1            string `json:"extra1,omitempty"`
	Extra2            string `json:"extra2,omitempty"`
	Extra3            string `json:"extra3,omitempty"`
	Extra4            string `json:"extra4,omitempty"`
}

// StandardTransaction is the transaction created by Create, pending
// confirmation via Confirm.
type StandardTransaction struct {
	RefEpayco         int    `json:"refEpayco"`
	Invoice           string `json:"invoice"`
	Description       string `json:"description"`
	Value             string `json:"value"`
	Tax               string `json:"tax"`
	TaxBase           string `json:"taxBase"`
	Currency          string `json:"currency"`
	Bank              string `json:"bank"`
	State             string `json:"state"`
	StateMessage      string `json:"stateMessage"`
	AuthorizationCode string `json:"authorizationCode"`
	Receipt           string `json:"receipt"`
	DateTime          string `json:"dateTime"`
	Channel           string `json:"channel"`
	ResponseCode      int    `json:"responseCode"`
	IP                string `json:"ip"`
	DocType           string `json:"docType"`
	DocNumber         string `json:"docNumber"`
	Name              string `json:"name"`
	LastName          string `json:"lastName"`
	Email             string `json:"email"`
	City              string `json:"city"`
	Address           string `json:"address"`
	URLRedirect       string `json:"urlRedirect"`
}

// Create starts a standard checkout transaction via
// POST /payment/process/standard.
func (s *StandardService) Create(ctx context.Context, params CreateStandardParams) (*StandardTransaction, error) {
	var result StandardTransaction
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/process/standard", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfirmStandardParams report the outcome of a standard checkout transaction
// back to ePayco for confirmation.
type ConfirmStandardParams struct {
	Channel                  string `json:"channel"`
	Value                    string `json:"value"`
	RefEpayco                string `json:"refEpayco"`
	CurrentStatusTransaction string `json:"currentStatusTransaction"`
	StatusTransaction        string `json:"statusTransaction"`
	AuthorizationCode        string `json:"authorizationCode"`
}

// StandardConfirmation is the confirmed state of a standard checkout
// transaction, returned by Confirm.
type StandardConfirmation struct {
	RefEpayco         int     `json:"refEpayco"`
	Invoice           string  `json:"invoice"`
	Description       string  `json:"description"`
	Value             float64 `json:"value"`
	Tax               float64 `json:"tax"`
	TaxBase           float64 `json:"taxBase"`
	Currency          string  `json:"currency"`
	Bank              string  `json:"bank"`
	State             string  `json:"state"`
	StateMessage      string  `json:"stateMessage"`
	AuthorizationCode string  `json:"authorizationCode"`
	Receipt           string  `json:"receipt"`
	DateTime          string  `json:"dateTime"`
	Channel           string  `json:"channel"`
	ResponseCode      int     `json:"responseCode"`
	IP                string  `json:"ip"`
	DocType           string  `json:"docType"`
	DocNumber         string  `json:"docNumber"`
	Name              string  `json:"name"`
	LastName          string  `json:"lastName"`
	Email             string  `json:"email"`
	City              string  `json:"city"`
	Address           string  `json:"address"`
}

// Confirm reports the final outcome of a standard checkout transaction via
// POST /payment/confirm/standard.
func (s *StandardService) Confirm(ctx context.Context, params ConfirmStandardParams) (*StandardConfirmation, error) {
	var result StandardConfirmation
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/confirm/standard", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
