package epayco

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// newAuthServer returns an httptest.Server that always answers
// POST /v1/auth/login with a bearer token, incrementing a counter each time so
// tests can assert on how many times the client re-authenticated.
func newAuthServer(t *testing.T, tokenPrefix string) (*httptest.Server, *int) {
	t.Helper()
	var mu sync.Mutex
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]string{"bearer_token": fmtToken(tokenPrefix, n)})
	}))
	return server, &calls
}

func fmtToken(prefix string, n int) string {
	return prefix + "-" + string(rune('0'+n))
}

func TestDoDecodesSuccessfulResponse(t *testing.T) {
	authServer, _ := newAuthServer(t, "tok")
	defer authServer.Close()

	resource := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok-1" {
			t.Errorf("Authorization = %q, want Bearer tok-1", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"value": "hello"})
	}))
	defer resource.Close()

	c := newTestClient(t, authServer.URL, authServer.URL)

	var out struct {
		Value string `json:"value"`
	}
	err := c.do(context.Background(), requestOptions{
		method: http.MethodGet,
		base:   resource.URL,
		path:   "/anything",
		out:    &out,
	})
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if out.Value != "hello" {
		t.Errorf("out.Value = %q, want hello", out.Value)
	}
}

func TestDoMapsHTTPErrorStatuses(t *testing.T) {
	authServer, _ := newAuthServer(t, "tok")
	defer authServer.Close()

	cases := []struct {
		status  int
		body    string
		wantMsg string
	}{
		{400, `{"message":"bad request"}`, "bad request"},
		{403, `{}`, "acceso prohibido, no tienes permisos para esta acción"},
		{404, `{}`, "la ruta solicitada no existe"},
		{405, `{}`, "método no permitido en esta ruta"},
	}

	for _, tc := range cases {
		tc := tc
		resource := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		}))

		c := newTestClient(t, authServer.URL, authServer.URL)
		err := c.do(context.Background(), requestOptions{
			method: http.MethodGet,
			base:   resource.URL,
			path:   "/anything",
		})
		resource.Close()

		if err == nil {
			t.Fatalf("status %d: expected error, got nil", tc.status)
		}
		epErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("status %d: expected *Error, got %T", tc.status, err)
		}
		if epErr.HTTPStatus != tc.status {
			t.Errorf("status %d: HTTPStatus = %d", tc.status, epErr.HTTPStatus)
		}
		if epErr.Message != tc.wantMsg {
			t.Errorf("status %d: Message = %q, want %q", tc.status, epErr.Message, tc.wantMsg)
		}
	}
}

func TestDoRefreshesTokenOnUnauthorized(t *testing.T) {
	authServer, _ := newAuthServer(t, "tok")
	defer authServer.Close()

	var mu sync.Mutex
	resourceCalls := 0
	resource := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		resourceCalls++
		mu.Unlock()

		if r.Header.Get("Authorization") == "Bearer tok-1" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "expired token"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"value": "hello"})
	}))
	defer resource.Close()

	c := newTestClient(t, authServer.URL, authServer.URL)

	var out struct {
		Value string `json:"value"`
	}
	err := c.do(context.Background(), requestOptions{
		method: http.MethodGet,
		base:   resource.URL,
		path:   "/anything",
		out:    &out,
	})
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if out.Value != "hello" {
		t.Errorf("out.Value = %q, want hello", out.Value)
	}
	if resourceCalls != 2 {
		t.Errorf("resourceCalls = %d, want 2 (one 401, one retry with a fresh token)", resourceCalls)
	}
}

func TestDoRetriesTransientServerErrors(t *testing.T) {
	authServer, _ := newAuthServer(t, "tok")
	defer authServer.Close()

	var mu sync.Mutex
	resourceCalls := 0
	resource := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		resourceCalls++
		n := resourceCalls
		mu.Unlock()

		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"value": "hello"})
	}))
	defer resource.Close()

	c, err := New("public-key", "private-key", WithBaseURL(authServer.URL), WithApifyBaseURL(authServer.URL), WithRetries(2))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var out struct {
		Value string `json:"value"`
	}
	err = c.do(context.Background(), requestOptions{
		method: http.MethodGet,
		base:   resource.URL,
		path:   "/anything",
		out:    &out,
	})
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if resourceCalls != 3 {
		t.Errorf("resourceCalls = %d, want 3", resourceCalls)
	}
}

func TestDoDoesNotRetryClientErrors(t *testing.T) {
	authServer, _ := newAuthServer(t, "tok")
	defer authServer.Close()

	var mu sync.Mutex
	resourceCalls := 0
	resource := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		resourceCalls++
		mu.Unlock()
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "bad request"})
	}))
	defer resource.Close()

	c, err := New("public-key", "private-key", WithBaseURL(authServer.URL), WithApifyBaseURL(authServer.URL), WithRetries(2))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	err = c.do(context.Background(), requestOptions{
		method: http.MethodGet,
		base:   resource.URL,
		path:   "/anything",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if resourceCalls != 1 {
		t.Errorf("resourceCalls = %d, want 1 (400 must not be retried)", resourceCalls)
	}
}
