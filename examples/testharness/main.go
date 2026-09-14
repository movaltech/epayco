// Command testharness is a throwaway browser-based test bench for
// github.com/movaltech/epayco: a tiny Go server wraps every apify resource
// method behind a JSON route, and a single static page (index.html) offers a
// form per method so you can exercise the real ePayco sandbox by hand and
// see the raw response — no separate backend project, no JS framework.
//
// It exists to close the live-validation gap noted in
// docs/epayco-go-architecture.md: only Charges.Create, Charges.Get and
// PSE.ListBanks were confirmed against the real sandbox API before this.
//
// Run it with:
//
//	EPAYCO_PUBLIC_KEY=... EPAYCO_PRIVATE_KEY=... go run ./examples/testharness
//
// then open http://localhost:8080.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	epayco "github.com/movaltech/epayco"
)

//go:embed index.html
var staticFiles embed.FS

func main() {
	// Sandbox vs. production is determined entirely by which key pair you
	// pass here — ePayco has no separate client-side "test mode" flag.
	client, err := epayco.New(
		os.Getenv("EPAYCO_PUBLIC_KEY"),
		os.Getenv("EPAYCO_PRIVATE_KEY"),
	)
	if err != nil {
		log.Fatalf("epayco.New: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticFiles)))

	// Tokens
	mux.HandleFunc("/api/token/create", handle(func(ctx context.Context, req epayco.CreateCardParams) (any, error) {
		return client.Tokens.Create(ctx, req)
	}))

	// Customers
	mux.HandleFunc("/api/customer/create", handle(func(ctx context.Context, req epayco.CreateParams) (any, error) {
		return client.Customers.Create(ctx, req)
	}))
	mux.HandleFunc("/api/customer/update", handle(func(ctx context.Context, req epayco.UpdateParams) (any, error) {
		return "ok", client.Customers.Update(ctx, req)
	}))
	mux.HandleFunc("/api/customer/get", handle(func(ctx context.Context, req customerIDRequest) (any, error) {
		return client.Customers.Get(ctx, req.CustomerID)
	}))
	mux.HandleFunc("/api/customer/list", handle(func(ctx context.Context, _ struct{}) (any, error) {
		return client.Customers.List(ctx)
	}))
	mux.HandleFunc("/api/customer/delete-token", handle(func(ctx context.Context, req epayco.DeleteTokenParams) (any, error) {
		return "ok", client.Customers.DeleteToken(ctx, req)
	}))
	mux.HandleFunc("/api/customer/add-token", handle(func(ctx context.Context, req epayco.AddTokenParams) (any, error) {
		return "ok", client.Customers.AddToken(ctx, req)
	}))
	mux.HandleFunc("/api/customer/set-default-token", handle(func(ctx context.Context, req epayco.SetDefaultTokenParams) (any, error) {
		return "ok", client.Customers.SetDefaultToken(ctx, req)
	}))

	// Charges
	mux.HandleFunc("/api/charge/create", handle(func(ctx context.Context, req epayco.CreateChargeParams) (any, error) {
		return client.Charges.Create(ctx, req)
	}))
	mux.HandleFunc("/api/charge/get", handle(func(ctx context.Context, req refPaycoRequest) (any, error) {
		return client.Charges.Get(ctx, req.ReferencePayco)
	}))

	// PSE
	mux.HandleFunc("/api/pse/create", handle(func(ctx context.Context, req epayco.CreatePSEParams) (any, error) {
		return client.PSE.Create(ctx, req)
	}))
	mux.HandleFunc("/api/pse/confirm", handle(func(ctx context.Context, req epayco.ConfirmPSEParams) (any, error) {
		return client.PSE.Confirm(ctx, req)
	}))
	mux.HandleFunc("/api/pse/banks", handle(func(ctx context.Context, _ struct{}) (any, error) {
		return client.PSE.ListBanks(ctx)
	}))

	// Cash
	mux.HandleFunc("/api/cash/create", handle(func(ctx context.Context, req epayco.CreateCashParams) (any, error) {
		return client.Cash.Create(ctx, req)
	}))
	mux.HandleFunc("/api/cash/entities", handle(func(ctx context.Context, _ struct{}) (any, error) {
		return client.Cash.ListEntities(ctx)
	}))

	// Daviplata
	mux.HandleFunc("/api/daviplata/create", handle(func(ctx context.Context, req epayco.CreateDaviplataParams) (any, error) {
		return client.Daviplata.Create(ctx, req)
	}))
	mux.HandleFunc("/api/daviplata/confirm", handle(func(ctx context.Context, req epayco.ConfirmDaviplataParams) (any, error) {
		return client.Daviplata.Confirm(ctx, req)
	}))

	// Safetypay
	mux.HandleFunc("/api/safetypay/create", handle(func(ctx context.Context, req epayco.CreateSafetypayParams) (any, error) {
		return client.Safetypay.Create(ctx, req)
	}))

	// Standard checkout
	mux.HandleFunc("/api/standard/create", handle(func(ctx context.Context, req epayco.CreateStandardParams) (any, error) {
		return client.Standard.Create(ctx, req)
	}))
	mux.HandleFunc("/api/standard/confirm", handle(func(ctx context.Context, req epayco.ConfirmStandardParams) (any, error) {
		return client.Standard.Confirm(ctx, req)
	}))

	// Webhook: receives ePayco's real confirmation callback (set this
	// handler's public URL as urlConfirmation on any Create call above) and
	// verifies its signature. p_cust_id_cliente/p_key are dashboard
	// credentials (Integraciones -> Llaves secretas) distinct from
	// EPAYCO_PUBLIC_KEY/EPAYCO_PRIVATE_KEY used to authenticate API calls.
	pCustIDCliente := os.Getenv("EPAYCO_P_CUST_ID_CLIENTE")
	pKey := os.Getenv("EPAYCO_P_KEY")
	mux.HandleFunc("/webhook/confirmation", handleWebhook(pCustIDCliente, pKey))
	mux.HandleFunc("/webhook/last", handleWebhookLast)

	addr := "localhost:8080"
	log.Printf("testharness listening on http://%s (sandbox mode)", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

var (
	lastWebhookMu      sync.Mutex
	lastWebhookPayload *epayco.WebhookPayload
	lastWebhookValid   bool
)

// handleWebhook receives ePayco's confirmation callback, parses and verifies
// it, logs the full result to stdout (visible wherever this server's output
// is redirected), and answers 200 OK immediately — ePayco expects a response
// within 30s and retries on failure.
func handleWebhook(pCustIDCliente, pKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := epayco.ParseWebhookPayload(r)
		if err != nil {
			log.Printf("webhook: failed to parse: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		valid := epayco.VerifySignature(pCustIDCliente, pKey, *payload)

		lastWebhookMu.Lock()
		lastWebhookPayload = payload
		lastWebhookValid = valid
		lastWebhookMu.Unlock()

		raw, _ := json.MarshalIndent(payload, "", "  ")
		log.Printf("webhook received (method=%s) signature_valid=%v\n%s", r.Method, valid, raw)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}

// handleWebhookLast exposes the most recently received webhook (and its
// signature verification result) as JSON, so index.html can display it
// without watching server logs.
func handleWebhookLast(w http.ResponseWriter, r *http.Request) {
	lastWebhookMu.Lock()
	payload, valid := lastWebhookPayload, lastWebhookValid
	lastWebhookMu.Unlock()

	if payload == nil {
		writeResult(w, nil, nil)
		return
	}
	writeResult(w, map[string]any{"payload": payload, "signatureValid": valid}, nil)
}

type customerIDRequest struct {
	CustomerID string `json:"customerId"`
}

type refPaycoRequest struct {
	ReferencePayco string `json:"referencePayco"`
}

// handle decodes the request body as T (an empty body decodes to T's zero
// value, so struct{} works for no-argument calls), invokes fn, and writes the
// result as {"ok":true,"data":...} or {"ok":false,"error":"..."}.
func handle[T any](fn func(ctx context.Context, req T) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req T
		if r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeResult(w, nil, err)
				return
			}
		}
		data, err := fn(r.Context(), req)
		writeResult(w, data, err)
	}
}

func writeResult(w http.ResponseWriter, data any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}
