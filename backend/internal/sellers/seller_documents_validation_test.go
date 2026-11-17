package sellers

import (
	"strings"
	"testing"
)

func TestSanitizeDocumentFileName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"plain name", "invoice.pdf", "invoice.pdf", false},
		{"unix traversal", "../../etc/passwd.pdf", "passwd.pdf", false},
		{"windows traversal", `..\..\secrets\id.png`, "id.png", false},
		{"mixed directory and name", "docs/2026/tax card.pdf", "tax card.pdf", false},
		{"surrounding spaces", "  tax card.pdf  ", "tax card.pdf", false},
		{"blank", "   ", "", true},
		{"single dot", ".", "", true},
		{"double dot", "..", "", true},
		{"clamps overlong name", strings.Repeat("a", 300), strings.Repeat("a", 255), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeDocumentFileName(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("sanitizeDocumentFileName(%q) expected validation error, got %q", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("sanitizeDocumentFileName(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("sanitizeDocumentFileName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsAllowedKYCDocumentMIME(t *testing.T) {
	allowed := []string{
		"application/pdf",
		"image/jpeg",
		"image/png",
		"image/webp",
		"application/pdf; charset=utf-8",
		"IMAGE/PNG",
	}
	for _, ct := range allowed {
		if !isAllowedKYCDocumentMIME(ct) {
			t.Errorf("isAllowedKYCDocumentMIME(%q) = false, want true", ct)
		}
	}

	rejected := []string{
		"",
		"image/gif",
		"image/svg+xml",
		"image/avif",
		"application/octet-stream",
		"text/plain",
		"application/msword",
	}
	for _, ct := range rejected {
		if isAllowedKYCDocumentMIME(ct) {
			t.Errorf("isAllowedKYCDocumentMIME(%q) = true, want false", ct)
		}
	}
}

func TestValidateKYCUploadSize(t *testing.T) {
	if err := validateKYCUploadSize(0); err == nil {
		t.Error("validateKYCUploadSize(0) = nil, want rejection")
	}
	if err := validateKYCUploadSize(1); err != nil {
		t.Errorf("validateKYCUploadSize(1) = %v, want nil", err)
	}
	if err := validateKYCUploadSize(MaxKYCDocumentBytes); err != nil {
		t.Errorf("validateKYCUploadSize(10MiB) = %v, want nil", err)
	}
	if err := validateKYCUploadSize(MaxKYCDocumentBytes + 1); err == nil {
		t.Error("validateKYCUploadSize(10MiB+1) = nil, want rejection")
	}
}
