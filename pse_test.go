package epayco

import (
	"context"
	"net/http"
	"testing"
)

func TestPSECreate(t *testing.T) {
	c, server := newApifyServer(t, "/payment/process/pse", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true,
			"titleResponse": "Success transaction pse",
			"textResponse": "Success transaction pse",
			"lastAction": "transaction_pse",
			"data": {
				"ref_payco": 26292245, "factura": "QR-APIFY-PSE1596837368", "descripcion": "Pago Factura",
				"valor": 10000, "iva": 0, "baseiva": 0, "moneda": "COP", "estado": "Pendiente",
				"respuesta": "Redireccionando al banco", "autorizacion": "705895997", "recibo": "48771596837369",
				"fecha": "2020-08-07 1656:09",
				"urlbanco": "https://registro.pse.com.co/PSEUserRegister/StartTransaction.aspx",
				"transactionID": "705895997", "ticketId": "48771596837369"
			}
		}`)
	})
	defer server.Close()

	got, err := c.PSE.Create(context.Background(), CreatePSEParams{
		Bank: "1077", Value: "5000", DocType: "CC", DocNumber: "1234567",
		Name: "Juan", LastName: "Mesa", Email: "test@test.co", CellPhone: "3113456768",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.RefPayco != 26292245 || got.TransactionID != "705895997" || got.BankURL == "" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestPSEConfirm(t *testing.T) {
	c, server := newApifyServer(t, "/payment/pse/transaction", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "PENDING", "textResponse": "Transacción Pendiente", "lastAction": "update_transaction",
			"data": {
				"ref_payco": 26291500, "factura": "QR-APIFY-PSE1596836491", "descripcion": "Pago Factura",
				"valor": 10000, "iva": 0, "baseiva": 0, "moneda": "COP", "banco": "BANCO DAVIVIENDA", "estado": "Pendiente",
				"respuesta": "Por favor verificar si el débito fue realizado en el Banco.",
				"autorizacion": "705885651", "recibo": "48771596836492", "fecha": "2020-08-07 16:41:32",
				"franquicia": "PSE", "cod_respuesta": 3, "ip": "190.0.0.0", "enpruebas": 2,
				"tipo_doc": "CC", "documento": "1035863428", "nombres": "Juan Felipe", "apellidos": "Mesa Ocampo",
				"email": "felipemesa14@hotmail.com", "ciudad": "", "direccion": "N/A", "ind_pais": null,
				"transactionID": "705885651", "ticketId": 48771596836492
			}
		}`)
	})
	defer server.Close()

	got, err := c.PSE.Confirm(context.Background(), ConfirmPSEParams{TransactionID: 705885651})
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if got.State != "Pendiente" || got.Bank != "BANCO DAVIVIENDA" || got.TicketID != 48771596836492 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestPSEListBanks(t *testing.T) {
	c, server := newApifyServer(t, "/payment/pse/banks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "SUCCESS", "textResponse": "Bancos Consultados Exitosamente", "lastAction": "Query Bancos",
			"data": [
				{"bankCode": "0", "bankName": "A continuación seleccione su banco"},
				{"bankCode": "1051", "bankName": "BANCO DAVIVIENDA"}
			]
		}`)
	})
	defer server.Close()

	got, err := c.PSE.ListBanks(context.Background())
	if err != nil {
		t.Fatalf("ListBanks: %v", err)
	}
	if len(got) != 2 || got[1].Code != "1051" || got[1].Name != "BANCO DAVIVIENDA" {
		t.Errorf("unexpected result: %+v", got)
	}
}

// TestPSEListBanksMixedCodeTypes reproduces a real quirk found by testing
// live against ePayco's sandbox API: the "banks" list mixes quoted and
// unquoted bankCode values within the same response.
func TestPSEListBanksMixedCodeTypes(t *testing.T) {
	c, server := newApifyServer(t, "/payment/pse/banks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Ok", "textResponse": "Bancos consultados exitosamente", "lastAction": "Query Bancos",
			"data": [
				{"bankCode": "1077", "bankName": "BANKA"},
				{"bankCode": 1022, "bankName": "BANCO UNION COLOMBIANO"}
			]
		}`)
	})
	defer server.Close()

	got, err := c.PSE.ListBanks(context.Background())
	if err != nil {
		t.Fatalf("ListBanks: %v", err)
	}
	if len(got) != 2 || got[0].Code != "1077" || got[1].Code != "1022" {
		t.Errorf("unexpected result: %+v", got)
	}
}
