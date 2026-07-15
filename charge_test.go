package epayco

import (
	"context"
	"net/http"
	"testing"
)

// TestChargeGetQuotedAmounts uses the shape documented in the production
// Postman collection (amount-like fields as quoted decimal strings).
func TestChargeGetQuotedAmounts(t *testing.T) {
	c, server := newApifyServer(t, "/payment/transaction", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Correcto", "textResponse": "Transacción consultada existosamente",
			"lastAction": "Consultar Transaccion",
			"data": {"transaction": {
				"clientId": "101291", "refPayco": 30615768, "invoice": "1472050778", "description": "pay test",
				"amount": "20000.00", "amountCountry": "20000.00", "amountOk": "20000.00", "tax": "0.00", "ico": 0,
				"baseTax": "0.00", "currency": "COP", "bank": "EFECTY", "response": "Pendiente",
				"autorizacion": "000000", "transactionId": "48771657326867", "date": "2021-07-19 11:11:39",
				"codeResponse": 3, "responseReasonText": "P004-Esperando pago del cliente en punto de servicio Efecty",
				"codTransactionState": "3", "status": "Pendiente", "errorCode": "P004", "franchise": "EF",
				"nameBusiness": "Gerson david Vasquez Rodriguez", "testMode": "FALSE"
			}}
		}`)
	})
	defer server.Close()

	got, err := c.Charges.Get(context.Background(), "30615768")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.RefPayco != 30615768 || got.Status != "Pendiente" || got.Franchise != "EF" || got.ClientID != "101291" || got.Amount != "20000.00" {
		t.Errorf("unexpected result: %+v", got)
	}
}

// TestChargeGetBareNumberAmounts uses a fixture captured live against
// ePayco's sandbox API (2026-07-02), where the same amount-like fields come
// back as bare JSON numbers instead — the exact quirk decodeFlexString exists
// to tolerate.
func TestChargeGetBareNumberAmounts(t *testing.T) {
	c, server := newApifyServer(t, "/payment/transaction", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Correcto", "textResponse": "Transacción consultada existosamente",
			"lastAction": "Consultar Transaccion",
			"data": {"transaction": {
				"clientId": 999999, "refPayco": 374217390, "invoice": "QR-APIFY1782967429",
				"description": "Compra referencia QR-APIFY1782967429", "amount": 50000, "amountCountry": 50000,
				"amountOk": 50000, "tax": 0, "ico": 0, "baseTax": 0, "currency": "COP", "bank": "BANCO DE PRUEBAS",
				"cardNumber": "457562*******0326", "quotas": "1", "response": "Aceptada", "autorizacion": "000000",
				"transactionId": "374217390", "date": "2026-07-01 23:43:50", "codeResponse": 1,
				"responseReasonText": "Aprobada", "codTransactionState": 1, "status": "Aceptada", "errorCode": "00",
				"franchise": "VS", "nameBusiness": "Test Merchant", "testMode": "TRUE"
			}}
		}`)
	})
	defer server.Close()

	got, err := c.Charges.Get(context.Background(), "374217390")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.RefPayco != 374217390 || got.Status != "Aceptada" || got.ClientID != "999999" || got.Amount != "50000" || got.CodTransactionState != "1" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestChargeGetNotFound(t *testing.T) {
	c, server := newApifyServer(t, "/payment/transaction", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": false, "titleResponse": "Error", "textResponse": "Transacción no existe",
			"data": {"error": {}}
		}`)
	})
	defer server.Close()

	_, err := c.Charges.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.(*Error).Message != "Transacción no existe" {
		t.Errorf("Message = %q", err.(*Error).Message)
	}
}

// TestChargeCreate uses a fixture captured live against ePayco's sandbox API
// (2026-07-02, real response with values redacted/simplified) — see the
// architecture doc section 8 note on how this corrected the original
// best-effort field guesses.
func TestChargeCreate(t *testing.T) {
	c, server := newApifyServer(t, "/payment/process", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Success transaction", "textResponse": "Success transaction",
			"lastAction": "transaction_split_payment_tc",
			"data": {
				"transaction": {
					"status": true, "success": true, "type": "Create payment",
					"data": {
						"ref_payco": 374217241, "factura": "QR-APIFY1782967217",
						"descripcion": "Compra referencia QR-APIFY1782967217", "valor": 50000, "iva": 0, "ico": 0,
						"baseiva": 0, "valorneto": 50000, "moneda": "COP", "banco": "BANCO DE PRUEBAS",
						"estado": "Aceptada", "respuesta": "Aprobada", "autorizacion": "000000",
						"recibo": "374217241", "fecha": "2026-07-01 23:40:18", "franquicia": "VS",
						"cod_respuesta": 1, "cod_error": "00", "ip": "96.58.181.209", "enpruebas": 1,
						"tipo_doc": "CC", "documento": "1234567890", "nombres": "Juan", "apellidos": "Perez",
						"email": "juan.perez@example.com", "ciudad": "Sin Ciudad", "direccion": "Calle 1 # 23",
						"ind_pais": "PE", "country_card": "PE",
						"extras": {"extra1": "", "extra2": ""},
						"cc_network_response": {"code": "00", "message": "Aprobada"},
						"paymentProviderData": {"accumulatedPoints": 0}
					},
					"object": "payment"
				},
				"tokenCard": {"email": "juan.perez@example.com", "cardTokenId": "a45ebb2c41b7923780ae89c", "customerId": "N/A"}
			}
		}`)
	})
	defer server.Close()

	got, err := c.Charges.Create(context.Background(), CreateChargeParams{
		Value: "50000", DocType: "CC", DocNumber: "123", Name: "Juan", LastName: "Mesa",
		Email: "test@test.com", CellPhone: "3000000000", Phone: "2000000",
		CardNumber: "4504070849985391", CardExpYear: "2025", CardExpMonth: "12", CardCVC: "123", Dues: "1",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.RefPayco != 374217241 || got.CardTokenID != "a45ebb2c41b7923780ae89c" || got.CustomerID != "N/A" {
		t.Errorf("unexpected result: %+v", got)
	}
	if got.State != "Aceptada" || got.Franchise != "VS" || got.NetworkResponse.Code != "00" {
		t.Errorf("unexpected result: %+v", got)
	}
}
