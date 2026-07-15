package epayco

import (
	"context"
	"encoding/json"
	"net/http"
)

// PSEService creates and confirms PSE (Pagos Seguros en Línea) bank-debit
// transactions through ePayco's apify API.
type PSEService struct {
	client *Client
}

// CreatePSEParams are the fields required to start a PSE transaction.
// URLResponse and URLConfirmation are required — confirmed live against the
// sandbox API ("field urlResponse required"), even though ePayco's own
// parameter docs list them as optional.
type CreatePSEParams struct {
	Bank            string `json:"bank"`
	Value           string `json:"value"`
	DocType         string `json:"docType"`
	DocNumber       string `json:"docNumber"`
	Name            string `json:"name"`
	LastName        string `json:"lastName"`
	Email           string `json:"email"`
	CellPhone       string `json:"cellPhone"`
	IP              string `json:"ip,omitempty"`
	URLResponse     string `json:"urlResponse"`
	URLConfirmation string `json:"urlConfirmation"`
}

// PSETransaction is the transaction created by Create, including the bank
// redirect URL the payer must be sent to.
type PSETransaction struct {
	RefPayco      int     `json:"ref_payco"`
	Invoice       string  `json:"factura"`
	Description   string  `json:"descripcion"`
	Value         float64 `json:"valor"`
	Tax           float64 `json:"iva"`
	TaxBase       float64 `json:"baseiva"`
	Currency      string  `json:"moneda"`
	State         string  `json:"estado"`
	Response      string  `json:"respuesta"`
	Authorization string  `json:"autorizacion"`
	Receipt       string  `json:"recibo"`
	Date          string  `json:"fecha"`
	BankURL       string  `json:"urlbanco"`
	TransactionID string  `json:"transactionID"`
	TicketID      string  `json:"ticketId"`
}

// Create starts a PSE transaction via POST /payment/process/pse. The payer
// must be redirected to PSETransaction.BankURL to complete the payment.
func (s *PSEService) Create(ctx context.Context, params CreatePSEParams) (*PSETransaction, error) {
	var result PSETransaction
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/process/pse", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfirmPSEParams identify the PSE transaction to confirm.
type ConfirmPSEParams struct {
	TransactionID int64 `json:"transactionID"`
}

// PSEConfirmation is the current status of a PSE transaction, returned by Confirm.
type PSEConfirmation struct {
	RefPayco      int     `json:"ref_payco"`
	Invoice       string  `json:"factura"`
	Description   string  `json:"descripcion"`
	Value         float64 `json:"valor"`
	Tax           float64 `json:"iva"`
	TaxBase       float64 `json:"baseiva"`
	Currency      string  `json:"moneda"`
	Bank          string  `json:"banco"`
	State         string  `json:"estado"`
	Response      string  `json:"respuesta"`
	Authorization string  `json:"autorizacion"`
	Receipt       string  `json:"recibo"`
	Date          string  `json:"fecha"`
	Franchise     string  `json:"franquicia"`
	DocType       string  `json:"tipo_doc"`
	Document      string  `json:"documento"`
	Names         string  `json:"nombres"`
	LastNames     string  `json:"apellidos"`
	Email         string  `json:"email"`
	City          string  `json:"ciudad"`
	Address       string  `json:"direccion"`
	TransactionID string  `json:"transactionID"`
	TicketID      int64   `json:"ticketId"`
}

// Confirm checks whether the bank debit for a PSE transaction went through,
// via POST /payment/pse/transaction.
func (s *PSEService) Confirm(ctx context.Context, params ConfirmPSEParams) (*PSEConfirmation, error) {
	var result PSEConfirmation
	if err := s.client.doApify(ctx, http.MethodPost, "/payment/pse/transaction", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Bank is a PSE-enabled bank, as returned by ListBanks.
type Bank struct {
	Code string `json:"bankCode"`
	Name string `json:"bankName"`
}

// UnmarshalJSON tolerates bankCode being a JSON string or a bare number.
// Confirmed live against the sandbox API: ePayco's production response mixes
// both within the same array (e.g. "bankCode":"1077" next to
// "bankCode":1022), so encoding/json's default struct decoding — which
// requires a consistent type — fails partway through the list.
func (b *Bank) UnmarshalJSON(data []byte) error {
	var raw struct {
		Code json.RawMessage `json:"bankCode"`
		Name string          `json:"bankName"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	b.Name = raw.Name
	b.Code = decodeFlexString(raw.Code)
	return nil
}

// ListBanks retrieves the banks available for PSE via GET /payment/pse/banks.
func (s *PSEService) ListBanks(ctx context.Context) ([]Bank, error) {
	var result []Bank
	if err := s.client.doApify(ctx, http.MethodGet, "/payment/pse/banks", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
