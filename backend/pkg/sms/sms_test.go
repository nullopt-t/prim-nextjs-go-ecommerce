package sms

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nullopt-t/prim-backend/pkg/log"
)

type noopLogger struct{}

func (n noopLogger) Debug(string, ...log.Meta) {}
func (n noopLogger) Info(string, ...log.Meta)  {}
func (n noopLogger) Warn(string, ...log.Meta)  {}
func (n noopLogger) Error(string, ...log.Meta) {}

func TestLogSender(t *testing.T) {
	sender := NewLogSender(noopLogger{})
	err := sender.Send(context.Background(), "+201012345678", "123456")
	if err != nil {
		t.Fatalf("expected nil error from LogSender, got %v", err)
	}
}

func TestTwilioSender_Success(t *testing.T) {
	var requestedPath string
	var requestedAuthUser, requestedAuthPass string
	var requestedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedAuthUser, requestedAuthPass, _ = r.BasicAuth()
		b, _ := io.ReadAll(r.Body)
		requestedBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"queued"}`))
	}))
	defer server.Close()

	sender := NewTwilioSender(
		"AC1234567890abcdef",
		"auth_secret_token",
		"+15005550006",
		noopLogger{},
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)

	err := sender.Send(context.Background(), "+201012345678", "Your code is: 987654")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := "/2010-04-01/Accounts/AC1234567890abcdef/Messages.json"
	if requestedPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, requestedPath)
	}
	if requestedAuthUser != "AC1234567890abcdef" || requestedAuthPass != "auth_secret_token" {
		t.Errorf("basic auth mismatch")
	}
	if requestedBody != "Body=Your+code+is%3A+987654&From=%2B15005550006&To=%2B201012345678" {
		t.Errorf("unexpected body: %s", requestedBody)
	}
}

func TestTwilioSender_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code": 21211, "message": "Invalid 'To' Phone Number"}`))
	}))
	defer server.Close()

	sender := NewTwilioSender(
		"AC1234567890abcdef",
		"auth_secret_token",
		"+15005550006",
		noopLogger{},
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)

	err := sender.Send(context.Background(), "invalid-phone", "Your code is: 987654")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewSender_Factory(t *testing.T) {
	// Log sender fallback
	sender := NewSender(Config{Provider: "log"}, noopLogger{})
	if _, ok := sender.(*LogSender); !ok {
		t.Errorf("expected *LogSender, got %T", sender)
	}

	// Incomplete Twilio config falls back to log
	sender = NewSender(Config{Provider: "twilio", TwilioAccountSID: "AC123"}, noopLogger{})
	if _, ok := sender.(*LogSender); !ok {
		t.Errorf("expected *LogSender fallback, got %T", sender)
	}

	// Complete Twilio config
	sender = NewSender(Config{
		Provider:         "twilio",
		TwilioAccountSID: "AC123",
		TwilioAuthToken:  "token",
		TwilioFromPhone:  "+15005550006",
	}, noopLogger{})
	if _, ok := sender.(*TwilioSender); !ok {
		t.Errorf("expected *TwilioSender, got %T", sender)
	}
}
