package epayco

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// authTokenTTL bounds how long a bearer token is cached before it is re-fetched.
// ePayco's login response does not include an expiry, so this mirrors the 14
// minute heuristic used by the reference PHP SDK's cookie cache.
const authTokenTTL = 14 * time.Minute

type tokenEntry struct {
	value     string
	expiresAt time.Time
}

func (t *tokenEntry) valid() bool {
	return t != nil && t.value != "" && time.Now().Before(t.expiresAt)
}

// authCache holds the bearer tokens for the two distinct login flows used by the
// ePayco API: the core flow (public/private key in the JSON body) and the Apify
// flow (HTTP Basic auth), each backed by a separate login endpoint.
type authCache struct {
	mu    sync.Mutex
	core  *tokenEntry
	apify *tokenEntry
}

func newAuthCache() *authCache {
	return &authCache{}
}

// bearerToken returns a cached bearer token for the given auth flow, fetching and
// caching a new one if missing or expired.
func (c *Client) bearerToken(ctx context.Context, apify bool) (string, error) {
	c.auth.mu.Lock()
	entry := c.auth.core
	if apify {
		entry = c.auth.apify
	}
	if entry.valid() {
		token := entry.value
		c.auth.mu.Unlock()
		return token, nil
	}
	c.auth.mu.Unlock()

	token, err := c.login(ctx, apify)
	if err != nil {
		return "", err
	}

	fresh := &tokenEntry{value: token, expiresAt: time.Now().Add(authTokenTTL)}
	c.auth.mu.Lock()
	if apify {
		c.auth.apify = fresh
	} else {
		c.auth.core = fresh
	}
	c.auth.mu.Unlock()

	return token, nil
}

// invalidateToken clears a cached token, forcing the next bearerToken call to
// re-authenticate. Used after receiving an HTTP 401 from a resource endpoint,
// which indicates the cached token was rejected or expired early.
func (c *Client) invalidateToken(apify bool) {
	c.auth.mu.Lock()
	defer c.auth.mu.Unlock()
	if apify {
		c.auth.apify = nil
	} else {
		c.auth.core = nil
	}
}

type authResponse struct {
	BearerToken string `json:"bearer_token"`
	Token       string `json:"token"`
	Message     string `json:"message"`
	Error       string `json:"error"`
}

// login exchanges the client's credentials for a bearer token. The core flow POSTs
// {public_key, private_key} to baseURL+"/v1/auth/login"; the Apify flow (used by
// Safetypay and Daviplata) POSTs an empty body with HTTP Basic auth to
// apifyBaseURL+"/login" instead. Both return a bearer token used the same way
// afterwards.
func (c *Client) login(ctx context.Context, apify bool) (string, error) {
	url := c.baseURL + "/v1/auth/login"
	reqBody := []byte("{}")
	var basicAuth string

	if apify {
		url = c.apifyBaseURL + "/login"
		basicAuth = base64.StdEncoding.EncodeToString([]byte(c.apiKey + ":" + c.privateKey))
	} else {
		body, err := json.Marshal(map[string]string{
			"public_key":  c.apiKey,
			"private_key": c.privateKey,
		})
		if err != nil {
			return "", &Error{Code: ErrCodeUnknown, Message: "failed to encode login request", Err: err}
		}
		reqBody = body
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", &Error{Code: ErrCodeCommunication, Message: "failed to build login request", Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if basicAuth != "" {
		req.Header.Set("Authorization", "Basic "+basicAuth)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", &Error{Code: ErrCodeNoCommunication, Message: "failed to reach ePayco auth endpoint", Err: err}
	}
	defer resp.Body.Close()

	var parsed authResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", &Error{Code: ErrCodeAuthentication, Message: "failed to decode auth response", HTTPStatus: resp.StatusCode, Err: err}
	}

	message := parsed.Message
	if message == "" {
		message = parsed.Error
	}

	if resp.StatusCode < 200 || resp.StatusCode > 206 {
		if message == "" {
			message = fmt.Sprintf("authentication failed with status %d", resp.StatusCode)
		}
		return "", &Error{Code: ErrCodeAuthentication, Message: message, HTTPStatus: resp.StatusCode}
	}

	token := parsed.BearerToken
	if token == "" {
		token = parsed.Token
	}
	if token == "" {
		if message == "" {
			message = "ePayco did not return a bearer token"
		}
		return "", &Error{Code: ErrCodeAuthentication, Message: message, HTTPStatus: resp.StatusCode}
	}

	return token, nil
}
