package epayco

import (
	"context"
	"encoding/json"
)

// apifyEnvelope is the response shape shared by every endpoint on
// apify.epayco.co: {success, titleResponse, textResponse, lastAction, data}.
// Confirmed across PSE, cash, Daviplata, Safetypay, standard checkout, card
// tokenization and customer management in the official production Postman
// collection (docs/reference/API Services ePayco Producción.postman_collection.json).
//
// Crucially, ePayco answers with HTTP 200 even for validation or business-logic
// failures and signals the real outcome through "success" — the HTTP status
// code alone is not enough, so every apify call goes through doApify below
// instead of Client.do directly.
type apifyEnvelope struct {
	Success       bool            `json:"success"`
	TitleResponse string          `json:"titleResponse"`
	TextResponse  string          `json:"textResponse"`
	LastAction    string          `json:"lastAction"`
	Data          json.RawMessage `json:"data"`
}

// apifyErrorData models the error-data shapes observed across apify
// endpoints when success is false. "errors" and "error" are both
// json.RawMessage because ePayco encodes them inconsistently across
// endpoints — confirmed live against the sandbox API:
//   - {"errors": [{"codError":500,"errorMessage":"field x required"}]} (most
//     validation failures)
//   - {"errors": "field x required"} (a plain string instead — seen on
//     Customers.Update)
//   - {"error": "plain string"} or {"error": {"status":"error","description":"..."}}
//     (business errors on other endpoints)
//   - {"description": "..."} as a sibling of "errors" (Customers.Update again)
type apifyErrorData struct {
	Description string          `json:"description"`
	Errors      json.RawMessage `json:"errors"`
	Error       json.RawMessage `json:"error"`
}

// firstValidationError extracts the first message from an "errors" field,
// whether it's an array of {codError, errorMessage} objects or a bare string.
// Returns "" if raw is empty or neither shape applies.
func firstValidationError(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var list []struct {
		Message string `json:"errorMessage"`
	}
	if json.Unmarshal(raw, &list) == nil && len(list) > 0 {
		return list[0].Message
	}

	var plain string
	if json.Unmarshal(raw, &plain) == nil {
		return plain
	}
	return ""
}

// doApify calls an apify.epayco.co endpoint and decodes a business-logic
// success (HTTP 2xx and envelope.Success == true) into out. On business-logic
// failure (envelope.Success == false, still HTTP 200) it returns a descriptive
// *Error built from titleResponse/textResponse and the validation/error
// details nested in envelope.Data.
func (c *Client) doApify(ctx context.Context, method, path string, body, out any) error {
	var env apifyEnvelope
	if err := c.do(ctx, requestOptions{
		method: method,
		base:   c.apifyBaseURL,
		path:   path,
		apify:  true,
		body:   body,
		out:    &env,
	}); err != nil {
		return err
	}

	if !env.Success {
		return apifyBusinessError(env)
	}

	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return &Error{Code: ErrCodeUnknown, Message: "failed to decode apify response data", Err: err}
		}
	}
	return nil
}

// apifyBusinessError builds an *Error from a business-logic failure
// (envelope.Success == false). titleResponse/textResponse are NOT trustworthy
// on their own — confirmed live against the sandbox API, Customers.Update
// returns titleResponse:"Success"/textResponse:"Customer updated successfully"
// on a failed update, with the real reason only in data.errors. So the most
// specific available detail wins, in order: a validation error, a
// description, a business "error" field, and only then titleResponse/textResponse.
func apifyBusinessError(env apifyEnvelope) *Error {
	var details apifyErrorData
	if len(env.Data) > 0 {
		_ = json.Unmarshal(env.Data, &details) // best-effort; details stays zero on failure
	}

	msg := firstValidationError(details.Errors)

	if msg == "" {
		msg = details.Description
	}

	if msg == "" && len(details.Error) > 0 {
		var plain string
		if json.Unmarshal(details.Error, &plain) == nil && plain != "" {
			msg = plain
		} else {
			var described struct {
				Description string `json:"description"`
			}
			if json.Unmarshal(details.Error, &described) == nil {
				msg = described.Description
			}
		}
	}

	if msg == "" {
		msg = env.TextResponse
	}
	if msg == "" {
		msg = env.TitleResponse
	}
	if msg == "" {
		msg = "ePayco rejected the request"
	}
	return &Error{Message: msg}
}

// decodeFlexString decodes raw as a JSON string, falling back to its literal
// text if it's a bare JSON number. ePayco's production API is inconsistent
// about quoting numeric-looking fields — sometimes a decimal string like
// "20000.00", sometimes a bare number like 50000, for what should be the same
// field — confirmed live against the sandbox API for PSE bank codes and
// transaction lookups. Used by custom UnmarshalJSON methods (see Bank,
// TransactionDetail) to tolerate either encoding.
func decodeFlexString(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return string(raw)
}
