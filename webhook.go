package epayco

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

// WebhookPayload is the data ePayco sends to a merchant's confirmation URL
// when a transaction's status changes. Field names and the signature formula
// below are corroborated against docs.epayco.com/docs/url-de-confirmacion and
// docs.epayco.com/docs/checkout-respuesta-y-confirmacion — no reference SDK
// implements this (see docs/epayco-go-architecture.md section 8).
type WebhookPayload struct {
	RefPayco           string
	TransactionID      string
	Response           string
	ResponseReasonText string
	Amount             string
	AmountCountry      string
	AmountOk           string
	Tax                string
	AmountBase         string
	CurrencyCode       string
	ApprovalCode       string
	TransactionDate    string
	TransactionState   string
	Franchise          string
	Description        string
	InvoiceID          string

	CustIDCliente    string
	CustomerDocType  string
	CustomerDocument string
	CustomerName     string
	CustomerLastname string
	CustomerEmail    string
	CustomerPhone    string
	CustomerCountry  string
	CustomerCity     string
	CustomerAddress  string
	CustomerIP       string

	Signature   string
	TestRequest string
	Business    string
	// Extras[0] is x_extra1 ... Extras[9] is x_extra10.
	Extras [10]string
}

// ParseWebhookPayload extracts a WebhookPayload from an incoming confirmation
// request. It reads through r.FormValue, which covers both confirmation
// methods ePayco supports (method_confirmation GET or POST): the URL query
// string for GET, and the form-encoded body for POST.
func ParseWebhookPayload(r *http.Request) (*WebhookPayload, error) {
	if err := r.ParseForm(); err != nil {
		return nil, &Error{Code: ErrCodeInvalidData, Message: "failed to parse webhook request", Err: err}
	}

	p := &WebhookPayload{
		RefPayco:           r.FormValue("x_ref_payco"),
		TransactionID:      r.FormValue("x_transaction_id"),
		Response:           r.FormValue("x_response"),
		ResponseReasonText: r.FormValue("x_response_reason_text"),
		Amount:             r.FormValue("x_amount"),
		AmountCountry:      r.FormValue("x_amount_country"),
		AmountOk:           r.FormValue("x_amount_ok"),
		Tax:                r.FormValue("x_tax"),
		AmountBase:         r.FormValue("x_amount_base"),
		CurrencyCode:       r.FormValue("x_currency_code"),
		ApprovalCode:       r.FormValue("x_approval_code"),
		TransactionDate:    r.FormValue("x_transaction_date"),
		TransactionState:   r.FormValue("x_transaction_state"),
		Franchise:          r.FormValue("x_franchise"),
		Description:        r.FormValue("x_description"),
		InvoiceID:          r.FormValue("x_id_invoice"),

		CustIDCliente:    r.FormValue("x_cust_id_cliente"),
		CustomerDocType:  r.FormValue("x_customer_doctype"),
		CustomerDocument: r.FormValue("x_customer_document"),
		CustomerName:     r.FormValue("x_customer_name"),
		CustomerLastname: r.FormValue("x_customer_lastname"),
		CustomerEmail:    r.FormValue("x_customer_email"),
		CustomerPhone:    r.FormValue("x_customer_phone"),
		CustomerCountry:  r.FormValue("x_customer_country"),
		CustomerCity:     r.FormValue("x_customer_city"),
		CustomerAddress:  r.FormValue("x_customer_address"),
		CustomerIP:       r.FormValue("x_customer_ip"),

		Signature:   r.FormValue("x_signature"),
		TestRequest: r.FormValue("x_test_request"),
		Business:    r.FormValue("x_business"),
	}
	for i := range p.Extras {
		p.Extras[i] = r.FormValue(fmt.Sprintf("x_extra%d", i+1))
	}
	return p, nil
}

// VerifySignature reports whether payload.Signature matches the signature
// ePayco computes for this transaction:
// sha256(pCustIDCliente^pKey^x_ref_payco^x_transaction_id^x_amount^x_currency_code).
//
// pCustIDCliente and pKey are dashboard credentials (Integraciones -> Llaves
// secretas) distinct from the Client's apiKey/privateKey used to authenticate
// API calls — the docs never confirm they're interchangeable, so they are
// passed explicitly here instead of read from a *Client.
func VerifySignature(pCustIDCliente, pKey string, payload WebhookPayload) bool {
	raw := strings.Join([]string{
		pCustIDCliente, pKey, payload.RefPayco, payload.TransactionID,
		payload.Amount, payload.CurrencyCode,
	}, "^")

	sum := sha256.Sum256([]byte(raw))
	expected := hex.EncodeToString(sum[:])

	return subtle.ConstantTimeCompare([]byte(expected), []byte(payload.Signature)) == 1
}
