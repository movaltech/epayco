// Command charge is a minimal, runnable example of the most common
// github.com/diegofxm/epayco-go operations: tokenize a card, charge it, look
// the transaction up by reference, and verify a confirmation webhook.
//
// Run it against ePayco's sandbox with:
//
//	EPAYCO_PUBLIC_KEY=... EPAYCO_PRIVATE_KEY=... go run ./examples/charge
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	epayco "github.com/diegofxm/epayco-go"
)

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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Tokenize a card. ePayco's own documented sandbox test card.
	token, err := client.Tokens.Create(ctx, epayco.CreateCardParams{
		CardNumber:   "4575623182290326",
		CardExpYear:  "2029",
		CardExpMonth: "12",
		CardCVC:      "123",
	})
	if err != nil {
		log.Fatalf("Tokens.Create: %v", err)
	}
	fmt.Printf("card tokenized: id=%s\n", token.ID)

	// 2. Charge the tokenized card.
	charge, err := client.Charges.Create(ctx, epayco.CreateChargeParams{
		Value: "50000", DocType: "CC", DocNumber: "1234567890",
		Name: "Juan", LastName: "Perez", Email: "juan.perez@example.com",
		CellPhone: "3010000000", Phone: "3010000000", Address: "Calle 1 # 2-3",
		CardTokenID: token.ID, Dues: "1",
	})
	if err != nil {
		log.Fatalf("Charges.Create: %v", err)
	}
	fmt.Printf("charge created: refPayco=%d state=%s response=%s\n", charge.RefPayco, charge.State, charge.Response)

	// 3. Look the transaction up by reference — useful for reconciling with
	// the confirmation webhook (step 4), or for polling a pending PSE/cash
	// payment until it settles.
	tx, err := client.Charges.Get(ctx, strconv.Itoa(charge.RefPayco))
	if err != nil {
		log.Fatalf("Charges.Get: %v", err)
	}
	fmt.Printf("transaction status: %s\n", tx.Status)

	// 4. Verify a confirmation webhook. In a real service this runs inside
	// your HTTP handler for the url_confirmation endpoint you configured;
	// here it's simulated with a synthetic request for illustration.
	demoWebhookVerification()
}

// demoWebhookVerification shows how to verify the x_signature ePayco sends to
// your confirmation URL. p_cust_id_cliente and p_key come from the ePayco
// dashboard (Integraciones -> Llaves secretas) — they are NOT the same as
// EPAYCO_PUBLIC_KEY/EPAYCO_PRIVATE_KEY used to authenticate API calls.
func demoWebhookVerification() {
	form := url.Values{
		"x_ref_payco":      {"12345"},
		"x_transaction_id": {"999"},
		"x_amount":         {"1000"},
		"x_currency_code":  {"COP"},
		"x_signature":      {"..."}, // whatever ePayco actually sent
	}
	req, _ := http.NewRequest(http.MethodPost, "/confirmation", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	payload, err := epayco.ParseWebhookPayload(req)
	if err != nil {
		log.Fatalf("ParseWebhookPayload: %v", err)
	}

	ok := epayco.VerifySignature(os.Getenv("EPAYCO_P_CUST_ID_CLIENTE"), os.Getenv("EPAYCO_P_KEY"), *payload)
	fmt.Printf("webhook signature valid: %v\n", ok)
}
