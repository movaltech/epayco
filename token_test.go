package epayco

import (
	"context"
	"net/http"
	"testing"
)

func TestTokenCreate(t *testing.T) {
	c, server := newApifyServer(t, "/token/card", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": true,
			"titleResponse": "Success token generate",
			"textResponse": "Success token generate",
			"lastAction": "token_card",
			"data": {
				"status": true,
				"id": "6MPbuNLPbMiBR2PR9",
				"success": true,
				"type": "card",
				"data": {"status": "exitoso", "id": "6MPbuNLPbMiBR2PR9", "created": "12/18/2019", "livemode": false},
				"card": {"exp_month": "12", "exp_year": "2019", "name": "visa"},
				"object": "token"
			}
		}`)
	})
	defer server.Close()

	got, err := c.Tokens.Create(context.Background(), CreateCardParams{
		CardNumber:   "4504070849985391",
		CardExpYear:  "2019",
		CardExpMonth: "12",
		CardCVC:      "677",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID != "6MPbuNLPbMiBR2PR9" {
		t.Errorf("ID = %q", got.ID)
	}
	if got.Card.Name != "visa" {
		t.Errorf("Card.Name = %q, want visa", got.Card.Name)
	}
	if got.Data.Status != "exitoso" {
		t.Errorf("Data.Status = %q, want exitoso", got.Data.Status)
	}
}

func TestTokenCreateValidationError(t *testing.T) {
	c, server := newApifyServer(t, "/token/card", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, `{
			"success": false,
			"titleResponse": "Error",
			"textResponse": "Some fields are required, please correct the errors and try again",
			"lastAction": "validation data",
			"data": {"totalErrors": 1, "errors": [{"codError": 500, "errorMessage": "field cardNumber required"}]}
		}`)
	})
	defer server.Close()

	_, err := c.Tokens.Create(context.Background(), CreateCardParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.(*Error).Message != "field cardNumber required" {
		t.Errorf("Message = %q", err.(*Error).Message)
	}
}
