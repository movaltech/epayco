//go:build integration

// Live smoke tests against ePayco's real sandbox API. Excluded from
// `go test ./...` by the "integration" build tag — run explicitly with:
//
//	EPAYCO_PUBLIC_KEY=... EPAYCO_PRIVATE_KEY=... go test -tags integration -run Integration ./...
//
// These exist to confirm (or correct) response shapes that no available
// source documents precisely — see the doc comment on ChargeTransaction in
// charge.go — by hitting the real API instead of a synthetic httptest
// fixture. They also double as a smoke test of the auth flow end to end.
package epayco_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	epayco "github.com/movaltech/epayco"
)

func liveClient(t *testing.T) *epayco.Client {
	t.Helper()
	apiKey := os.Getenv("EPAYCO_PUBLIC_KEY")
	privateKey := os.Getenv("EPAYCO_PRIVATE_KEY")
	if apiKey == "" || privateKey == "" {
		t.Skip("EPAYCO_PUBLIC_KEY/EPAYCO_PRIVATE_KEY not set, skipping live integration test")
	}

	client, err := epayco.New(apiKey, privateKey)
	if err != nil {
		t.Fatalf("epayco.New: %v", err)
	}
	return client
}

// TestIntegrationChargeCreateAndGet tokenizes and charges ePayco's own
// documented sandbox test card (4575623182290326, taken from the production
// Postman collection's own request examples), then looks the transaction up
// by reference. This is the primary target of this test file: confirming
// ChargeTransaction's real field names.
func TestIntegrationChargeCreateAndGet(t *testing.T) {
	client := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	charge, err := client.Charges.Create(ctx, epayco.CreateChargeParams{
		Value: "50000", DocType: "CC", DocNumber: "1234567890",
		Name: "Juan", LastName: "Perez", Email: "juan.perez@example.com",
		CellPhone: "3010000000", Phone: "3010000000", Address: "Calle 1 # 2-3",
		CardNumber: "4575623182290326", CardExpYear: "2029", CardExpMonth: "12",
		CardCVC: "123", Dues: "1",
	})
	if err != nil {
		t.Fatalf("Charges.Create: %v", err)
	}
	t.Logf("Charges.Create result: %+v", charge)

	if charge.RefPayco == 0 {
		t.Error("RefPayco is zero — ChargeTransaction may not match the real response shape")
	}

	tx, err := client.Charges.Get(ctx, strconv.Itoa(charge.RefPayco))
	if err != nil {
		t.Fatalf("Charges.Get: %v", err)
	}
	t.Logf("Charges.Get result: %+v", tx)
}

// TestIntegrationPSEListBanks is a read-only call with no side effects, used
// to confirm the Bank struct and the legacy-vs-apify auth wiring end to end.
func TestIntegrationPSEListBanks(t *testing.T) {
	client := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	banks, err := client.PSE.ListBanks(ctx)
	if err != nil {
		t.Fatalf("PSE.ListBanks: %v", err)
	}
	if len(banks) == 0 {
		t.Error("expected at least one bank")
	}
	t.Logf("got %d banks, first: %+v", len(banks), banks[0])
}
