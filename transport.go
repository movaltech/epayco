package epayco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// requestOptions configures a single call to the ePayco API. Resource methods
// (charge.go, customer.go, etc., added in later phases) build one of these and
// hand it to Client.do.
type requestOptions struct {
	method string // http.MethodGet or http.MethodPost
	base   string // c.baseURL, c.secureBaseURL or c.apifyBaseURL
	path   string
	apify  bool // true selects the Apify Basic-auth login flow for this call
	body   any  // marshaled as the JSON request body when non-nil
	out    any  // unmarshaled from a successful JSON response when non-nil
}

// do executes opts against the ePayco API. It resolves a bearer token (fetching
// and caching one on first use), sends the request, decodes a successful JSON
// response into opts.out, and maps non-2xx responses to a typed *Error.
//
// A stale token (HTTP 401) is always retried exactly once with a freshly fetched
// token, regardless of the client's configured retry budget — that is a
// correctness step, not a resilience knob. Network errors and HTTP 5xx responses
// are retried up to c.retries times with a short linear backoff; other 4xx
// responses are returned immediately.
func (c *Client) do(ctx context.Context, opts requestOptions) error {
	err := c.doOnce(ctx, opts)
	if err == nil {
		return nil
	}

	if isUnauthorized(err) {
		c.invalidateToken(opts.apify)
		if err = c.doOnce(ctx, opts); err == nil {
			return nil
		}
	}

	for attempt := 0; attempt < c.retries && isRetryable(err); attempt++ {
		select {
		case <-ctx.Done():
			return err
		case <-time.After(backoff(attempt)):
		}
		if err = c.doOnce(ctx, opts); err == nil {
			return nil
		}
	}

	return err
}

func (c *Client) doOnce(ctx context.Context, opts requestOptions) error {
	token, err := c.bearerToken(ctx, opts.apify)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if opts.body != nil {
		encoded, err := json.Marshal(opts.body)
		if err != nil {
			return &Error{Code: ErrCodeUnknown, Message: "failed to encode request body", Err: err}
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, opts.method, opts.base+opts.path, bodyReader)
	if err != nil {
		return &Error{Code: ErrCodeCommunication, Message: "failed to build request", Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("type", "sdk-jwt")
	req.Header.Set("lang", "GO")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &Error{Code: ErrCodeNoCommunication, Message: "failed to reach ePayco", Err: err}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Error{Code: ErrCodeCommunication, Message: "failed to read response body", HTTPStatus: resp.StatusCode, Err: err}
	}

	if resp.StatusCode < 200 || resp.StatusCode > 206 {
		return statusError(resp.StatusCode, extractServerMessage(raw))
	}

	if opts.out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, opts.out); err != nil {
			return &Error{Code: ErrCodeUnknown, Message: "failed to decode response body", HTTPStatus: resp.StatusCode, Err: err}
		}
	}

	return nil
}

// extractServerMessage best-effort parses the {"message": "..."}, {"error": "..."}
// or {"errors": ["..."]} envelopes returned by different ePayco endpoints.
func extractServerMessage(raw []byte) string {
	var body struct {
		Message string   `json:"message"`
		Error   string   `json:"error"`
		Errors  []string `json:"errors"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return ""
	}
	switch {
	case body.Message != "":
		return body.Message
	case body.Error != "":
		return body.Error
	case len(body.Errors) > 0:
		return body.Errors[0]
	default:
		return ""
	}
}

func isUnauthorized(err error) bool {
	var epErr *Error
	return errors.As(err, &epErr) && epErr.HTTPStatus == http.StatusUnauthorized
}

// isRetryable reports whether err is a transient failure worth retrying: a network
// error (no HTTPStatus set) or an HTTP 5xx response. Other HTTP 4xx responses are
// caused by the request itself and would fail again identically.
func isRetryable(err error) bool {
	var epErr *Error
	if errors.As(err, &epErr) {
		return epErr.HTTPStatus == 0 || epErr.HTTPStatus >= 500
	}
	return true
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt+1) * 200 * time.Millisecond
	const max = 2 * time.Second
	if d > max {
		return max
	}
	return d
}
