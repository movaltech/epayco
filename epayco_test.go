package epayco

import (
	"net/http"
	"testing"
)

func TestNewDefaults(t *testing.T) {
	c, err := New("public-key", "private-key")
	if err != nil {
		t.Fatalf("New returned unexpected error: %v", err)
	}
	if c.apiKey != "public-key" || c.privateKey != "private-key" {
		t.Fatalf("credentials not stored correctly: %+v", c)
	}
	if c.lang != "ES" {
		t.Errorf("lang default = %q, want ES", c.lang)
	}
	if c.baseURL != defaultBaseURL || c.apifyBaseURL != defaultApifyBaseURL {
		t.Errorf("base URLs not defaulted correctly: %+v", c)
	}
	if c.httpClient != http.DefaultClient {
		t.Errorf("httpClient default should be http.DefaultClient")
	}
	if c.auth == nil {
		t.Errorf("auth cache should be initialized")
	}
}

func TestNewRequiresCredentials(t *testing.T) {
	cases := []struct {
		name       string
		apiKey     string
		privateKey string
	}{
		{"empty both", "", ""},
		{"empty apiKey", "", "private-key"},
		{"empty privateKey", "public-key", ""},
		{"whitespace apiKey", "   ", "private-key"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.apiKey, tc.privateKey)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			epErr, ok := err.(*Error)
			if !ok {
				t.Fatalf("expected *Error, got %T", err)
			}
			if epErr.Code != ErrCodeInvalidConfig {
				t.Errorf("Code = %d, want %d", epErr.Code, ErrCodeInvalidConfig)
			}
		})
	}
}

func TestNewRejectsInvalidLang(t *testing.T) {
	_, err := New("public-key", "private-key", WithLang("FR"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	epErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if epErr.Code != ErrCodeInvalidConfig {
		t.Errorf("Code = %d, want %d", epErr.Code, ErrCodeInvalidConfig)
	}
}

func TestNewAppliesOptions(t *testing.T) {
	customClient := &http.Client{}
	c, err := New("public-key", "private-key",
		WithBaseURL("https://core.example.com"),
		WithApifyBaseURL("https://apify.example.com"),
		WithLang("EN"),
		WithRetries(3),
		WithHTTPClient(customClient),
	)
	if err != nil {
		t.Fatalf("New returned unexpected error: %v", err)
	}
	if c.baseURL != "https://core.example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
	if c.apifyBaseURL != "https://apify.example.com" {
		t.Errorf("apifyBaseURL = %q", c.apifyBaseURL)
	}
	if c.lang != "EN" {
		t.Errorf("lang = %q, want EN", c.lang)
	}
	if c.retries != 3 {
		t.Errorf("retries = %d, want 3", c.retries)
	}
	if c.httpClient != customClient {
		t.Errorf("httpClient not overridden")
	}
}
