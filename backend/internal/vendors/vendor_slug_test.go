package vendors

import "testing"

func TestNormalizeSlug(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "lowercases and trims", in: "  My-Store  ", want: "my-store"},
		{name: "applies NFKC", in: "ＡＢＣ", want: "abc"},
		{name: "preserves internal hyphens", in: "ab-cd", want: "ab-cd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeSlug(tt.in); got != tt.want {
				t.Fatalf("NormalizeSlug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantSlug   string
		wantReason bool
	}{
		{name: "valid", in: "amrs-gadgets", wantSlug: "amrs-gadgets"},
		{name: "minimum length", in: "abc", wantSlug: "abc"},
		{name: "too short", in: "ab", wantSlug: "ab", wantReason: true},
		{name: "too long", in: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", wantSlug: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", wantReason: true},
		{name: "leading hyphen", in: "-store", wantSlug: "-store", wantReason: true},
		{name: "double hyphen", in: "my--store", wantSlug: "my--store", wantReason: true},
		{name: "underscore", in: "my_store", wantSlug: "my_store", wantReason: true},
		{name: "reserved", in: "Admin", wantSlug: "admin", wantReason: true},
		{name: "reserved locale", in: "en", wantSlug: "en", wantReason: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug, reason := validateSlug(tt.in)
			if slug != tt.wantSlug {
				t.Fatalf("validateSlug(%q) slug = %q, want %q", tt.in, slug, tt.wantSlug)
			}
			if (reason != "") != tt.wantReason {
				t.Fatalf("validateSlug(%q) reason = %q, wantReason %v", tt.in, reason, tt.wantReason)
			}
		})
	}
}
