package epayco

import "net/http"

// Option configures a Client created with New.
type Option func(*Client)

// WithHTTPClient overrides the *http.Client used for every request. Useful for
// injecting timeouts, custom transports, or test doubles.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithBaseURL overrides the legacy core API base URL, used only for Plans and
// Subscriptions (recurring billing). Defaults to https://api.secure.payco.co.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithApifyBaseURL overrides the base URL used for ePayco's unified payments API
// (login, charges, PSE, cash, Daviplata, Safetypay, standard checkout, card
// tokenization, customer management, transaction lookup). Defaults to
// https://apify.epayco.co.
func WithApifyBaseURL(url string) Option {
	return func(c *Client) { c.apifyBaseURL = url }
}

// WithLang sets the language ("ES" or "EN") used for error messages returned by the
// ePayco API. Defaults to "ES".
func WithLang(lang string) Option {
	return func(c *Client) { c.lang = lang }
}

// WithRetries sets how many times a request is retried after a transient failure
// (network error or HTTP 5xx). Defaults to 0 (no retries).
func WithRetries(n int) Option {
	return func(c *Client) { c.retries = n }
}
