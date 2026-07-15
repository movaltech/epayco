package epayco

import (
	"errors"
	"testing"
)

func TestErrorMessageFormatting(t *testing.T) {
	cases := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "http status takes precedence",
			err:  &Error{Code: ErrCodeAuthentication, Message: "no autorizado", HTTPStatus: 401},
			want: "epayco: no autorizado (http 401)",
		},
		{
			name: "code only",
			err:  &Error{Code: ErrCodeInvalidConfig, Message: "apiKey and privateKey are required"},
			want: "epayco: [100] apiKey and privateKey are required",
		},
		{
			name: "message only",
			err:  &Error{Message: "something went wrong"},
			want: "epayco: something went wrong",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	sentinel := errors.New("network down")
	err := &Error{Code: ErrCodeNoCommunication, Message: "failed to reach ePayco", Err: sentinel}

	if !errors.Is(err, sentinel) {
		t.Errorf("errors.Is should find the wrapped sentinel error")
	}

	var target *Error
	if !errors.As(err, &target) {
		t.Fatalf("errors.As should match *Error")
	}
	if target.Code != ErrCodeNoCommunication {
		t.Errorf("unwrapped Code = %d, want %d", target.Code, ErrCodeNoCommunication)
	}
}

func TestStatusErrorDefaults(t *testing.T) {
	cases := []struct {
		status  int
		message string
		want    string
	}{
		{400, "", "solicitud incorrecta, verifica los datos enviados"},
		{400, "custom message", "custom message"},
		{401, "", "no autorizado, revisa tus credenciales"},
		{403, "", "acceso prohibido, no tienes permisos para esta acción"},
		{404, "should be ignored", "la ruta solicitada no existe"},
		{405, "", "método no permitido en esta ruta"},
		{500, "", "error inesperado del servidor (http 500)"},
	}

	for _, tc := range cases {
		got := statusError(tc.status, tc.message)
		if got.Message != tc.want {
			t.Errorf("statusError(%d, %q).Message = %q, want %q", tc.status, tc.message, got.Message, tc.want)
		}
		if got.HTTPStatus != tc.status {
			t.Errorf("statusError(%d, ...).HTTPStatus = %d, want %d", tc.status, got.HTTPStatus, tc.status)
		}
	}
}
