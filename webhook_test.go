package epayco

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// validSignature is sha256("cust123^key456^12345^999^1000^COP") in hex,
// precomputed independently (Node's crypto.createHash) so this test proves
// VerifySignature implements the exact byte-for-byte formula, not just that
// it's internally consistent with itself.
const validSignature = "69b86c8d29817aba051aabae0b63ce4fd1f808950c269a6e200bfd67260ef44a"

func TestVerifySignatureAccepted(t *testing.T) {
	payload := WebhookPayload{
		RefPayco:      "12345",
		TransactionID: "999",
		Amount:        "1000",
		CurrencyCode:  "COP",
		Signature:     validSignature,
	}
	if !VerifySignature("cust123", "key456", payload) {
		t.Fatal("expected signature to be accepted")
	}
}

func TestVerifySignatureRejectsTamperedAmount(t *testing.T) {
	payload := WebhookPayload{
		RefPayco:      "12345",
		TransactionID: "999",
		Amount:        "999999", // tampered: signature was computed for amount 1000
		CurrencyCode:  "COP",
		Signature:     validSignature,
	}
	if VerifySignature("cust123", "key456", payload) {
		t.Fatal("expected signature to be rejected for a tampered amount")
	}
}

func TestVerifySignatureRejectsWrongKey(t *testing.T) {
	payload := WebhookPayload{
		RefPayco:      "12345",
		TransactionID: "999",
		Amount:        "1000",
		CurrencyCode:  "COP",
		Signature:     validSignature,
	}
	if VerifySignature("cust123", "wrong-key", payload) {
		t.Fatal("expected signature to be rejected for a mismatched pKey")
	}
}

func TestParseWebhookPayloadFromPOSTForm(t *testing.T) {
	form := url.Values{
		"x_ref_payco":       {"12345"},
		"x_transaction_id":  {"999"},
		"x_amount":          {"1000"},
		"x_currency_code":   {"COP"},
		"x_signature":       {validSignature},
		"x_response":        {"Aceptada"},
		"x_cust_id_cliente": {"cust123"},
		"x_customer_email":  {"buyer@example.com"},
		"x_extra1":          {"order-42"},
		"x_extra10":         {"last-extra"},
	}
	req := httptest.NewRequest(http.MethodPost, "/confirmation", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	payload, err := ParseWebhookPayload(req)
	if err != nil {
		t.Fatalf("ParseWebhookPayload: %v", err)
	}
	if payload.RefPayco != "12345" || payload.TransactionID != "999" || payload.Amount != "1000" {
		t.Errorf("unexpected core fields: %+v", payload)
	}
	if payload.CustomerEmail != "buyer@example.com" {
		t.Errorf("CustomerEmail = %q", payload.CustomerEmail)
	}
	if payload.Extras[0] != "order-42" || payload.Extras[9] != "last-extra" {
		t.Errorf("unexpected extras: %+v", payload.Extras)
	}
	if !VerifySignature("cust123", "key456", *payload) {
		t.Errorf("expected signature parsed from form to verify")
	}
}

func TestParseWebhookPayloadFromGETQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/confirmation?x_ref_payco=12345&x_transaction_id=999&x_amount=1000&x_currency_code=COP&x_signature="+validSignature, nil)

	payload, err := ParseWebhookPayload(req)
	if err != nil {
		t.Fatalf("ParseWebhookPayload: %v", err)
	}
	if !VerifySignature("cust123", "key456", *payload) {
		t.Errorf("expected signature parsed from query string to verify")
	}
}
