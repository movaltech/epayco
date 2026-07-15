package epayco

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, baseURL, apifyURL string) *Client {
	t.Helper()
	c, err := New("public-key", "private-key", WithBaseURL(baseURL), WithApifyBaseURL(apifyURL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestBearerTokenIsCachedAndInvalidated(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/login" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		calls++
		_ = json.NewEncoder(w).Encode(map[string]string{"bearer_token": fmt.Sprintf("token-%d", calls)})
	}))
	defer server.Close()

	c := newTestClient(t, server.URL, server.URL)
	ctx := context.Background()

	token, err := c.bearerToken(ctx, false)
	if err != nil {
		t.Fatalf("bearerToken: %v", err)
	}
	if token != "token-1" {
		t.Errorf("token = %q, want token-1", token)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}

	token, err = c.bearerToken(ctx, false)
	if err != nil {
		t.Fatalf("bearerToken (cached): %v", err)
	}
	if token != "token-1" {
		t.Errorf("cached token = %q, want token-1 (should not re-authenticate)", token)
	}
	if calls != 1 {
		t.Fatalf("calls after cached read = %d, want 1", calls)
	}

	c.invalidateToken(false)
	token, err = c.bearerToken(ctx, false)
	if err != nil {
		t.Fatalf("bearerToken (after invalidate): %v", err)
	}
	if token != "token-2" {
		t.Errorf("token after invalidate = %q, want token-2", token)
	}
	if calls != 2 {
		t.Fatalf("calls after invalidate = %d, want 2", calls)
	}
}

func TestApifyLoginUsesBasicAuthAndEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("public-key:private-key"))
		if got := r.Header.Get("Authorization"); got != wantAuth {
			t.Errorf("Authorization header = %q, want %q", got, wantAuth)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "apify-token"})
	}))
	defer server.Close()

	c := newTestClient(t, "https://unused.example.com", server.URL)
	token, err := c.bearerToken(context.Background(), true)
	if err != nil {
		t.Fatalf("bearerToken: %v", err)
	}
	if token != "apify-token" {
		t.Errorf("token = %q, want apify-token", token)
	}

	// The core cache entry must remain untouched by an apify login.
	if c.auth.core != nil {
		t.Errorf("core token cache should be empty after an apify-only login")
	}
}

func TestLoginFailureReturnsTypedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "invalid credentials"})
	}))
	defer server.Close()

	c := newTestClient(t, server.URL, server.URL)
	_, err := c.bearerToken(context.Background(), false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	epErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if epErr.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("HTTPStatus = %d, want 401", epErr.HTTPStatus)
	}
	if epErr.Message != "invalid credentials" {
		t.Errorf("Message = %q, want %q", epErr.Message, "invalid credentials")
	}
}
