package vendors

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// VariantCatalog is the vendor portal's window into variant persistence.
// SKU, pricing, and attribute rules stay in the variant domain.
type VariantCatalog interface {
	CreateVariant(ctx context.Context, in *variant.CreateVariantInput) (*model.ProductVariant, error)
	UpdateVariant(ctx context.Context, variantID uuid.UUID, in variant.UpdateVariantInput) error
	DeleteVariantByID(ctx context.Context, variantID uuid.UUID) error
	ListVariantsByProductID(ctx context.Context, productID uuid.UUID, q *pagination.ListQuery, includeDeleted bool) (*pagination.PagedResult[model.ProductVariant], error)
	GetVariantByID(ctx context.Context, variantID uuid.UUID) (*model.ProductVariant, error)
}

// StockAdjuster records merchant stock movements (restock, returns, manual
// corrections). Reservations and sales stay checkout-owned.
type StockAdjuster interface {
	AdjustStock(ctx context.Context, in inventory.AdjustStockInput) (*model.InventoryStock, error)
}

// VendorVariantResponse is the merchant-facing variant payload.
type VendorVariantResponse struct {
	// Variant UUID.
	ID uuid.UUID `json:"id"`
	// Parent product UUID.
	ProductID uuid.UUID `json:"productId"`
	// Stock keeping unit.
	SKU string `json:"sku"`
	// Display title.
	Title string `json:"title"`
	// Unit price in minor units.
	Price *int64 `json:"price,omitempty"`
	// Crossed-out price in minor units.
	CrossedOutPrice *int64 `json:"crossedOutPrice,omitempty"`
	// ISO currency code.
	Currency *string `json:"currency,omitempty"`
	// Attribute dimensions.
	Attributes map[string]any `json:"attributes,omitempty"`
	// Default-selection flag.
	IsDefault bool `json:"isDefault"`
}

func toVendorVariantResponse(v *model.ProductVariant) VendorVariantResponse {
	return VendorVariantResponse{
		ID:              v.ID,
		ProductID:       v.ProductID,
		SKU:             v.SKU,
		Title:           v.Title,
		Price:           v.Price,
		CrossedOutPrice: v.CrossedOutPrice,
		Currency:        v.Currency,
		Attributes:      v.Attributes,
		IsDefault:       v.IsDefault,
	}
}

// requireOwnedVariant resolves a variant and proves its parent product belongs
// to the caller. Foreign and platform variants answer 404.
func (s *Service) requireOwnedVariant(
	ctx context.Context,
	userID uuid.UUID,
	variantID uuid.UUID,
) (*model.Vendor, *model.ProductVariant, error) {
	vendor, err := s.requireActiveVendor(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	existing, err := s.variants.GetVariantByID(ctx, variantID)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, nil, apierr.NotFoundError("Variant not found").WithCode(errcode.CodeVariantNotFound)
		}
		return nil, nil, s.handleError(err, "Failed to load variant")
	}

	product, err := s.products.GetByID(ctx, existing.ProductID)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, nil, apierr.NotFoundError("Variant not found").WithCode(errcode.CodeVariantNotFound)
		}
		return nil, nil, s.handleError(err, "Failed to load variant product")
	}

	if product.VendorID == nil || *product.VendorID != vendor.ID {
		return nil, nil, apierr.NotFoundError("Variant not found").WithCode(errcode.CodeVariantNotFound)
	}

	return vendor, existing, nil
}

// ListProductVariants lists the variants of a merchant-owned product.
func (s *Service) ListProductVariants(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[VendorVariantResponse], error) {
	var result *pagination.PagedResult[VendorVariantResponse]
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		if _, err := s.requireOwnedProduct(ctx, userID, productID); err != nil {
			return err
		}

		page, err := s.variants.ListVariantsByProductID(ctx, productID, q, false)
		if err != nil {
			return s.handleError(err, "Failed to list variants")
		}

		result = &pagination.PagedResult[VendorVariantResponse]{
			Items:  make([]*VendorVariantResponse, 0, len(page.Items)),
			Page:   page.Page,
			Cursor: page.Cursor,
		}
		for _, v := range page.Items {
			res := toVendorVariantResponse(v)
			result.Items = append(result.Items, &res)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateVariant adds a variant to a merchant-owned product.
func (s *Service) CreateVariant(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	in *variant.CreateVariantInput,
) (*model.ProductVariant, error) {
	var created *model.ProductVariant
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.requireOwnedProduct(ctx, userID, productID); err != nil {
			return err
		}

		in.ProductID = productID
		var err error
		created, err = s.variants.CreateVariant(ctx, in)
		return err
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, s.handleError(err, "Failed to create variant")
	}
	return created, nil
}

// UpdateVariant edits a variant of a merchant-owned product.
func (s *Service) UpdateVariant(ctx context.Context, userID, variantID uuid.UUID, in variant.UpdateVariantInput) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, _, err := s.requireOwnedVariant(ctx, userID, variantID); err != nil {
			return err
		}
		return s.variants.UpdateVariant(ctx, variantID, in)
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return s.handleError(err, "Failed to update variant")
	}
	return nil
}

// DeleteVariant soft-deletes a variant of a merchant-owned product.
func (s *Service) DeleteVariant(ctx context.Context, userID, variantID uuid.UUID) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, _, err := s.requireOwnedVariant(ctx, userID, variantID); err != nil {
			return err
		}
		return s.variants.DeleteVariantByID(ctx, variantID)
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return s.handleError(err, "Failed to delete variant")
	}
	return nil
}

// AdjustStock records a stock movement for a variant of a merchant-owned
// product. Sales and reservation movements remain checkout-owned.
func (s *Service) AdjustStock(
	ctx context.Context,
	userID uuid.UUID,
	in inventory.AdjustStockInput,
) (*model.InventoryStock, error) {
	var stock *model.InventoryStock
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, _, err := s.requireOwnedVariant(ctx, userID, in.VariantID); err != nil {
			return err
		}
		var err error
		stock, err = s.stock.AdjustStock(ctx, in)
		return err
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, s.handleError(err, "Failed to adjust stock")
	}
	return stock, nil
}
