package epayco

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newLegacyServer starts a test server that answers POST /v1/auth/login (the
// legacy core auth flow) and serves handler for path, then returns a *Client
// wired to it. Used by plan.go/subscription.go tests, the only resources
// still on the legacy core API.
func newLegacyServer(t *testing.T, path string, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"bearer_token": "test-token"})
	})
	mux.HandleFunc(path, handler)
	server := httptest.NewServer(mux)

	c, err := New("public-key", "private-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, server
}

func TestPlanCreate(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/plan/create", func(w http.ResponseWriter, r *http.Request) {
		var got CreatePlanParams
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got.IDPlan != "coursereact" || got.Amount != 30000 {
			t.Errorf("unexpected request body: %+v", got)
		}
		writeJSON(t, w, `{"id_plan":"coursereact","status":"created"}`)
	})
	defer server.Close()

	result, err := c.Plans.Create(context.Background(), CreatePlanParams{
		IDPlan: "coursereact", Name: "Course react js", Amount: 30000,
		Currency: "cop", Interval: "month", IntervalCount: 1,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if decoded["id_plan"] != "coursereact" {
		t.Errorf("unexpected result: %+v", decoded)
	}
}

func TestPlanGet(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/plan/public-key/coursereact", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"id_plan":"coursereact"}`)
	})
	defer server.Close()

	result, err := c.Plans.Get(context.Background(), "coursereact")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(result) != `{"id_plan":"coursereact"}` {
		t.Errorf("result = %s", result)
	}
}

func TestPlanList(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/plans/public-key", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `[{"id_plan":"coursereact"}]`)
	})
	defer server.Close()

	result, err := c.Plans.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var decoded []map[string]string
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if len(decoded) != 1 || decoded[0]["id_plan"] != "coursereact" {
		t.Errorf("unexpected result: %+v", decoded)
	}
}

func TestPlanUpdate(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/plan/edit/coursereact", func(w http.ResponseWriter, r *http.Request) {
		var got UpdatePlanParams
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got.Amount != 11900 {
			t.Errorf("unexpected request body: %+v", got)
		}
		writeJSON(t, w, `{"status":"updated"}`)
	})
	defer server.Close()

	if _, err := c.Plans.Update(context.Background(), "coursereact", UpdatePlanParams{Amount: 11900}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestPlanDelete(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/plan/remove/public-key/coursereact", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"status":"deleted"}`)
	})
	defer server.Close()

	if _, err := c.Plans.Delete(context.Background(), "coursereact"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
