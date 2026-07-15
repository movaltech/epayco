package epayco

import "fmt"

// SDK-level error codes, kept at parity with the errors.json table shared by the
// reference SDKs (epayco-node, epayco-php, epayco-python).
const (
	ErrCodeInvalidConfig     = 100
	ErrCodeCommunication     = 101
	ErrCodeUnknown           = 102
	ErrCodeInvalidData       = 103
	ErrCodeAuthentication    = 104
	ErrCodeNoCommunication   = 105
	ErrCodeInvalidPublicKey  = 106
	ErrCodeUnsupportedMethod = 107
	ErrCodeEncryption        = 108
	ErrCodeInvalidCashMethod = 109
)

// Error is the typed error returned by every operation in this SDK.
type Error struct {
	// Code is the SDK error code (see the ErrCode* constants above). Zero when the
	// error originates from an HTTP response instead of client-side validation.
	Code int
	// Message is a human-readable description of the failure.
	Message string
	// HTTPStatus is the HTTP status code returned by ePayco. Zero for client-side
	// errors that never reached the network (e.g. invalid configuration).
	HTTPStatus int
	// Err is the underlying error, if any (e.g. a network failure).
	Err error
}

func (e *Error) Error() string {
	switch {
	case e.HTTPStatus != 0:
		return fmt.Sprintf("epayco: %s (http %d)", e.Message, e.HTTPStatus)
	case e.Code != 0:
		return fmt.Sprintf("epayco: [%d] %s", e.Code, e.Message)
	default:
		return "epayco: " + e.Message
	}
}

func (e *Error) Unwrap() error { return e.Err }

// statusError maps an HTTP status code and an optional server-provided message to a
// typed *Error, mirroring the status-code handling in epayco-php's Client::request
// (epayco-node does not map status codes at all and just forwards {error: message}).
func statusError(status int, serverMessage string) *Error {
	msg := serverMessage
	switch status {
	case 400:
		if msg == "" {
			msg = "solicitud incorrecta, verifica los datos enviados"
		}
	case 401:
		if msg == "" {
			msg = "no autorizado, revisa tus credenciales"
		}
	case 403:
		if msg == "" {
			msg = "acceso prohibido, no tienes permisos para esta acción"
		}
	case 404:
		msg = "la ruta solicitada no existe"
	case 405:
		if msg == "" {
			msg = "método no permitido en esta ruta"
		}
	default:
		if msg == "" {
			msg = fmt.Sprintf("error inesperado del servidor (http %d)", status)
		}
	}
	return &Error{Message: msg, HTTPStatus: status}
}
