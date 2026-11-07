package sellers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// BundleItemInput is one component of a bundle being created.
type BundleItemInput struct {
	ProductID uuid.UUID
	VariantID *uuid.UUID
	Quantity  int
}

// CreateBundleInput carries the bundle definition from the handler.
type CreateBundleInput struct {
	Title       string
	Description *string
	BundlePrice int64
	Currency    string
	ThumbnailURL string
	Items       []BundleItemInput
}

// ListBundles returns the merchant's bundles with their composition.
func (s *Service) ListBundles(ctx context.Context, userID uuid.UUID) ([]SellerBundle, error) {
	var bundles []SellerBundle
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		bundles, err = s.bundles.ListBundles(ctx, seller.ID)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to list bundles")
	}
	return bundles, nil
}

// CreateBundle builds a kit from existing catalog items: it creates a
// sellable draft product with one variant priced at the bundle price,
// then records the composition.
func (s *Service) CreateBundle(
	ctx context.Context,
	userID uuid.UUID,
	in CreateBundleInput,
) (*SellerBundle, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, apierr.BadRequestError("Bundle title is required").WithCode(apierr.CodeInvalidInput)
	}
	if in.BundlePrice <= 0 {
		return nil, apierr.BadRequestError("Bundle price must be greater than zero").WithCode(apierr.CodeInvalidInput)
	}
	if len(in.Items) == 0 {
		return nil, apierr.BadRequestError("Bundle needs at least one item").WithCode(apierr.CodeInvalidInput)
	}

	var created *SellerBundle
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}

		// Every component must be the seller's own catalog item.
		var categoryID uuid.UUID
		items := make([]BundleItem, 0, len(in.Items))
		for _, item := range in.Items {
			if item.Quantity <= 0 {
				return apierr.BadRequestError("Bundle item quantity must be positive").WithCode(apierr.CodeInvalidInput)
			}
			component, err := s.products.GetByID(ctx, item.ProductID)
			if err != nil {
				if errors.Is(database.MapError(err), database.ErrNotFound) {
					return apierr.NotFoundError("Product not found").WithCode(errcode.CodeProductNotFound)
				}
				return s.handleError(err, "Failed to load bundle item product")
			}
			if component.SellerID == nil || *component.SellerID != seller.ID {
				return apierr.NotFoundError("Product not found").WithCode(errcode.CodeProductNotFound)
			}
			if item.VariantID != nil {
				variant, err := s.variants.GetVariantByID(ctx, *item.VariantID)
				if err != nil {
					if errors.Is(database.MapError(err), database.ErrNotFound) {
						return apierr.NotFoundError("Variant not found").WithCode(errcode.CodeVariantNotFound)
					}
					return s.handleError(err, "Failed to load bundle item variant")
				}
				if variant.ProductID != item.ProductID {
					return apierr.BadRequestError("Variant does not belong to the selected product").
						WithCode(apierr.CodeInvalidInput)
				}
			}
			if categoryID == uuid.Nil {
				categoryID = component.CategoryID
			}
			items = append(items, BundleItem{
				ProductID: item.ProductID,
				VariantID: item.VariantID,
				Quantity:  item.Quantity,
			})
		}

		var thumbnailID *uuid.UUID
		if strings.TrimSpace(in.ThumbnailURL) != "" {
			thumbnailID, err = s.resolveThumbnailURL(ctx, in.ThumbnailURL)
			if err != nil {
				return err
			}
		}

		createdProduct, err := s.products.CreateProductAsDraft(ctx, product.CreateProductInput{
			SellerID:          &seller.ID,
			CategoryID:        categoryID,
			Title:             strings.TrimSpace(in.Title),
			Description:       in.Description,
			ProductType:       string(model.ProductTypeSimple),
			ThumbnailObjectID: thumbnailID,
		})
		if err != nil {
			return err
		}

		currency := in.Currency
		if currency == "" {
			currency = "EGP"
		}
		if _, err := s.variants.CreateVariant(ctx, &variant.CreateVariantInput{
			ProductID:         createdProduct.ID,
			Title:             strings.TrimSpace(in.Title),
			Price:             &in.BundlePrice,
			Currency:          &currency,
			ThumbnailObjectID: thumbnailID,
		}); err != nil {
			return err
		}

		created, err = s.bundles.CreateBundle(ctx, seller.ID, in.BundlePrice, currency, createdProduct.ID, items)
		if err != nil {
			return err
		}
		created.ProductSlug = createdProduct.Slug
		created.ProductTitle = createdProduct.Title
		return nil
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, s.handleError(err, "Failed to create bundle")
	}
	return created, nil
}

// DeleteBundle removes a bundle and its backing draft/archive product.
func (s *Service) DeleteBundle(ctx context.Context, userID, bundleID uuid.UUID) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		productID, err := s.bundles.DeleteBundle(ctx, seller.ID, bundleID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Bundle not found").WithCode(errcode.CodeProductNotFound)
			}
			return err
		}
		return s.products.SoftDeleteProduct(ctx, productID)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete bundle")
	}
	return nil
}
