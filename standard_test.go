package epayco

import (
	"context"
	"net/http"
	"testing"
)

func TestStandardCreate(t *testing.T) {
	c, server := newApifyServer(t, "/payment/process/standard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Success transaction", "textResponse": "Success transaction", "lastAction": "create_transaction",
			"data": {
				"refEpayco": 26922530, "invoice": "QR-APIFY-RDP1600227267", "description": "", "value": "50000",
				"tax": "0", "taxBase": "0", "currency": "COP", "bank": "RDP", "state": "Pendiente",
				"stateMessage": "Esperando confirmación del cliente", "authorizationCode": "000000", "receipt": "000000",
				"dateTime": "2020-09-15 22:34:30", "channel": "RDP", "responseCode": 3, "ip": "190.000.000.000",
				"docType": "CC", "docNumber": "1234567", "name": "Juan", "lastName": "Mesa", "email": "juanmesa@hotmail.com",
				"city": "", "address": "N/A", "country": null,
				"urlRedirect": "https://recarga-daviplata.epayco.io/pagar?terminal=26922530"
			}
		}`)
	})
	defer server.Close()

	got, err := c.Standard.Create(context.Background(), CreateStandardParams{
		Channel: "RDP", Value: "50000", DocType: "CC", DocNumber: "1234567",
		Name: "Juan", LastName: "Mesa", Email: "juanmesa@hotmail.com", CellPhone: "584127751699",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.RefEpayco != 26922530 || got.URLRedirect == "" || got.State != "Pendiente" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestStandardConfirm(t *testing.T) {
	c, server := newApifyServer(t, "/payment/confirm/standard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Success transaction", "textResponse": "Success transaction", "lastAction": "confirm_transaction",
			"data": {
				"refEpayco": 26922533, "invoice": "QR-APIFY-RDP1600228001", "description": "", "value": 50000,
				"tax": 0, "taxBase": 0, "currency": "COP", "bank": "cualquier banco", "state": "Aceptada",
				"stateMessage": "Confirmada por el cliente", "authorizationCode": "123456677889", "receipt": "000000",
				"dateTime": "2020-09-15 22:46:45", "channel": "RDP", "responseCode": 1, "ip": "190.000.000.000",
				"docType": "CC", "docNumber": "1234567", "name": "Juan", "lastName": "Mesa",
				"email": "juanmesa@hotmail.com", "city": "", "address": "N/A", "country": null
			}
		}`)
	})
	defer server.Close()

	got, err := c.Standard.Confirm(context.Background(), ConfirmStandardParams{
		Channel: "RDP", Value: "50000", RefEpayco: "26922533",
		CurrentStatusTransaction: "Pendiente", StatusTransaction: "Aceptada", AuthorizationCode: "123456677889",
	})
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if got.State != "Aceptada" || got.Value != 50000 {
		t.Errorf("unexpected result: %+v", got)
	}
}
