package epayco

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSubscriptionCreate(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/subscription/create", func(w http.ResponseWriter, r *http.Request) {
		var got CreateSubscriptionParams
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got.IDPlan != "coursereact" || got.TokenCard != "id_token" {
			t.Errorf("unexpected request body: %+v", got)
		}
		writeJSON(t, w, `{"id_subscription":"sub_1"}`)
	})
	defer server.Close()

	result, err := c.Subscriptions.Create(context.Background(), CreateSubscriptionParams{
		IDPlan: "coursereact", Customer: "id_customer", TokenCard: "id_token",
		DocType: "CC", DocNumber: "5234567",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if string(result) != `{"id_subscription":"sub_1"}` {
		t.Errorf("result = %s", result)
	}
}

func TestSubscriptionGet(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/subscription/sub_1/public-key", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{"id_subscription":"sub_1"}`)
	})
	defer server.Close()

	if _, err := c.Subscriptions.Get(context.Background(), "sub_1"); err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestSubscriptionList(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/subscriptions/public-key", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `[{"id_subscription":"sub_1"}]`)
	})
	defer server.Close()

	if _, err := c.Subscriptions.List(context.Background()); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestSubscriptionCancel(t *testing.T) {
	c, server := newLegacyServer(t, "/recurring/v1/subscription/cancel", func(w http.ResponseWriter, r *http.Request) {
		var got struct {
			ID        string `json:"id"`
			PublicKey string `json:"public_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got.ID != "sub_1" || got.PublicKey != "public-key" {
			t.Errorf("unexpected request body: %+v", got)
		}
		writeJSON(t, w, `{"status":"cancelled"}`)
	})
	defer server.Close()

	if _, err := c.Subscriptions.Cancel(context.Background(), "sub_1"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
}

func TestSubscriptionCharge(t *testing.T) {
	c, server := newLegacyServer(t, "/payment/v1/charge/subscription/create", func(w http.ResponseWriter, r *http.Request) {
		var got ChargeSubscriptionParams
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got.IP != "190.000.000.000" {
			t.Errorf("unexpected request body: %+v", got)
		}
		writeJSON(t, w, `{"status":"Aceptada"}`)
	})
	defer server.Close()

	_, err := c.Subscriptions.Charge(context.Background(), ChargeSubscriptionParams{
		IDPlan: "coursereact", Customer: "id_customer", TokenCard: "id_token",
		DocType: "CC", DocNumber: "5234567", IP: "190.000.000.000",
	})
	if err != nil {
		t.Fatalf("Charge: %v", err)
	}
}
