package epayco

import (
	"context"
	"encoding/json"
	"net/http"
)

// SafetypayService creates Safetypay transactions through ePayco's apify API.
type SafetypayService struct {
	client *Client
}

// CreateSafetypayParams are the fields accepted to start a Safetypay
// transaction. Cash is "1" for online banking or "2" for cash payment.
// City, Address and MethodConfirmation are required — confirmed live against
// the sandbox API, even though ePayco's own parameter docs list City/Address
// as optional. ExpirationDate must be no more than 15 days out ("La fecha de
// expiración no debe superar los 15 dias" for a date further than that).
type CreateSafetypayParams struct {
	Cash               string `json:"cash"`
	ExpirationDate     string `json:"expirationDate"`
	DocType            string `json:"docType"`
	Document           string `json:"document"`
	Name               string `json:"name"`
	LastName           string `json:"lastName"`
	Email              string `json:"email"`
	IndCountry         string `json:"indCountry"`
	Phone              string `json:"phone"`
	Country            string `json:"country"`
	City               string `json:"city"`
	Address            string `json:"address"`
	IP                 string `json:"ip"`
	Currency           string `json:"currency,omitempty"`
	Description        string `json:"description,omitempty"`
	Value              string `json:"value"`
	Tax                string `json:"tax,omitempty"`
	Ico                string `json:"ico,omitempty"`
	TaxBase            string `json:"taxBase,omitempty"`
	URLResponse        string `json:"urlResponse,omitempty"`
	URLConfirmation    string `json:"urlConfirmation,omitempty"`
	MethodConfirmation string `json:"methodConfirmation"`
}

// SafetypayTransaction is the transaction created by Create, including the
// bank redirect URL the payer must be sent to. Confirmed live against
// ePayco's sandbox API (2026-07-02): unlike the shape documented in the
// production Postman collection, the real response has no "transaction"
// wrapper (the fields are directly under data). Receipt and TicketID are
// decoded leniently (see decodeFlexString): both came back as quoted strings
// live, but the Postman-documented example shows TicketID as a bare number.
type SafetypayTransaction struct {
	RefPayco      int
	Invoice       string
	Description   string
	Value         float64
	Tax           float64
	Ico           float64
	TaxBase       float64
	Currency      string
	Status        string
	Response      string
	ResponseCode  string
	ErrorCode     string
	Authorization string
	Receipt       string
	Date          string
	Country       string
	City          string
	BankURL       string
	TransactionID int64
	TicketID      string
}

func (s *SafetypayTransaction) UnmarshalJSON(data []byte) error {
	var raw struct {
		RefPayco     int     `json:"refPayco"`
		Invoice      string  `json:"invoice"`
		Description  string  `json:"description"`
		Value        float64 `json:"value"`
		Tax          float64 `json:"tax"`
		Ico          float64 `json:"ico"`
		TaxBase      float64 `json:"taxBase"`
		Currency     string  `json:"currency"`
		Status       string  `json:"status"`
		Response     string  `json:"response"`
		ResponseCode string  `json:"codResponse"`
		ErrorCode    string  `json:"codError"`
		// Note the spelling: "autorization", not "autorizacion" as PSE/cash/
		// Daviplata use — confirmed live against the sandbox API.
		Authorization string          `json:"autorization"`
		Receipt       json.RawMessage `json:"receipt"`
		Date          string          `json:"date"`
		Country       string          `json:"country"`
		City          string          `json:"city"`
		BankURL       string          `json:"urlBank"`
		TransactionID int64           `json:"transactionId"`
		TicketID      json.RawMessage `json:"ticketId"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*s = SafetypayTransaction{
		RefPayco: raw.RefPayco, Invoice: raw.Invoice, Description: raw.Description,
		Value: raw.Value, Tax: raw.Tax, Ico: raw.Ico, TaxBase: raw.TaxBase, Currency: raw.Currency,
		Status: raw.Status, Response: raw.Response, ResponseCode: raw.ResponseCode, ErrorCode: raw.ErrorCode,
		Authorization: raw.Authorization, Receipt: decodeFlexString(raw.Receipt), Date: raw.Date,
		Country: raw.Country, City: raw.City, BankURL: raw.BankURL, TransactionID: raw.TransactionID,
		TicketID: decodeFlexString(raw.TicketID),
	}
	return nil
}

// Create starts a Safetypay transaction via POST /payment/process/safetypay.
// The payer must be redirected to SafetypayTransaction.BankURL to complete
// the payment.
func (s *SafetypayService) Create(ctx context.Context, params CreateSafetypayParams) (*SafetypayTransaction, error) {
	var result SafetypayTransaction
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/process/safetypay", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
