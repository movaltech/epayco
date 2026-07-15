package epayco

import (
	"context"
	"encoding/json"
	"net/http"
)

// ChargeService creates credit-card transactions and looks up transactions of
// any payment method by reference, through ePayco's apify API.
type ChargeService struct {
	client *Client
}

// CreateChargeParams are the fields accepted to charge a credit card. Send
// CardNumber/CardExpYear/CardExpMonth/CardCVC/Dues the first time a card is
// used; ePayco returns CardTokenID/CustomerID in the response so subsequent
// charges for the same card+customer can omit the raw card data and send
// CardTokenID/CustomerID instead.
type CreateChargeParams struct {
	Value     string `json:"value"`
	DocType   string `json:"docType"`
	DocNumber string `json:"docNumber"`
	Name      string `json:"name"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	CellPhone string `json:"cellPhone"`
	Phone     string `json:"phone"`
	Address   string `json:"address,omitempty"`

	// First-time card data (required unless CardTokenID/CustomerID are set).
	CardNumber   string `json:"cardNumber,omitempty"`
	CardExpYear  string `json:"cardExpYear,omitempty"`
	CardExpMonth string `json:"cardExpMonth,omitempty"`
	CardCVC      string `json:"cardCvc,omitempty"`
	Dues         string `json:"dues,omitempty"`

	// Returning-customer references, in place of raw card data.
	CardTokenID string `json:"cardTokenId,omitempty"`
	CustomerID  string `json:"customerId,omitempty"`

	Tax                string `json:"tax,omitempty"`
	TaxBase            string `json:"taxBase,omitempty"`
	Description        string `json:"description,omitempty"`
	Invoice            string `json:"invoice,omitempty"`
	Currency           string `json:"currency,omitempty"`
	TypePerson         string `json:"typePerson,omitempty"`
	Country            string `json:"country,omitempty"`
	City               string `json:"city,omitempty"`
	IP                 string `json:"ip,omitempty"`
	URLResponse        string `json:"urlResponse,omitempty"`
	URLConfirmation    string `json:"urlConfirmation,omitempty"`
	MethodConfirmation string `json:"methodConfirmation,omitempty"`
}

// NetworkResponse is the card network's own response code for a charge,
// nested in ChargeTransaction.
type NetworkResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ChargeTransaction is the transaction created by Create. Confirmed live
// against ePayco's sandbox API (2026-07-02): unlike the other apify payment
// methods (PSE, cash, Daviplata, Safetypay, standard checkout), a card charge
// nests the transaction under transaction.data (not transaction directly) and
// uses the same snake_case field names as PSE rather than camelCase.
type ChargeTransaction struct {
	RefPayco        int             `json:"ref_payco"`
	Invoice         string          `json:"factura"`
	Description     string          `json:"descripcion"`
	Value           float64         `json:"valor"`
	Tax             float64         `json:"iva"`
	Ico             float64         `json:"ico"`
	TaxBase         float64         `json:"baseiva"`
	NetValue        float64         `json:"valorneto"`
	Currency        string          `json:"moneda"`
	Bank            string          `json:"banco"`
	State           string          `json:"estado"`
	Response        string          `json:"respuesta"`
	Authorization   string          `json:"autorizacion"`
	Receipt         string          `json:"recibo"`
	Date            string          `json:"fecha"`
	Franchise       string          `json:"franquicia"`
	ResponseCode    int             `json:"cod_respuesta"`
	ErrorCode       string          `json:"cod_error"`
	IP              string          `json:"ip"`
	DocType         string          `json:"tipo_doc"`
	Document        string          `json:"documento"`
	Name            string          `json:"nombres"`
	LastName        string          `json:"apellidos"`
	Email           string          `json:"email"`
	City            string          `json:"ciudad"`
	Address         string          `json:"direccion"`
	NetworkResponse NetworkResponse `json:"cc_network_response"`

	// CardTokenID and CustomerID come from a sibling "tokenCard" object in
	// the raw response (not from the transaction itself) — store the data
	// you'll need for subsequent charges (see CreateChargeParams).
	CardTokenID string `json:"-"`
	CustomerID  string `json:"-"`
}

// Create charges a credit card via POST /payment/process.
func (s *ChargeService) Create(ctx context.Context, params CreateChargeParams) (*ChargeTransaction, error) {
	var wrapped struct {
		Transaction struct {
			Data ChargeTransaction `json:"data"`
		} `json:"transaction"`
		TokenCard struct {
			CardTokenID string `json:"cardTokenId"`
			CustomerID  string `json:"customerId"`
		} `json:"tokenCard"`
	}
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/process", params, &wrapped); err != nil {
		return nil, err
	}

	result := wrapped.Transaction.Data
	result.CardTokenID = wrapped.TokenCard.CardTokenID
	result.CustomerID = wrapped.TokenCard.CustomerID
	return &result, nil
}

// TransactionDetail is the status of a transaction of any payment method,
// returned by Get. Confirmed live against ePayco's sandbox API (2026-07-02):
// the real response wraps the record under transaction (not directly under
// data, as an earlier Postman-documented example showed for a cash lookup).
// ClientID, Amount, AmountCountry, AmountOk, Tax, BaseTax and
// CodTransactionState are decoded leniently because ePayco encodes them
// inconsistently across responses — quoted decimals in some, bare numbers in
// others (the same quirk found in Bank.Code) — see decodeFlexString.
type TransactionDetail struct {
	ClientID            string
	RefPayco            int
	Invoice             string
	Description         string
	Amount              string
	AmountCountry       string
	AmountOk            string
	Tax                 string
	Ico                 int
	BaseTax             string
	Currency            string
	Bank                string
	CardNumber          string
	Quotas              string
	Response            string
	Authorization       string
	TransactionID       string
	Date                string
	CodeResponse        int
	ResponseReasonText  string
	CodTransactionState string
	Status              string
	ErrorCode           string
	Franchise           string
	NameBusiness        string
	DocType             string
	Document            string
	Name                string
	LastName            string
	Email               string
	Phone               string
	IndCountry          string
	Country             string
	City                string
	Address             string
	IP                  string
	Signature           string
	TestMode            string
}

func (t *TransactionDetail) UnmarshalJSON(data []byte) error {
	var raw struct {
		ClientID            json.RawMessage `json:"clientId"`
		RefPayco            int             `json:"refPayco"`
		Invoice             string          `json:"invoice"`
		Description         string          `json:"description"`
		Amount              json.RawMessage `json:"amount"`
		AmountCountry       json.RawMessage `json:"amountCountry"`
		AmountOk            json.RawMessage `json:"amountOk"`
		Tax                 json.RawMessage `json:"tax"`
		Ico                 int             `json:"ico"`
		BaseTax             json.RawMessage `json:"baseTax"`
		Currency            string          `json:"currency"`
		Bank                string          `json:"bank"`
		CardNumber          string          `json:"cardNumber"`
		Quotas              string          `json:"quotas"`
		Response            string          `json:"response"`
		Authorization       string          `json:"autorizacion"`
		TransactionID       string          `json:"transactionId"`
		Date                string          `json:"date"`
		CodeResponse        int             `json:"codeResponse"`
		ResponseReasonText  string          `json:"responseReasonText"`
		CodTransactionState json.RawMessage `json:"codTransactionState"`
		Status              string          `json:"status"`
		ErrorCode           string          `json:"errorCode"`
		Franchise           string          `json:"franchise"`
		NameBusiness        string          `json:"nameBusiness"`
		DocType             string          `json:"docType"`
		Document            string          `json:"document"`
		Name                string          `json:"name"`
		LastName            string          `json:"lastName"`
		Email               string          `json:"email"`
		Phone               string          `json:"phone"`
		IndCountry          string          `json:"indCountry"`
		Country             string          `json:"country"`
		City                string          `json:"city"`
		Address             string          `json:"address"`
		IP                  string          `json:"ip"`
		Signature           string          `json:"signature"`
		TestMode            string          `json:"testMode"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*t = TransactionDetail{
		ClientID:            decodeFlexString(raw.ClientID),
		RefPayco:            raw.RefPayco,
		Invoice:             raw.Invoice,
		Description:         raw.Description,
		Amount:              decodeFlexString(raw.Amount),
		AmountCountry:       decodeFlexString(raw.AmountCountry),
		AmountOk:            decodeFlexString(raw.AmountOk),
		Tax:                 decodeFlexString(raw.Tax),
		Ico:                 raw.Ico,
		BaseTax:             decodeFlexString(raw.BaseTax),
		Currency:            raw.Currency,
		Bank:                raw.Bank,
		CardNumber:          raw.CardNumber,
		Quotas:              raw.Quotas,
		Response:            raw.Response,
		Authorization:       raw.Authorization,
		TransactionID:       raw.TransactionID,
		Date:                raw.Date,
		CodeResponse:        raw.CodeResponse,
		ResponseReasonText:  raw.ResponseReasonText,
		CodTransactionState: decodeFlexString(raw.CodTransactionState),
		Status:              raw.Status,
		ErrorCode:           raw.ErrorCode,
		Franchise:           raw.Franchise,
		NameBusiness:        raw.NameBusiness,
		DocType:             raw.DocType,
		Document:            raw.Document,
		Name:                raw.Name,
		LastName:            raw.LastName,
		Email:               raw.Email,
		Phone:               raw.Phone,
		IndCountry:          raw.IndCountry,
		Country:             raw.Country,
		City:                raw.City,
		Address:             raw.Address,
		IP:                  raw.IP,
		Signature:           raw.Signature,
		TestMode:            raw.TestMode,
	}
	return nil
}

// Get looks up a transaction of any payment method by its ePayco reference,
// via POST /payment/transaction.
func (s *ChargeService) Get(ctx context.Context, referencePayco string) (*TransactionDetail, error) {
	body := struct {
		ReferencePayco string `json:"referencePayco"`
	}{ReferencePayco: referencePayco}

	var wrapped struct {
		Transaction TransactionDetail `json:"transaction"`
	}
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/transaction", body, &wrapped); err != nil {
		return nil, err
	}
	return &wrapped.Transaction, nil
}
