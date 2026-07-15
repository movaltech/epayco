package epayco

import (
	"context"
	"net/http"
	"testing"
)

// TestCashCreate uses a fixture captured live against ePayco's sandbox API
// (2026-07-02) — the shape documented in the production Postman collection
// (a "transaction" wrapper, "valueNeto"/"valuePesos", a bare-number receipt)
// turned out to be wrong; see the doc comment on CashTransaction.
func TestCashCreate(t *testing.T) {
	c, server := newApifyServer(t, "/payment/process/cash", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "SUCCESS", "textResponse": "Transacción y pin generados exitosamente",
			"lastAction": "Crear pin efecty",
			"data": {
				"refPayco": 374389735, "invoice": "QR-APIFY-EFECTIVO4ea9d92cf80c783ffcf347325e1208bb80941661",
				"description": "pay test", "value": 20000, "tax": 0, "ico": 0, "taxBase": 0, "total": 20000,
				"currency": "COP", "bank": "EFECTY", "status": "Pendiente",
				"response": "Esperando pago del cliente en punto de servicio Efecty", "autorization": "000000",
				"receipt": "48772157429542", "date": "2026-07-02 19:50:07", "franchise": "EF", "codResponse": 3,
				"codError": "P004", "ip": "190.0.0.0", "testMode": 2, "docType": "CC", "document": "1234567890",
				"name": "Juan", "lastName": "Perez", "email": "juan.perez@example.com", "city": "", "address": "NA",
				"pin": "P374389735", "codeProject": 111992, "paymentDate": "2026-07-02 19:50:07",
				"expirationDate": "2026-07-07 19:50:07", "conversionFactor": 3403.35, "pesos": 20000
			}
		}`)
	})
	defer server.Close()

	got, err := c.Cash.Create(context.Background(), CreateCashParams{
		Description: "pay test", Value: "20000", DocType: "CC", DocNumber: "1234567890",
		Name: "Juan", LastName: "Perez", Email: "juan.perez@example.com", CellPhone: "3010000000",
		IP: "190.0.0.0", PaymentMethod: "EF",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.RefPayco != 374389735 || got.Pin != "P374389735" || got.Receipt != "48772157429542" || got.Total != 20000 || got.Franchise != "EF" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestCashListEntities(t *testing.T) {
	c, server := newApifyServer(t, "/payment/cash/entities", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "SUCCESS", "textResponse": "Bancos Consultados Exitosamente", "lastAction": "Query Bancos",
			"data": [{"Id": "EF", "nombre": "Efecty"}]
		}`)
	})
	defer server.Close()

	got, err := c.Cash.ListEntities(context.Background())
	if err != nil {
		t.Fatalf("ListEntities: %v", err)
	}
	if len(got) != 1 || got[0].ID != "EF" || got[0].Name != "Efecty" {
		t.Errorf("unexpected result: %+v", got)
	}
}
