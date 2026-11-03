package vendors

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// ProductCatalog is the vendor portal's window into the product domain.
// Catalog rules (slug uniqueness, category/brand references, publishing
// preconditions) stay in that domain; this package only adds merchant
// standing and ownership checks before delegating.
type ProductCatalog interface {
	CreateProductAsDraft(ctx context.Context, input product.CreateProductInput) (*model.Product, error)
	UpdateProduct(ctx context.Context, productID uuid.UUID, input product.UpdateProductInput) error
	SoftDeleteProduct(ctx context.Context, productID uuid.UUID) error
	PublishProduct(ctx context.Context, productID uuid.UUID) error
	ArchiveProduct(ctx context.Context, productID uuid.UUID) error
	GetByID(ctx context.Context, productID uuid.UUID) (*model.Product, error)
	List(ctx context.Context, q *product.ListProductsQuery, includeDeleted bool) (*pagination.PagedResult[product.ProductCardReadModel], error)
}

// timeLayout is the wire format for vendor timestamps.
const timeLayout = time.RFC3339

// VendorProductSummary is the merchant-facing listing row: ownership and
// merchandising state, without storefront-only decoration.
type VendorProductSummary struct {
	ID           uuid.UUID               `json:"id"`
	Slug         string                  `json:"slug"`
	Title        string                  `json:"title"`
	Status       model.PublicationStatus `json:"status"`
	ProductType  model.ProductType       `json:"productType"`
	Price        *int64                  `json:"price,omitempty"`
	Currency     *string                 `json:"currency,omitempty"`
	SKU          *string                 `json:"sku,omitempty"`
	Stock        int                     `json:"stock"`
	ThumbnailURL string                  `json:"thumbnailUrl,omitempty"`
	CreatedAt    string                  `json:"createdAt"`
	UpdatedAt    string                  `json:"updatedAt"`
}

// requireActiveVendor resolves the caller's profile and enforces that the
// merchant may currently sell. It must run inside the caller's transaction so
// the check and the mutation observe the same state.
func (s *Service) requireActiveVendor(ctx context.Context, userID uuid.UUID) (*model.Vendor, error) {
	vendor, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
		}
		return nil, s.handleError(err, "Failed to load vendor")
	}

	if vendor.Status != model.VendorStatusActive {
		return nil, apierr.ForbiddenError("Vendor is not authorized for this action").
			WithCode(errcode.CodeVendorNotAuthorizedForAction)
	}

	return vendor, nil
}

// requireOwnedProduct resolves a product that belongs to the vendor. A
// product owned by someone else, or a platform product, is reported as not
// found so the portal never discloses the existence of foreign listings.
func (s *Service) requireOwnedProduct(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
) (*model.Vendor, error) {
	vendor, err := s.requireActiveVendor(ctx, userID)
	if err != nil {
		return nil, err
	}

	existing, err := s.products.GetByID(ctx, productID)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.NotFoundError("Product not found").WithCode(errcode.CodeProductNotFound)
		}
		return nil, s.handleError(err, "Failed to load product")
	}

	if existing.VendorID == nil || *existing.VendorID != vendor.ID {
		return nil, apierr.NotFoundError("Product not found").WithCode(errcode.CodeProductNotFound)
	}

	return vendor, nil
}

// CreateProduct creates a merchant-owned draft. Selling rights are checked in
// the same transaction as the insert.
func (s *Service) CreateProduct(
	ctx context.Context,
	userID uuid.UUID,
	input product.CreateProductInput,
) (*model.Product, error) {
	var created *model.Product
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		vendor, err := s.requireActiveVendor(ctx, userID)
		if err != nil {
			return err
		}

		input.VendorID = &vendor.ID
		created, err = s.products.CreateProductAsDraft(ctx, input)
		return err
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, s.handleError(err, "Failed to create product")
	}
	return created, nil
}

// UpdateProduct edits a merchant-owned product after an ownership check.
func (s *Service) UpdateProduct(ctx context.Context, userID, productID uuid.UUID, input product.UpdateProductInput) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.requireOwnedProduct(ctx, userID, productID); err != nil {
			return err
		}
		return s.products.UpdateProduct(ctx, productID, input)
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return s.handleError(err, "Failed to update product")
	}
	return nil
}

// DeleteProduct soft-deletes a merchant-owned product.
func (s *Service) DeleteProduct(ctx context.Context, userID, productID uuid.UUID) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.requireOwnedProduct(ctx, userID, productID); err != nil {
			return err
		}
		return s.products.SoftDeleteProduct(ctx, productID)
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return s.handleError(err, "Failed to delete product")
	}
	return nil
}

// PublishProduct publishes a merchant-owned product. Publishing preconditions
// (for example at least one variant) are owned by the product domain.
func (s *Service) PublishProduct(ctx context.Context, userID, productID uuid.UUID) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.requireOwnedProduct(ctx, userID, productID); err != nil {
			return err
		}
		return s.products.PublishProduct(ctx, productID)
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return s.handleError(err, "Failed to publish product")
	}
	return nil
}

// ArchiveProduct removes a merchant-owned product from the storefront.
func (s *Service) ArchiveProduct(ctx context.Context, userID, productID uuid.UUID) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.requireOwnedProduct(ctx, userID, productID); err != nil {
			return err
		}
		return s.products.ArchiveProduct(ctx, productID)
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return s.handleError(err, "Failed to archive product")
	}
	return nil
}

// ListOwnProducts returns the merchant's catalog slice (drafts included),
// narrowed to the caller's vendor regardless of any client filter.
func (s *Service) ListOwnProducts(
	ctx context.Context,
	userID uuid.UUID,
	q *product.ListProductsQuery,
) (*pagination.PagedResult[VendorProductSummary], error) {
	if q == nil {
		q = &product.ListProductsQuery{}
	}

	var result *pagination.PagedResult[VendorProductSummary]
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		vendor, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}

		q.VendorUUIDs = []uuid.UUID{vendor.ID}
		q.IncludeUnpublished = true

		page, err := s.products.List(ctx, q, false)
		if err != nil {
			return s.handleError(err, "Failed to list products")
		}

		result = &pagination.PagedResult[VendorProductSummary]{
			Items:  make([]*VendorProductSummary, 0, len(page.Items)),
			Page:   page.Page,
			Cursor: page.Cursor,
		}
		for _, card := range page.Items {
			result.Items = append(result.Items, s.toVendorProductSummary(card))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) toVendorProductSummary(card *product.ProductCardReadModel) *VendorProductSummary {
	summary := &VendorProductSummary{
		ID:          card.ID,
		Slug:        card.Slug,
		Title:       card.Title,
		Status:      card.Status,
		ProductType: card.ProductType,
		Price:       card.Price,
		Currency:    card.Currency,
		SKU:         card.SKU,
		Stock:       card.StockQuantity,
		CreatedAt:   card.CreatedAt.UTC().Format(timeLayout),
		UpdatedAt:   card.UpdatedAt.UTC().Format(timeLayout),
	}
	if card.Thumbnail != nil {
		summary.ThumbnailURL = s.storage.GetPublicURL(card.Thumbnail.Bucket, card.Thumbnail.Key)
	}
	return summary
}
