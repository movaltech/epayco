package epayco

import (
	"context"
	"net/http"
	"testing"
)

// TestDaviplataCreate uses a fixture captured live against ePayco's sandbox
// API (2026-07-02) — the shape documented in the production Postman
// collection (a "transaction" wrapper) turned out to be wrong; see the doc
// comment on DaviplataTransaction.
func TestDaviplataCreate(t *testing.T) {
	c, server := newApifyServer(t, "/payment/process/daviplata", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "SUCCESS",
			"textResponse": "Transacción generada exitosamente, pendiente confirmacion del usuario con el OTP generado por DaviPlata",
			"lastAction": "Registrar pago en daviplata",
			"data": {
				"refPayco": 374390450, "invoice": "QR-APIFY-DAVIPLATAcb9865a0af4bb42530cdb0a0f88606c523780174",
				"description": "", "value": 15000, "tax": 0, "ico": 0, "taxBase": 0, "netoValue": 15000,
				"currency": "COP", "bank": "DaviPlata", "estatus": "Pendiente",
				"response": "Esperando confirmacion token DaviPlata", "autorization": "000000",
				"receipt": 48772157430545, "date": "2026-07-02 19:54:54", "franchise": "DP", "codResponse": 3,
				"codError": "P004", "ip": "190.0.0.0", "testMode": 2, "docType": "CC", "document": "1234567890",
				"name": "Juan", "lastName": "Perez", "email": "juan.perez@example.com", "city": "Bogota",
				"address": "Calle 1", "idSessionToken": "166883441"
			}
		}`)
	})
	defer server.Close()

	got, err := c.Daviplata.Create(context.Background(), CreateDaviplataParams{
		DocType: "CC", Document: "1234567890", Name: "Juan", LastName: "Perez", Email: "juan.perez@example.com",
		IndCountry: "CO", City: "Bogota", Address: "Calle 1", Value: "15000", MethodConfirmation: "GET",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.RefPayco != 374390450 || got.Status != "Pendiente" || got.Bank != "DaviPlata" || got.Franchise != "DP" || got.IDSessionToken != "166883441" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestDaviplataConfirm(t *testing.T) {
	c, server := newApifyServer(t, "/payment/confirm/daviplata", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "SUCCESS", "textResponse": "Aprobada", "lastAction": "Confirmar pago en daviplata",
			"data": {
				"refPayco": "45512234", "status": "Aprobado", "date": "2021-08-23T11:54:33",
				"numApproval": "433790", "idTransactionDaviplata": 8393,
				"idTransactionAutorization": "000000008393", "response": "Aprobado"
			}
		}`)
	})
	defer server.Close()

	got, err := c.Daviplata.Confirm(context.Background(), ConfirmDaviplataParams{
		RefPayco: "45512234", IDSessionToken: "sess", OTP: "1234",
	})
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if got.Status != "Aprobado" || got.IDTransactionDaviplata != 8393 {
		t.Errorf("unexpected result: %+v", got)
	}
}
