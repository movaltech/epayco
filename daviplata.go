package epayco

import (
	"context"
	"encoding/json"
	"net/http"
)

// DaviplataService creates and confirms Daviplata mobile-wallet transactions
// through ePayco's apify API.
type DaviplataService struct {
	client *Client
}

// CreateDaviplataParams are the fields accepted to start a Daviplata
// transaction. MethodConfirmation is required — confirmed live against the
// sandbox API ("El campo 'methodConfirmation' es requerido"), even though
// ePayco's own parameter docs list it as optional. ePayco also enforces an
// undocumented minimum Value ("monto minimo no superado" was returned for
// "100" COP; "15000" succeeded).
type CreateDaviplataParams struct {
	DocType            string `json:"docType"`
	Document           string `json:"document"`
	Name               string `json:"name"`
	LastName           string `json:"lastName,omitempty"`
	Email              string `json:"email,omitempty"`
	IndCountry         string `json:"indCountry"`
	Phone              string `json:"phone,omitempty"`
	Country            string `json:"country,omitempty"`
	City               string `json:"city"`
	Address            string `json:"address"`
	IP                 string `json:"ip,omitempty"`
	Currency           string `json:"currency,omitempty"`
	Invoice            string `json:"invoice,omitempty"`
	Description        string `json:"description,omitempty"`
	Value              string `json:"value"`
	Tax                string `json:"tax,omitempty"`
	TaxBase            string `json:"taxBase,omitempty"`
	URLResponse        string `json:"urlResponse,omitempty"`
	URLConfirmation    string `json:"urlConfirmation,omitempty"`
	MethodConfirmation string `json:"methodConfirmation"`
}

// DaviplataTransaction is the transaction created by Create, pending
// confirmation with the OTP sent to the payer's phone. Confirmed live against
// ePayco's sandbox API (2026-07-02): unlike the shape documented in the
// production Postman collection, the real response has no "transaction"
// wrapper (the fields are directly under data). Receipt is decoded leniently
// (see decodeFlexString) for consistency with the same quirk found on Cash
// and PSE, even though both known examples for this endpoint show it as a
// bare number.
type DaviplataTransaction struct {
	RefPayco       int
	Invoice        string
	Description    string
	Value          float64
	Tax            float64
	Ico            float64
	TaxBase        float64
	NetoValue      float64
	Currency       string
	Bank           string
	Status         string
	Response       string
	Authorization  string
	Receipt        string
	Date           string
	Franchise      string
	ResponseCode   int
	ErrorCode      string
	IP             string
	TestMode       int
	DocType        string
	Document       string
	Name           string
	LastName       string
	Email          string
	City           string
	Address        string
	IDSessionToken string
}

func (d *DaviplataTransaction) UnmarshalJSON(data []byte) error {
	var raw struct {
		RefPayco       int             `json:"refPayco"`
		Invoice        string          `json:"invoice"`
		Description    string          `json:"description"`
		Value          float64         `json:"value"`
		Tax            float64         `json:"tax"`
		Ico            float64         `json:"ico"`
		TaxBase        float64         `json:"taxBase"`
		NetoValue      float64         `json:"netoValue"`
		Currency       string          `json:"currency"`
		Bank           string          `json:"bank"`
		Status         string          `json:"estatus"`
		Response       string          `json:"response"`
		Authorization  string          `json:"autorization"`
		Receipt        json.RawMessage `json:"receipt"`
		Date           string          `json:"date"`
		Franchise      string          `json:"franchise"`
		ResponseCode   int             `json:"codResponse"`
		ErrorCode      string          `json:"codError"`
		IP             string          `json:"ip"`
		TestMode       int             `json:"testMode"`
		DocType        string          `json:"docType"`
		Document       string          `json:"document"`
		Name           string          `json:"name"`
		LastName       string          `json:"lastName"`
		Email          string          `json:"email"`
		City           string          `json:"city"`
		Address        string          `json:"address"`
		IDSessionToken string          `json:"idSessionToken"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*d = DaviplataTransaction{
		RefPayco: raw.RefPayco, Invoice: raw.Invoice, Description: raw.Description,
		Value: raw.Value, Tax: raw.Tax, Ico: raw.Ico, TaxBase: raw.TaxBase, NetoValue: raw.NetoValue,
		Currency: raw.Currency, Bank: raw.Bank, Status: raw.Status, Response: raw.Response,
		Authorization: raw.Authorization, Receipt: decodeFlexString(raw.Receipt), Date: raw.Date,
		Franchise: raw.Franchise, ResponseCode: raw.ResponseCode, ErrorCode: raw.ErrorCode,
		IP: raw.IP, TestMode: raw.TestMode, DocType: raw.DocType, Document: raw.Document,
		Name: raw.Name, LastName: raw.LastName, Email: raw.Email, City: raw.City, Address: raw.Address,
		IDSessionToken: raw.IDSessionToken,
	}
	return nil
}

// Create starts a Daviplata transaction via POST /payment/process/daviplata.
// ePayco sends an OTP to the payer's phone; confirm the payment with Confirm.
func (s *DaviplataService) Create(ctx context.Context, params CreateDaviplataParams) (*DaviplataTransaction, error) {
	var result DaviplataTransaction
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/process/daviplata", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfirmDaviplataParams carry the OTP the payer received on their phone.
type ConfirmDaviplataParams struct {
	RefPayco       string `json:"refPayco"`
	IDSessionToken string `json:"idSessionToken"`
	OTP            string `json:"otp"`
}

// DaviplataConfirmation is the outcome of confirming a Daviplata transaction.
type DaviplataConfirmation struct {
	RefPayco                  string `json:"refPayco"`
	Status                    string `json:"status"`
	Date                      string `json:"date"`
	NumApproval               string `json:"numApproval"`
	IDTransactionDaviplata    int64  `json:"idTransactionDaviplata"`
	IDTransactionAutorization string `json:"idTransactionAutorization"`
	Response                  string `json:"response"`
}

// Confirm completes a Daviplata transaction with the OTP sent to the payer's
// phone, via POST /payment/confirm/daviplata.
func (s *DaviplataService) Confirm(ctx context.Context, params ConfirmDaviplataParams) (*DaviplataConfirmation, error) {
	var result DaviplataConfirmation
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/confirm/daviplata", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
