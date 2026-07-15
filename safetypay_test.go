package epayco

import (
	"context"
	"net/http"
	"testing"
)

// TestSafetypayCreate uses a fixture captured live against ePayco's sandbox
// API (2026-07-02) — the shape documented in the production Postman
// collection (a "transaction" wrapper, ticketId as a bare number) turned out
// to be wrong; see the doc comment on SafetypayTransaction.
func TestSafetypayCreate(t *testing.T) {
	c, server := newApifyServer(t, "/payment/process/safetypay", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Ok", "textResponse": "Token generado exitosamente!.",
			"lastAction": "Envio Transaction Safetypay",
			"data": {
				"refPayco": 374391484, "invoice": "QR-APIFY-SAFETYPAY1a8bc7128bcf24e7f2d36ccdc139dcfe59383058",
				"description": "", "value": 100000, "tax": 0, "ico": 0, "taxBase": 0, "currency": "COP",
				"status": "Pendiente", "response": "Esperando pago del cliente en SafetyPay", "codResponse": "",
				"codError": "", "autorization": "000000", "receipt": "15744621783040549",
				"date": "2026-07-02 20:01:36", "country": "CO", "city": "Bogota",
				"urlBank": "https://sandbox-gateway.safetypay.com/Express4/Checkout/index",
				"transactionId": 374391484, "ticketId": "15744621783040549"
			}
		}`)
	})
	defer server.Close()

	got, err := c.Safetypay.Create(context.Background(), CreateSafetypayParams{
		Cash: "1", ExpirationDate: "2026-07-10", DocType: "CC", Document: "1234567890",
		Name: "Juan", LastName: "Perez", Email: "juan.perez@example.com", IndCountry: "57",
		Phone: "3010000000", Country: "CO", City: "Bogota", Address: "Calle 1", IP: "190.0.0.0",
		Value: "100000", MethodConfirmation: "GET",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.RefPayco != 374391484 || got.BankURL == "" || got.TransactionID != 374391484 || got.TicketID != "15744621783040549" || got.Authorization != "000000" {
		t.Errorf("unexpected result: %+v", got)
	}
}
