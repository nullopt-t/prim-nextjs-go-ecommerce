package product_test

// The public product detail payload must carry the fields the storefront PDP
// needs beyond the bare listing: the product's category (taxonomy links and
// attribute lookups), the default variant's SKU (product feeds and JSON-LD),
// and the authoritative rating aggregate.

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/model"
)

func detailsWith(variants []*model.ProductVariant, category *model.ProductCategory) *product.ProductDetails {
	return &product.ProductDetails{
		Product: &model.Product{
			ID:          uuid.New(),
			Slug:        "detail-product",
			Title:       "Detail Product",
			Status:      model.PublicationStatusPublished,
			ProductType: model.ProductTypeVariable,
		},
		Variants: variants,
		Category: category,
	}
}

func TestMapProductDetailsResponse_CarriesCategory(t *testing.T) {
	categoryID := uuid.New()
	details := detailsWith(nil, &model.ProductCategory{ID: categoryID, Name: "Shoes"})

	res := product.MapProductDetailsResponse(details, nil)

	require.NotNil(t, res.Category)
	require.Equal(t, categoryID.String(), res.Category.ID)
	require.Equal(t, "Shoes", res.Category.Name)
}

func TestMapProductDetailsResponse_CarriesDefaultVariantSKU(t *testing.T) {
	details := detailsWith([]*model.ProductVariant{
		{ID: uuid.New(), SKU: "SKU-OTHER", IsDefault: false},
		{ID: uuid.New(), SKU: "SKU-DEFAULT", IsDefault: true},
	}, nil)

	res := product.MapProductDetailsResponse(details, nil)

	require.NotNil(t, res.SKU)
	require.Equal(t, "SKU-DEFAULT", *res.SKU)
}

func TestMapProductDetailsResponse_NoSKUWithoutDefaultVariant(t *testing.T) {
	details := detailsWith([]*model.ProductVariant{
		{ID: uuid.New(), SKU: "SKU-OTHER", IsDefault: false},
	}, nil)

	res := product.MapProductDetailsResponse(details, nil)

	require.Nil(t, res.SKU, "a product without a default variant has no product-level SKU")
}

func TestMapProductDetailsResponse_CarriesRatingAggregate(t *testing.T) {
	details := detailsWith(nil, nil)
	summary := &model.RatingSummary{
		AverageRating: 4.5,
		ReviewCount:   24,
		Distribution:  map[int16]int{1: 1, 5: 13},
	}

	res := product.MapProductDetailsResponse(details, summary)

	require.NotNil(t, res.Rating)
	require.Equal(t, 4.5, res.Rating.AverageRating)
	require.Equal(t, 24, res.Rating.ReviewCount)
	require.Equal(t, 13, res.Rating.Distribution[5])
}

func TestMapProductDetailsResponse_DefaultVariantSetsProductPrice(t *testing.T) {
	price := int64(249900)
	crossedOut := int64(299900)
	details := detailsWith([]*model.ProductVariant{
		{ID: uuid.New(), Price: &price, IsDefault: true},
		{ID: uuid.New(), Price: &crossedOut, IsDefault: false},
	}, nil)

	res := product.MapProductDetailsResponse(details, nil)

	require.NotNil(t, res.ExtractedPrice)
	require.Equal(t, 2499.0, *res.ExtractedPrice)
}
