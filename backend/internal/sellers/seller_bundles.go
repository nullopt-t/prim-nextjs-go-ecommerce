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
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
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

// UpdateBundleInput carries the bundle edit payload. Nil fields are unchanged.
type UpdateBundleInput struct {
	Title       *string
	BundlePrice  *int64
	Currency    *string
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

// isCurrencyCode reports whether value is an uppercase three-letter ISO 4217
// style code, matching the storefront's price formatting contract.
func isCurrencyCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
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

	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "EGP"
	}
	if !isCurrencyCode(currency) {
		return nil, apierr.BadRequestError("Currency must be a three-letter ISO code").
			WithCode(apierr.CodeInvalidInput)
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
			// A bundle may not contain another bundle: its backing product is
			// already a kit, so nesting would double the composition and the
			// price with no way to see either level on the order line.
			isBundle, err := s.bundles.IsBundleProduct(ctx, item.ProductID)
			if err != nil {
				return s.handleError(err, "Failed to check bundle item")
			}
			if isBundle {
				return apierr.BadRequestError("A bundle cannot include another bundle").
					WithCode(apierr.CodeInvalidInput)
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
		// The POST response must carry the same shape as the list projection,
		// including the backing product's lifecycle status.
		created.ProductSlug = createdProduct.Slug
		created.ProductTitle = createdProduct.Title
		created.Status = string(createdProduct.Status)
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

// UpdateBundle edits a bundle's title, price, and/or composition. The
// backing product and its selling variant follow the new values.
func (s *Service) UpdateBundle(
	ctx context.Context,
	userID, bundleID uuid.UUID,
	in UpdateBundleInput,
) error {
	if in.BundlePrice != nil && *in.BundlePrice <= 0 {
		return apierr.BadRequestError("Bundle price must be greater than zero").WithCode(apierr.CodeInvalidInput)
	}
	if in.Title != nil && strings.TrimSpace(*in.Title) == "" {
		return apierr.BadRequestError("Bundle title is required").WithCode(apierr.CodeInvalidInput)
	}

	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}

		productID, err := s.bundles.GetBundleProductID(ctx, seller.ID, bundleID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Bundle not found").WithCode(errcode.CodeProductNotFound)
			}
			return err
		}

		backing, err := s.products.GetByID(ctx, productID)
		if err != nil {
			return s.handleError(err, "Failed to load bundle product")
		}
		if backing.SellerID == nil || *backing.SellerID != seller.ID {
			return apierr.NotFoundError("Bundle not found").WithCode(errcode.CodeProductNotFound)
		}

		if len(in.Items) > 0 {
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
			}
		}

		var storeItems []BundleItem
		for _, item := range in.Items {
			storeItems = append(storeItems, BundleItem{
				ProductID: item.ProductID,
				VariantID: item.VariantID,
				Quantity:  item.Quantity,
			})
		}
		if err := s.bundles.UpdateBundle(ctx, seller.ID, bundleID, in.BundlePrice, in.Currency, len(in.Items) > 0, storeItems); err != nil {
			return err
		}

		if in.Title != nil {
			if err := s.products.UpdateProduct(ctx, productID, product.UpdateProductInput{Title: in.Title}); err != nil {
				return err
			}
		}
		if in.BundlePrice != nil {
			variants, err := s.variants.ListVariantsByProductID(ctx, productID, &pagination.ListQuery{PageSize: 1}, false)
			if err != nil {
				return err
			}
			if len(variants.Items) > 0 {
				currency := in.Currency
				if err := s.variants.UpdateVariant(ctx, variants.Items[0].ID, variant.UpdateVariantInput{
					Title:    in.Title,
					Price:    in.BundlePrice,
					Currency: currency,
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return s.handleError(err, "Failed to update bundle")
	}
	return nil
}

// DeleteBundle removes a bundle. A published bundle is archived (hidden from
// the storefront) instead of soft-deleted, so live listings are never yanked.
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
		product, err := s.products.GetByID(ctx, productID)
		if err != nil {
			return s.handleError(err, "Failed to load bundle product")
		}
		if product.Status == model.PublicationStatusPublished {
			return s.products.ArchiveProduct(ctx, productID)
		}
		return s.products.SoftDeleteProduct(ctx, productID)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete bundle")
	}
	return nil
}
