package epayco

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newApifyServer starts a test server that answers POST /login (the Apify auth
// flow) and serves handler for path, then returns a *Client wired to it. It is
// reused by every apify-backed resource's tests (token, customer, pse, cash,
// daviplata, safetypay, standard, charge).
func newApifyServer(t *testing.T, path string, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
	})
	mux.HandleFunc(path, handler)
	server := httptest.NewServer(mux)

	c, err := New("public-key", "private-key", WithApifyBaseURL(server.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, server
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatalf("write response: %v", err)
	}
}

func TestDoApifyDecodesSuccess(t *testing.T) {
	c, server := newApifyServer(t, "/anything", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"success":true,"titleResponse":"SUCCESS","textResponse":"ok","lastAction":"x","data":{"value":"hello"}}`)
	})
	defer server.Close()

	var out struct {
		Value string `json:"value"`
	}
	if err := c.doApify(context.Background(), http.MethodPost, "/anything", nil, &out); err != nil {
		t.Fatalf("doApify: %v", err)
	}
	if out.Value != "hello" {
		t.Errorf("out.Value = %q, want hello", out.Value)
	}
}

func TestDoApifyMapsValidationErrors(t *testing.T) {
	c, server := newApifyServer(t, "/anything", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": false,
			"titleResponse": "Error",
			"textResponse": "Some fields are required, please correct the errors and try again",
			"lastAction": "validation data",
			"data": {
				"totalErrors": 2,
				"errors": [
					{"codError": 500, "errorMessage": "field cardNumber required"},
					{"codError": 500, "errorMessage": "field cardCvc required"}
				]
			}
		}`)
	})
	defer server.Close()

	err := c.doApify(context.Background(), http.MethodPost, "/anything", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	epErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if epErr.Message != "field cardNumber required" {
		t.Errorf("Message = %q, want the first validation error", epErr.Message)
	}
}

func TestDoApifyMapsStringBusinessError(t *testing.T) {
	c, server := newApifyServer(t, "/anything", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": false,
			"titleResponse": "Error customer",
			"textResponse": "Error customer",
			"lastAction": "customer",
			"data": {"error": "Error valid customer"}
		}`)
	})
	defer server.Close()

	err := c.doApify(context.Background(), http.MethodPost, "/anything", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	epErr := err.(*Error)
	if epErr.Message != "Error valid customer" {
		t.Errorf("Message = %q, want %q", epErr.Message, "Error valid customer")
	}
}

// TestDoApifyMapsErrorsFieldAsPlainString reproduces a real bug found live
// against the sandbox API: Customers.Update on a failed update returns
// titleResponse:"Success" and textResponse:"Customer updated successfully"
// (misleadingly implying success) with data.errors as a bare string (not the
// {codError,errorMessage} array shape) holding the real reason. The most
// specific detail (data.errors, then data.description) must win over the
// contradictory titleResponse/textResponse.
func TestDoApifyMapsErrorsFieldAsPlainString(t *testing.T) {
	c, server := newApifyServer(t, "/anything", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": false,
			"titleResponse": "Success",
			"textResponse": "Customer updated successfully",
			"lastAction": "update customer",
			"data": {
				"status": "error",
				"description": "Los datos son erroneos o son requeridos por favor compruebe.",
				"errors": "El campo last_name es requerido"
			}
		}`)
	})
	defer server.Close()

	err := c.doApify(context.Background(), http.MethodPost, "/anything", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	epErr := err.(*Error)
	if epErr.Message != "El campo last_name es requerido" {
		t.Errorf("Message = %q, want the specific validation reason, not titleResponse/textResponse", epErr.Message)
	}
}

// TestDoApifyMapsDescriptionWithoutErrors covers the case where "errors" is
// absent but "description" is present.
func TestDoApifyMapsDescriptionWithoutErrors(t *testing.T) {
	c, server := newApifyServer(t, "/anything", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": false,
			"titleResponse": "Error",
			"textResponse": "Error",
			"lastAction": "x",
			"data": {"status": "error", "description": "Los datos son erroneos o son requeridos por favor compruebe."}
		}`)
	})
	defer server.Close()

	err := c.doApify(context.Background(), http.MethodPost, "/anything", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	epErr := err.(*Error)
	if epErr.Message != "Los datos son erroneos o son requeridos por favor compruebe." {
		t.Errorf("Message = %q, want the description field", epErr.Message)
	}
}

func TestDoApifyMapsDescribedBusinessError(t *testing.T) {
	c, server := newApifyServer(t, "/anything", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": false,
			"titleResponse": "Error customer",
			"textResponse": "Error customer Cliente ya asociado o token inexistente",
			"lastAction": "create_customer",
			"data": {"error": {"status": "error", "description": "El token no se puede asociar al cliente"}}
		}`)
	})
	defer server.Close()

	err := c.doApify(context.Background(), http.MethodPost, "/anything", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	epErr := err.(*Error)
	if epErr.Message != "El token no se puede asociar al cliente" {
		t.Errorf("Message = %q, want the description field", epErr.Message)
	}
}

func TestDoApifyPropagatesHTTPErrors(t *testing.T) {
	c, server := newApifyServer(t, "/anything", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	err := c.doApify(context.Background(), http.MethodPost, "/anything", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	epErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if epErr.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %d, want 500", epErr.HTTPStatus)
	}
}
