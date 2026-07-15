package epayco

import (
	"context"
	"net/http"
	"testing"
)

func TestCustomerCreate(t *testing.T) {
	c, server := newApifyServer(t, "/token/customer", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true,
			"titleResponse": "Success token generate",
			"textResponse": "Success token generate",
			"lastAction": "token_customer",
			"data": {
				"status": true, "success": true, "type": "Create customer",
				"data": {
					"status": "exitoso",
					"description": "El cliente fue creado exitosamente con el id: DiYPdhgncppes9KAb",
					"customerId": "DiYPdhgncppes9KAb",
					"name": "jon",
					"email": "jondoe@hotmail.com"
				},
				"object": "customer"
			}
		}`)
	})
	defer server.Close()

	got, err := c.Customers.Create(context.Background(), CreateParams{
		DocType: "CC", DocNumber: "123456789", Name: "jon", LastName: "doe",
		Email: "jondoe@hotmail.com", CellPhone: "0000000000", Phone: "0000000",
		RequireCardToken: true, CardTokenID: "rdSHdTmJFioQdfp9k",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.Data.CustomerID != "DiYPdhgncppes9KAb" {
		t.Errorf("Data.CustomerID = %q", got.Data.CustomerID)
	}
}

func TestCustomerCreateAlreadyAssociatedError(t *testing.T) {
	c, server := newApifyServer(t, "/token/customer", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": false,
			"titleResponse": "Error customer",
			"textResponse": "Error customer Cliente ya asociado o token inexistente",
			"lastAction": "create_customer",
			"data": {"error": {"status": "error", "description": "El token no se puede asociar al cliente, verifique que: el token existe, el cliente no esté asociado y que el token no este asociado a otro cliente ."}}
		}`)
	})
	defer server.Close()

	_, err := c.Customers.Create(context.Background(), CreateParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCustomerUpdate(t *testing.T) {
	c, server := newApifyServer(t, "/subscriptions/customer/update", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true,
			"titleResponse": "Customer successfully updated",
			"textResponse": "Customer successfully updated",
			"lastAction": "update_customer",
			"data": {"status": true, "success": true, "type": "Edit customer",
				"data": {"status": "exitoso", "description": "actualizado", "customerId": "x", "name": "Juan Felipe", "email": "a@b.com"},
				"object": "customer"}
		}`)
	})
	defer server.Close()

	if err := c.Customers.Update(context.Background(), UpdateParams{CustomerID: "cDrjtT6J3NXgkorWB", Name: "Juan Felipe"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestCustomerGet(t *testing.T) {
	c, server := newApifyServer(t, "/subscriptions/customer", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true,
			"titleResponse": "Customer successfully recovered",
			"textResponse": "Customer successfully recovered",
			"lastAction": "get_customer",
			"data": {
				"status": true, "success": true, "type": "Find customer",
				"data": {
					"id_customer": "cDrjtT6J3NXgkorWB",
					"name": "Juan Felipe",
					"email": "felipemesa14@hotmail.com",
					"doc_type": "CC",
					"doc_number": "1035863428",
					"created": "03/31/2020",
					"cards": [
						{"token": "**********FqctY4", "franchise": "visa", "mask": "************0326", "created": "11/29/2019", "default": false}
					]
				},
				"object": "customer"
			}
		}`)
	})
	defer server.Close()

	got, err := c.Customers.Get(context.Background(), "cDrjtT6J3NXgkorWB")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "cDrjtT6J3NXgkorWB" || len(got.Cards) != 1 || got.Cards[0].Franchise != "visa" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestCustomerList(t *testing.T) {
	c, server := newApifyServer(t, "/subscriptions/customers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true,
			"titleResponse": "Customers successfully recovered",
			"textResponse": "Customers successfully recovered",
			"lastAction": "get_customers",
			"data": {
				"status": true, "success": true, "type": "Find all customer",
				"data": [
					{"id_customer": "cEqSJMvbx2ENFXe3h", "object": "customer", "name": "Juan Felipe", "email": "a@b.com", "phone": "3024133765", "created": "03/13/2019"}
				],
				"object": "customers"
			}
		}`)
	})
	defer server.Close()

	got, err := c.Customers.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].ID != "cEqSJMvbx2ENFXe3h" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestCustomerDeleteToken(t *testing.T) {
	c, server := newApifyServer(t, "/subscription/token/card/delete", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Success token delete", "textResponse": "Success token delete",
			"lastAction": "delete_token_card",
			"data": {"status": true, "message": "Token removido", "success": true, "type": "Token", "data": {}, "object": "token"}
		}`)
	})
	defer server.Close()

	err := c.Customers.DeleteToken(context.Background(), DeleteTokenParams{Franchise: "visa", Mask: "1234", CustomerID: "x"})
	if err != nil {
		t.Fatalf("DeleteToken: %v", err)
	}
}

func TestCustomerAddToken(t *testing.T) {
	c, server := newApifyServer(t, "/subscriptions/customer/add/new/token", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Success add token customer", "textResponse": "Success add token customer",
			"lastAction": "add_token_customer",
			"data": {"status": true, "type": "add card customer", "message": "El token fue atachado con éxito al customer", "object": "customer"}
		}`)
	})
	defer server.Close()

	err := c.Customers.AddToken(context.Background(), AddTokenParams{CardToken: "doDhtM4jrahhf3Hyv", CustomerID: "x"})
	if err != nil {
		t.Fatalf("AddToken: %v", err)
	}
}

func TestCustomerSetDefaultToken(t *testing.T) {
	c, server := newApifyServer(t, "/subscriptions/customer/add/new/token/default", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true, "titleResponse": "Success add new token default", "textResponse": "Success add new token default",
			"lastAction": "add_new_token_default",
			"data": {"status": true, "message": "Reassignment card by default successful"}
		}`)
	})
	defer server.Close()

	err := c.Customers.SetDefaultToken(context.Background(), SetDefaultTokenParams{CardToken: "x", CustomerID: "y", Franchise: "visa", Mask: "1234"})
	if err != nil {
		t.Fatalf("SetDefaultToken: %v", err)
	}
}
