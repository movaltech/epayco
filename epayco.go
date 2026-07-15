// Package epayco is a professional, dependency-free Go SDK for the ePayco payment
// API (https://epayco.com). It is designed to be imported by any Go service that
// needs to create charges, tokenize cards, manage customers, subscriptions and
// plans, accept cash/PSE/Safetypay/Daviplata/standard-checkout payments, and
// verify confirmation webhooks.
package epayco

import (
	"net/http"
	"strings"
)

const (
	// defaultBaseURL is the legacy core API. It is only used for Plans and
	// Subscriptions (recurring billing): ePayco's official documentation
	// (docs.epayco.com/docs/planes) does not publish a REST contract for that
	// domain and directs integrators to the official SDKs instead, which all
	// target this host.
	defaultBaseURL = "https://api.secure.payco.co"
	// defaultApifyBaseURL is ePayco's current unified payments API (login,
	// card tokenization, customer management, charges, PSE, cash, Daviplata,
	// Safetypay, standard checkout, transaction lookup) — plain JSON over
	// Bearer auth, no payload encryption. Confirmed against the official
	// "API Services ePayco Producción" Postman collection and
	// docs.epayco.com/docs/api.
	defaultApifyBaseURL = "https://apify.epayco.co"
)

// Client is the entry point to the ePayco API. Create one with New and reuse it —
// it is safe for concurrent use by multiple goroutines.
type Client struct {
	apiKey     string
	privateKey string
	lang       string

	baseURL      string
	apifyBaseURL string

	httpClient *http.Client
	retries    int

	auth *authCache

	// apify-backed services (apify.epayco.co)
	Charges   *ChargeService
	Tokens    *TokenService
	Customers *CustomerService
	PSE       *PSEService
	Cash      *CashService
	Safetypay *SafetypayService
	Daviplata *DaviplataService
	Standard  *StandardService

	// legacy core services (api.secure.payco.co) — recurring billing only,
	// see PlanService's doc comment for why.
	Plans         *PlanService
	Subscriptions *SubscriptionService
}

// New creates a Client authenticated with the given ePayco public API key and
// private key (both available in the ePayco merchant dashboard). It does not make
// any network call — authentication happens lazily on the first request.
func New(apiKey, privateKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(privateKey) == "" {
		return nil, &Error{Code: ErrCodeInvalidConfig, Message: "apiKey and privateKey are required"}
	}

	c := &Client{
		apiKey:       apiKey,
		privateKey:   privateKey,
		lang:         "ES",
		baseURL:      defaultBaseURL,
		apifyBaseURL: defaultApifyBaseURL,
		httpClient:   http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.lang != "ES" && c.lang != "EN" {
		return nil, &Error{Code: ErrCodeInvalidConfig, Message: "lang must be \"ES\" or \"EN\", got " + c.lang}
	}

	c.auth = newAuthCache()

	c.Charges = &ChargeService{client: c}
	c.Tokens = &TokenService{client: c}
	c.Customers = &CustomerService{client: c}
	c.PSE = &PSEService{client: c}
	c.Cash = &CashService{client: c}
	c.Safetypay = &SafetypayService{client: c}
	c.Daviplata = &DaviplataService{client: c}
	c.Standard = &StandardService{client: c}
	c.Plans = &PlanService{client: c}
	c.Subscriptions = &SubscriptionService{client: c}

	return c, nil
}
