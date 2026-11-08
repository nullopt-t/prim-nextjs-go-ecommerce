package model_test

import (
	"errors"
	"testing"

	"github.com/nullopt-t/prim-backend/internal/model"
)

func TestParseProductType(t *testing.T) {
	t.Parallel()

	cases := map[string]model.ProductType{
		"simple":       model.ProductTypeSimple,
		"variable":     model.ProductTypeVariable,
		"downloadable": model.ProductTypeDownloadable,
	}
	for raw, want := range cases {
		got, err := model.ParseProductType(raw)
		if err != nil {
			t.Fatalf("ParseProductType(%q) error: %v", raw, err)
		}
		if got != want {
			t.Fatalf("ParseProductType(%q) = %q, want %q", raw, got, want)
		}
	}

	if _, err := model.ParseProductType("physical"); !errors.Is(err, model.ErrInvalidProductType) {
		t.Fatalf("ParseProductType(physical) error = %v, want ErrInvalidProductType", err)
	}
}
