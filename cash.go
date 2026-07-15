package epayco

import (
	"context"
	"encoding/json"
	"net/http"
)

// CashService creates cash-network transactions (Efecty and similar) through
// ePayco's apify API.
type CashService struct {
	client *Client
}

// CreateCashParams are the fields accepted to start a cash transaction.
// PaymentMethod identifies the cash network (e.g. "EF" for Efecty) and must
// be one of the IDs returned by ListEntities.
type CreateCashParams struct {
	Invoice            string `json:"invoice,omitempty"`
	Description        string `json:"description,omitempty"`
	Value              string `json:"value"`
	Tax                string `json:"tax,omitempty"`
	TaxBase            string `json:"taxBase,omitempty"`
	Currency           string `json:"currency,omitempty"`
	TypePerson         string `json:"typePerson,omitempty"`
	DocType            string `json:"docType"`
	DocNumber          string `json:"docNumber"`
	Name               string `json:"name"`
	LastName           string `json:"lastName"`
	Email              string `json:"email"`
	CellPhone          string `json:"cellPhone"`
	EndDate            string `json:"endDate,omitempty"`
	IP                 string `json:"ip"`
	URLResponse        string `json:"urlResponse,omitempty"`
	URLConfirmation    string `json:"urlConfirmation,omitempty"`
	MethodConfirmation string `json:"methodConfirmation,omitempty"`
	PaymentMethod      string `json:"paymentMethod"`
}

// CashTransaction is the transaction created by Create, including the pin/
// reference the payer must present at the cash network's point of service.
// Confirmed live against ePayco's sandbox API (2026-07-02): unlike the shape
// documented in the production Postman collection, the real response has no
// "transaction" wrapper (the fields are directly under data), and uses
// "total"/"pesos" instead of "valorneto"/"valuePesos". Receipt is decoded
// leniently because ePayco encodes it inconsistently — a quoted string live,
// a bare number in the Postman-documented example (the same quirk as
// Bank.Code and TransactionDetail) — see decodeFlexString.
type CashTransaction struct {
	RefPayco         int
	Invoice          string
	Description      string
	Value            float64
	Tax              float64
	Ico              float64
	TaxBase          float64
	Total            float64
	Currency         string
	Bank             string
	Status           string
	Response         string
	Authorization    string
	Receipt          string
	Date             string
	Franchise        string
	ResponseCode     int
	ErrorCode        string
	IP               string
	TestMode         int
	DocType          string
	Document         string
	Name             string
	LastName         string
	Email            string
	City             string
	Address          string
	Pin              string
	CodeProject      int
	PaymentDate      string
	ExpirationDate   string
	ConversionFactor float64
	Pesos            float64
}

func (c *CashTransaction) UnmarshalJSON(data []byte) error {
	var raw struct {
		RefPayco         int             `json:"refPayco"`
		Invoice          string          `json:"invoice"`
		Description      string          `json:"description"`
		Value            float64         `json:"value"`
		Tax              float64         `json:"tax"`
		Ico              float64         `json:"ico"`
		TaxBase          float64         `json:"taxBase"`
		Total            float64         `json:"total"`
		Currency         string          `json:"currency"`
		Bank             string          `json:"bank"`
		Status           string          `json:"status"`
		Response         string          `json:"response"`
		Authorization    string          `json:"autorization"`
		Receipt          json.RawMessage `json:"receipt"`
		Date             string          `json:"date"`
		Franchise        string          `json:"franchise"`
		ResponseCode     int             `json:"codResponse"`
		ErrorCode        string          `json:"codError"`
		IP               string          `json:"ip"`
		TestMode         int             `json:"testMode"`
		DocType          string          `json:"docType"`
		Document         string          `json:"document"`
		Name             string          `json:"name"`
		LastName         string          `json:"lastName"`
		Email            string          `json:"email"`
		City             string          `json:"city"`
		Address          string          `json:"address"`
		Pin              string          `json:"pin"`
		CodeProject      int             `json:"codeProject"`
		PaymentDate      string          `json:"paymentDate"`
		ExpirationDate   string          `json:"expirationDate"`
		ConversionFactor float64         `json:"conversionFactor"`
		Pesos            float64         `json:"pesos"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*c = CashTransaction{
		RefPayco: raw.RefPayco, Invoice: raw.Invoice, Description: raw.Description,
		Value: raw.Value, Tax: raw.Tax, Ico: raw.Ico, TaxBase: raw.TaxBase, Total: raw.Total,
		Currency: raw.Currency, Bank: raw.Bank, Status: raw.Status, Response: raw.Response,
		Authorization: raw.Authorization, Receipt: decodeFlexString(raw.Receipt), Date: raw.Date,
		Franchise: raw.Franchise, ResponseCode: raw.ResponseCode, ErrorCode: raw.ErrorCode,
		IP: raw.IP, TestMode: raw.TestMode, DocType: raw.DocType, Document: raw.Document,
		Name: raw.Name, LastName: raw.LastName, Email: raw.Email, City: raw.City, Address: raw.Address,
		Pin: raw.Pin, CodeProject: raw.CodeProject, PaymentDate: raw.PaymentDate,
		ExpirationDate: raw.ExpirationDate, ConversionFactor: raw.ConversionFactor, Pesos: raw.Pesos,
	}
	return nil
}

// Create starts a cash transaction via POST /payment/process/cash.
func (s *CashService) Create(ctx context.Context, params CreateCashParams) (*CashTransaction, error) {
	var result CashTransaction
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/process/cash", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CashEntity is a cash network available to receive payments (e.g. Efecty).
type CashEntity struct {
	ID   string `json:"Id"`
	Name string `json:"nombre"`
}

// ListEntities retrieves the cash networks available for CreateCashParams.PaymentMethod,
// via GET /payment/cash/entities.
func (s *CashService) ListEntities(ctx context.Context) ([]CashEntity, error) {
	var result []CashEntity
	if err := s.client.doApify(ctx, http.MethodGet, "/payment/cash/entities", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
