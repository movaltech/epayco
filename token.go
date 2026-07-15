package epayco

import (
	"context"
	"net/http"
)

// TokenService tokenizes credit cards through ePayco's apify API
// (apify.epayco.co), so future charges can reference a token instead of raw
// card data.
type TokenService struct {
	client *Client
}

// CreateCardParams are the fields required to tokenize a credit card.
type CreateCardParams struct {
	CardNumber   string `json:"cardNumber"`
	CardExpYear  string `json:"cardExpYear"`
	CardExpMonth string `json:"cardExpMonth"`
	CardCVC      string `json:"cardCvc"`
}

// CardToken is the token generated for a credit card, returned by Create.
type CardToken struct {
	Status  bool   `json:"status"`
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Type    string `json:"type"`
	Data    struct {
		Status   string `json:"status"`
		ID       string `json:"id"`
		Created  string `json:"created"`
		Livemode bool   `json:"livemode"`
	} `json:"data"`
	Card struct {
		ExpMonth string `json:"exp_month"`
		ExpYear  string `json:"exp_year"`
		Name     string `json:"name"`
	} `json:"card"`
	Object string `json:"object"`
}

// Create tokenizes a credit card via POST /token/card.
func (s *TokenService) Create(ctx context.Context, params CreateCardParams) (*CardToken, error) {
	var result CardToken
	if err := s.client.doApify(ctx, http.MethodPost, "/token/card", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
