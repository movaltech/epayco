# epayco-go

[![Go Reference](https://pkg.go.dev/badge/github.com/diegofxm/epayco-go.svg)](https://pkg.go.dev/github.com/diegofxm/epayco-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/diegofxm/epayco-go)](https://goreportcard.com/report/github.com/diegofxm/epayco-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A professional, dependency-free Go SDK for the [ePayco](https://epayco.com) payment API.

```go
go get github.com/diegofxm/epayco-go
```

## Quick start

```go
// Sandbox vs. production is determined entirely by which key pair you pass
// here — ePayco has no separate client-side "test mode" flag.
client, err := epayco.New(os.Getenv("EPAYCO_PUBLIC_KEY"), os.Getenv("EPAYCO_PRIVATE_KEY"))
if err != nil {
	log.Fatal(err)
}

charge, err := client.Charges.Create(ctx, epayco.CreateChargeParams{
	Value: "50000", DocType: "CC", DocNumber: "1234567890",
	Name: "Juan", LastName: "Perez", Email: "juan@example.com",
	CellPhone: "3010000000", Phone: "3010000000",
	CardNumber: "4575623182290326", CardExpYear: "2029", CardExpMonth: "12",
	CardCVC: "123", Dues: "1",
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(charge.RefPayco, charge.State)
```

See [`examples/charge`](examples/charge) for a complete, runnable example covering
tokenization, charging, transaction lookup and webhook verification.

## Two API surfaces

ePayco's own production API is not a single consistent surface, and this SDK reflects that
instead of hiding it:

| Client field | Backend | Used for |
|---|---|---|
| `Charges`, `Tokens`, `Customers`, `PSE`, `Cash`, `Daviplata`, `Safetypay`, `Standard` | `apify.epayco.co` | Charging, tokenizing cards, managing customers, PSE, cash, Daviplata, Safetypay, standard checkout, transaction lookup — everything ePayco documents as the current way to integrate (`docs.epayco.com/docs/api`). |
| `Plans`, `Subscriptions` | `api.secure.payco.co` | Recurring billing. ePayco's own docs (`docs.epayco.com/docs/planes`) don't publish a REST contract for this and point integrators to the official SDKs instead — this is that contract, ported to Go. |

Both flows authenticate automatically and transparently; you never call login yourself. See
[`docs/epayco-go-architecture.md`](../docs/epayco-go-architecture.md) in the parent repository
for the full rationale, including how this was corroborated against ePayco's production Postman
collection and confirmed live against the sandbox API.

## Resources

```go
client.Tokens.Create(ctx, epayco.CreateCardParams{...})           // tokenize a card
client.Customers.Create(ctx, epayco.CreateParams{...})            // create/attach a customer
client.Customers.Get(ctx, customerID)
client.Customers.List(ctx)

client.Charges.Create(ctx, epayco.CreateChargeParams{...})        // credit card
client.Charges.Get(ctx, refPayco)                                 // look up any transaction

client.PSE.Create(ctx, epayco.CreatePSEParams{...})
client.PSE.Confirm(ctx, epayco.ConfirmPSEParams{...})
client.PSE.ListBanks(ctx)

client.Cash.Create(ctx, epayco.CreateCashParams{...})              // Efecty and similar
client.Cash.ListEntities(ctx)

client.Daviplata.Create(ctx, epayco.CreateDaviplataParams{...})
client.Daviplata.Confirm(ctx, epayco.ConfirmDaviplataParams{...})  // with the OTP sent to the payer

client.Safetypay.Create(ctx, epayco.CreateSafetypayParams{...})

client.Standard.Create(ctx, epayco.CreateStandardParams{...})
client.Standard.Confirm(ctx, epayco.ConfirmStandardParams{...})

client.Plans.Create(ctx, epayco.CreatePlanParams{...})            // recurring billing
client.Subscriptions.Create(ctx, epayco.CreateSubscriptionParams{...})
```

`Plans`/`Subscriptions` methods return `json.RawMessage` instead of a typed struct: no source
(official docs, the production Postman collection, or the reference SDKs) documents that API's
response schema, so the SDK is honest about that instead of guessing. Decode it into your own type
once you've confirmed the shape against your account.

## Errors

Every operation returns `(*T, error)`; the error is always `*epayco.Error` when non-nil,
implementing `errors.Is`/`errors.As`:

```go
_, err := client.Charges.Create(ctx, params)
var epErr *epayco.Error
if errors.As(err, &epErr) {
	fmt.Println(epErr.Code, epErr.Message, epErr.HTTPStatus)
}
```

ePayco's apify API returns HTTP 200 even for validation and business-logic failures, signaling the
real outcome through a `success` field — the SDK already accounts for this, so a non-nil error
here always means the operation genuinely failed, not just that the HTTP status was non-2xx.

## Webhooks

```go
func confirmationHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := epayco.ParseWebhookPayload(r)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// p_cust_id_cliente and p_key come from the ePayco dashboard
	// (Integraciones -> Llaves secretas) — NOT the same as the
	// apiKey/privateKey used to authenticate API calls.
	if !epayco.VerifySignature(pCustIDCliente, pKey, *payload) {
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	// ... update your own records using payload.RefPayco / payload.TransactionState ...
	w.WriteHeader(http.StatusOK) // ePayco expects 200 within 30s; handle idempotently, it retries on failure
}
```

## Testing

```
go test ./...
```

runs the full unit test suite (60+ tests) against `httptest` fixtures captured from ePayco's
production Postman collection and sandbox API — no network access required.

A separate, opt-in integration suite hits the real sandbox API and is excluded from the command
above by a build tag:

```
EPAYCO_PUBLIC_KEY=... EPAYCO_PRIVATE_KEY=... go test -tags integration -run TestIntegration ./...
```

## Requirements

Go 1.26.4+. No external dependencies — only the standard library.
