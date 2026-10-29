// Package product provides business logic orchestration, transaction boundaries,
// publication workflows, thumbnail uploading, and cross-subdomain integration for products.
package product

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/category/attribute"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	fileUtil "github.com/nullopt-t/prim-backend/pkg/file"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
	"github.com/nullopt-t/prim-backend/pkg/utils"
)

// BrandReader defines read methods for product brand associations.
type BrandReader interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.ProductBrand, error)
}

// CategoryReader defines read methods for product category associations.
type CategoryReader interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.ProductCategory, error)
}

// TagReader defines read methods for product tag queries.
type TagReader interface {
	GetTagsByProductID(ctx context.Context, qe database.QueryExecutor, productID uuid.UUID) ([]*model.ProductTag, error)
}

// TagWriter defines mutation methods for product tag associations.
type TagWriter interface {
	ReplaceProductTags(ctx context.Context, productID uuid.UUID, tagIDs []uuid.UUID) error
}

// InventoryReader defines stock lookup methods for variants.
type InventoryReader interface {
	GetStockForVariants(ctx context.Context, qe database.QueryExecutor, variantIDs []uuid.UUID) (map[uuid.UUID]*model.InventoryStock, error)
	GetStock(ctx context.Context, qe database.QueryExecutor, variantID uuid.UUID) (*model.InventoryStock, error)
}

// ReviewReader defines review aggregation methods for products.
type ReviewReader interface {
	GetRatingSummaryByProductID(ctx context.Context, qe database.QueryExecutor, productID uuid.UUID) (*model.RatingSummary, error)
	GetRatingSummariesByProductIDs(ctx context.Context, qe database.QueryExecutor, productIDs []uuid.UUID) (map[uuid.UUID]*model.RatingSummary, error)
	List(ctx context.Context, qe database.QueryExecutor, q *pagination.ListQuery, productID *uuid.UUID, userID *uuid.UUID, status *model.ReviewStatus) ([]model.Review, int, error)
}

// ObjectRepository defines persistence methods for storage media records.
type ObjectRepository interface {
	Create(ctx context.Context, tx *database.Tx, object *model.Object) error
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Object, error)
	GetByIDs(ctx context.Context, qe database.QueryExecutor, ids []uuid.UUID) ([]*model.Object, error)
	MarkDeletingByKey(ctx context.Context, tx *database.Tx, bucket, key string) error
	MarkDeletedByKey(ctx context.Context, tx *database.Tx, bucket, key string) error
}

// VariantReader defines read methods for product variant queries.
type VariantReader interface {
	CountByProductID(ctx context.Context, qe database.QueryExecutor, productID uuid.UUID) (int, error)
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.ProductVariant, error)
	ListByProductID(ctx context.Context, qe database.QueryExecutor, opts variant.ListVariantOptions) (*pagination.PagedResult[model.ProductVariant], error)
	ListMediaByVariantIDs(ctx context.Context, qe database.QueryExecutor, variantIDs []uuid.UUID) (map[uuid.UUID][]*model.VariantMedia, error)
}

// handleError maps database and driver errors to standard domain APIErrors.
func (s *ProductService) handleError(err error, defaultMsg string) error {
	if err == nil {
		return nil
	}
	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError(defaultMsg + " not found").Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError(defaultMsg + " conflict").Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Invalid reference for " + defaultMsg).Wrap(err)
	default:
		s.logger.Error("internal server error in product service", log.Meta{"error": err.Error()})
		return apierr.InternalServerError("Failed to " + defaultMsg).Wrap(err).WithStack()
	}
}

// ProductService manages product catalog lifecycles, cross-domain dependencies, and administration.
type ProductService struct {
	dbRunner        database.Runner
	logger          log.Logger
	productRepo     *ProductRepository
	brandReader     BrandReader
	categoryReader  CategoryReader
	tagReader       TagReader
	tagWriter       TagWriter
	inventoryReader InventoryReader
	reviewReader    ReviewReader
	objectRepo      ObjectRepository
	storageProvider storage.StorageProvider
	variantReader   VariantReader
}

// NewService constructs a new ProductService with decoupled repository interfaces and storage provider injected.
func NewService(
	r database.Runner,
	logger log.Logger,
	productRepo *ProductRepository,
	brandReader BrandReader,
	categoryReader CategoryReader,
	tagReader TagReader,
	tagWriter TagWriter,
	inventoryReader InventoryReader,
	reviewReader ReviewReader,
	objectRepo ObjectRepository,
	storageProvider storage.StorageProvider,
	variantReader VariantReader,
) *ProductService {
	if r == nil {
		panic("product: database runner is required")
	}
	if logger == nil {
		panic("product: logger is required")
	}
	if productRepo == nil {
		panic("product: product repository is required")
	}
	if brandReader == nil {
		panic("product: brand reader is required")
	}
	if categoryReader == nil {
		panic("product: category reader is required")
	}
	if tagReader == nil {
		panic("product: tag reader is required")
	}
	if tagWriter == nil {
		panic("product: tag writer is required")
	}
	if inventoryReader == nil {
		panic("product: inventory reader is required")
	}
	if reviewReader == nil {
		panic("product: review reader is required")
	}
	if objectRepo == nil {
		panic("product: object repository is required")
	}
	if storageProvider == nil {
		panic("product: storage provider is required")
	}
	if variantReader == nil {
		panic("product: variant reader is required")
	}
	return &ProductService{
		dbRunner:        r,
		logger:          logger,
		productRepo:     productRepo,
		brandReader:     brandReader,
		categoryReader:  categoryReader,
		tagReader:       tagReader,
		tagWriter:       tagWriter,
		inventoryReader: inventoryReader,
		reviewReader:    reviewReader,
		objectRepo:      objectRepo,
		storageProvider: storageProvider,
		variantReader:   variantReader,
	}
}

// CreateProductInput defines the creation parameters for a new draft product.
type CreateProductInput struct {
	BrandID       *uuid.UUID
	CategoryID    uuid.UUID
	Slug          *string
	Title         string
	Description   *string
	ProductType   string
	VariationAxes []string
}

// UpdateProductInput defines mutable fields for updating an existing product.
type UpdateProductInput struct {
	BrandID       *uuid.UUID
	CategoryID    *uuid.UUID
	Title         *string
	Description   *string
	Highlights    []string
	ProductType   *string
	VariationAxes *[]string
}


// ProductDetails aggregates full details of a product, its brand, category, variants, inventory stock, and tags.
type ProductDetails struct {
	Product  *model.Product                      `json:"product"`
	Brand    *model.ProductBrand                 `json:"brand,omitempty"`
	Category *model.ProductCategory              `json:"category,omitempty"`
	Variants []*model.ProductVariant             `json:"variants,omitempty"`
	Stock    map[uuid.UUID]*model.InventoryStock `json:"stock,omitempty"`
	Tags     []*model.ProductTag                 `json:"tags,omitempty"`
}

// normalizeVariationAxes trims, removes empty entries, and preserves order while
// rejecting duplicates. It returns a non-nil empty slice for empty input so the
// database always receives a valid text[] value.
func normalizeVariationAxes(axes []string) ([]string, error) {
	normalized := make([]string, 0, len(axes))
	seen := make(map[string]struct{}, len(axes))
	for _, raw := range axes {
		key := strings.TrimSpace(raw)
		if key == "" {
			return nil, apierr.BadRequestError("Variation axis keys cannot be empty").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{Field: "variationAxes", Message: "axis keys cannot be empty"})
		}
		if _, ok := seen[key]; ok {
			return nil, apierr.BadRequestError("Variation axes must be unique").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{Field: "variationAxes", Message: "duplicate axis: " + key})
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	return normalized, nil
}

// validateVariationAxes ensures every axis is a defined attribute for the
// product's category. Category attributes are the candidate pool from which an
// administrator chooses variation themes.
func (s *ProductService) validateVariationAxes(
	ctx context.Context,
	qe database.QueryExecutor,
	categoryID uuid.UUID,
	axes []string,
) ([]string, error) {
	normalized, err := normalizeVariationAxes(axes)
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return normalized, nil
	}

	defs, err := attribute.NewRepository().ListByCategory(ctx, qe, categoryID)
	if err != nil {
		return nil, err
	}
	available := make(map[string]struct{}, len(defs))
	for _, def := range defs {
		available[def.Key] = struct{}{}
	}

	for _, axis := range normalized {
		if _, ok := available[axis]; !ok {
			return nil, apierr.BadRequestError("Variation axis is not defined for this category").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{
					Field:   "variationAxes",
					Message: "axis " + axis + " is not a category attribute",
				})
		}
	}
	return normalized, nil
}

// CreateProductAsDraft validates category and brand references, computes an initial slug from title,
// resolves any slug collisions across all products (including soft-deleted),
// and inserts the product into the database with PublicationStatusDraft.
func (s *ProductService) CreateProductAsDraft(
	ctx context.Context,
	input CreateProductInput,
) (*model.Product, error) {
	pt := model.ProductTypeSimple
	if strings.TrimSpace(input.ProductType) != "" {
		parsed, err := model.ParseProductType(strings.TrimSpace(input.ProductType))
		if err != nil {
			return nil, apierr.BadRequestError("Invalid product type").
				WithCode(apierr.CodeInvalidInput).
				Wrap(err)
		}
		pt = parsed
	}

	if input.CategoryID == uuid.Nil {
		return nil, apierr.BadRequestError("Category ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	baseSlug := ""
	if input.Slug != nil {
		baseSlug = utils.Slugify(*input.Slug)
	}
	if baseSlug == "" {
		baseSlug = utils.Slugify(input.Title)
	}
	if baseSlug == "" {
		baseSlug = "product"
	}

	var product *model.Product
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		// Validate category existence and active status.
		if s.categoryReader != nil {
			if _, err := s.categoryReader.GetByID(ctx, tx, input.CategoryID); err != nil {
				return apierr.BadRequestError("Referenced category does not exist").
					WithCode(apierr.CodeInvalidReference).
					Wrap(err)
			}
		}

		// Validate brand existence and active status if specified.
		if input.BrandID != nil && *input.BrandID != uuid.Nil && s.brandReader != nil {
			if _, err := s.brandReader.GetByID(ctx, tx, *input.BrandID); err != nil {
				return apierr.BadRequestError("Referenced brand does not exist").
					WithCode(apierr.CodeInvalidReference).
					Wrap(err)
			}
		}

		axes, err := s.validateVariationAxes(ctx, tx, input.CategoryID, input.VariationAxes)
		if err != nil {
			return err
		}
		if len(axes) > 0 {
			pt = model.ProductTypeVariable
		} else if pt == model.ProductTypeVariable {
			return apierr.BadRequestError("Variable products must define at least one variation axis").
				WithCode(errcode.CodeVariationAxesInvalid).
				WithFields(api.FieldError{
					Field:   "variationAxes",
					Message: "variationAxes is required for variable products",
				})
		}

		finalSlug := baseSlug
		// Check for slug collisions across all products (including soft-deleted).
		exists, err := s.productRepo.SlugExists(ctx, tx, finalSlug)
		if err != nil {
			return err
		}
		if exists {
			for i := 1; i <= 10; i++ {
				candidate := fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:8])
				candidateExists, err := s.productRepo.SlugExists(ctx, tx, candidate)
				if err != nil {
					return err
				}
				if !candidateExists {
					finalSlug = candidate
					break
				}
			}
		}

		product = &model.Product{
			ID:            uuid.New(),
			Slug:          finalSlug,
			BrandID:       input.BrandID,
			CategoryID:    input.CategoryID,
			Title:         input.Title,
			Description:   input.Description,
			Status:        model.PublicationStatusDraft,
			ProductType:   pt,
			VariationAxes: axes,
		}

		return s.productRepo.Create(ctx, tx, product)
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, s.handleError(err, "create product")
	}

	return product, nil
}

// UpdateProduct modifies product properties inside a database transaction, validating category/brand references.
func (s *ProductService) UpdateProduct(
	ctx context.Context,
	productID uuid.UUID,
	input UpdateProductInput,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		if input.CategoryID != nil && *input.CategoryID != uuid.Nil && s.categoryReader != nil {
			if _, err := s.categoryReader.GetByID(ctx, tx, *input.CategoryID); err != nil {
				return apierr.BadRequestError("Referenced category does not exist").
					WithCode(apierr.CodeInvalidReference).
					Wrap(err)
			}
		}

		if input.BrandID != nil && *input.BrandID != uuid.Nil && s.brandReader != nil {
			if _, err := s.brandReader.GetByID(ctx, tx, *input.BrandID); err != nil {
				return apierr.BadRequestError("Referenced brand does not exist").
					WithCode(apierr.CodeInvalidReference).
					Wrap(err)
			}
		}

		// Serialize variation-theme changes against concurrent child creation.
		if input.VariationAxes != nil {
			var lockedID uuid.UUID
			if err := tx.QueryRow(ctx,
				`SELECT id FROM products WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`,
				productID,
			).Scan(&lockedID); err != nil {
				return err
			}
		}

		product, err := s.productRepo.GetByID(ctx, tx, productID)
		if err != nil {
			return err
		}

		if input.Title != nil {
			product.Title = *input.Title
		}

		if input.Description != nil {
			product.Description = input.Description
		}

		if input.BrandID != nil {
			product.BrandID = input.BrandID
		}

		if input.CategoryID != nil {
			product.CategoryID = *input.CategoryID
		}

		if input.VariationAxes != nil {
			axes, err := s.validateVariationAxes(ctx, tx, product.CategoryID, *input.VariationAxes)
			if err != nil {
				return err
			}

			if !slices.Equal(axes, product.VariationAxes) {
				var activeVariants int
				if err := tx.QueryRow(ctx,
					`SELECT count(*) FROM product_variants WHERE product_id = $1 AND deleted_at IS NULL`,
					product.ID,
				).Scan(&activeVariants); err != nil {
					return err
				}
				if activeVariants > 0 {
					return apierr.ConflictError("Variation axes cannot be changed while a product has child variants. Delete all children first.").
						WithCode(errcode.CodeVariationAxesLocked)
				}
			}

			product.VariationAxes = axes
			if len(axes) > 0 {
				product.ProductType = model.ProductTypeVariable
			} else {
				product.ProductType = model.ProductTypeSimple
			}
		} else if input.CategoryID != nil && len(product.VariationAxes) > 0 {
			// A category change must not silently leave axes that the new
			// category cannot describe.
			if _, err := s.validateVariationAxes(ctx, tx, product.CategoryID, product.VariationAxes); err != nil {
				return err
			}
		}

		if input.Highlights != nil {
			product.Highlights = input.Highlights
		}

		if input.ProductType != nil && input.VariationAxes == nil {
			pt, err := model.ParseProductType(*input.ProductType)
			if err != nil {
				return apierr.BadRequestError("Invalid product type").
					WithCode(apierr.CodeInvalidInput).
					Wrap(err)
			}
			if pt == model.ProductTypeVariable && len(product.VariationAxes) == 0 {
				return apierr.BadRequestError("Variable products must define at least one variation axis").
					WithCode(errcode.CodeVariationAxesInvalid)
			}
			if pt == model.ProductTypeSimple && len(product.VariationAxes) > 0 {
				return apierr.BadRequestError("Clear variation axes before making the product simple").
					WithCode(errcode.CodeVariationAxesInvalid)
			}
			product.ProductType = pt
		}

		return s.productRepo.Update(ctx, tx, product)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}

		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return apierr.BadRequestError("Referenced brand or category does not exist").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)

		default:
			return apierr.InternalServerError("Failed to update product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// GetByID fetches a single product by its unique identifier.
func (s *ProductService) GetByID(
	ctx context.Context,
	productID uuid.UUID,
) (*model.Product, error) {
	var product *model.Product
	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var repoErr error
		product, repoErr = s.getByID(ctx, exec, productID)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) getByID(
	ctx context.Context,
	exec database.QueryExecutor,
	productID uuid.UUID,
) (*model.Product, error) {
	product, err := s.productRepo.GetByID(ctx, exec, productID)
	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(mappedError, database.ErrNotFound):
			return nil, apierr.NotFoundError("Product not found").
				WithCode(errcode.CodeProductNotFound)

		default:
			return nil, apierr.InternalServerError("Failed to fetch product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return product, nil
}

// populateProductDetails hydrates brand, category, thumbnail, variants, inventory stock, and tags
// using the provided QueryExecutor so all lookups execute over a single database connection.
func (s *ProductService) populateProductDetails(
	ctx context.Context,
	qe database.QueryExecutor,
	prod *model.Product,
	includeDeletedVariants bool,
) (*ProductDetails, error) {
	details := &ProductDetails{
		Product: prod,
	}

	if prod.BrandID != nil && s.brandReader != nil {
		if brandObj, err := s.brandReader.GetByID(ctx, qe, *prod.BrandID); err == nil && brandObj != nil {
			if brandObj.LogoObjectID != nil && s.objectRepo != nil && s.storageProvider != nil {
				if logoObj, err := s.objectRepo.GetByID(ctx, qe, *brandObj.LogoObjectID); err == nil && logoObj != nil {
					logoURL := s.storageProvider.GetPublicURL(logoObj.Bucket, logoObj.Key)
					brandObj.LogoURL = &logoURL
				}
			}
			details.Brand = brandObj
		}
	}

	if prod.ThumbnailObjectID != nil && s.objectRepo != nil && s.storageProvider != nil {
		if thumbObj, err := s.objectRepo.GetByID(ctx, qe, *prod.ThumbnailObjectID); err == nil && thumbObj != nil {
			thumbObj.PublicURL = s.storageProvider.GetPublicURL(thumbObj.Bucket, thumbObj.Key)
			details.Product.Thumbnail = thumbObj
		}
	}

	if prod.CategoryID != uuid.Nil && s.categoryReader != nil {
		if catObj, err := s.categoryReader.GetByID(ctx, qe, prod.CategoryID); err == nil {
			details.Category = catObj
		}
	}

	if s.variantReader != nil {
		variantRes, err := s.variantReader.ListByProductID(ctx, qe, variant.ListVariantOptions{
			ProductID:      prod.ID,
			Query:          &pagination.ListQuery{PageSize: 100},
			IncludeDeleted: includeDeletedVariants,
		})
		if err == nil && variantRes != nil {
			details.Variants = variantRes.Items
			if len(details.Variants) > 0 {
				vIDs := make([]uuid.UUID, 0, len(details.Variants))
				for _, v := range details.Variants {
					vIDs = append(vIDs, v.ID)
				}

				// Hydrate variant thumbnails and gallery media with public URLs.
				// ListByProductID only loads the raw thumbnail object; media rows
				// live in variant_media and are fetched separately in one batch.
				if s.storageProvider != nil {
					for _, v := range details.Variants {
						if v.Thumbnail != nil {
							v.Thumbnail.PublicURL = s.storageProvider.GetPublicURL(v.Thumbnail.Bucket, v.Thumbnail.Key)
						}
					}
					if mediaByVariant, mediaErr := s.variantReader.ListMediaByVariantIDs(ctx, qe, vIDs); mediaErr == nil {
						for _, v := range details.Variants {
							media := mediaByVariant[v.ID]
							for _, m := range media {
								if m != nil && m.Object != nil {
									m.Object.PublicURL = s.storageProvider.GetPublicURL(m.Object.Bucket, m.Object.Key)
								}
							}
							v.Media = media
						}
					}
				}

				if s.inventoryReader != nil {
					if stocks, err := s.inventoryReader.GetStockForVariants(ctx, qe, vIDs); err == nil {
						details.Stock = stocks
					}
				}
			}
		}
	}

	if s.tagReader != nil {
		if tagsList, err := s.tagReader.GetTagsByProductID(ctx, qe, prod.ID); err == nil {
			details.Tags = tagsList
		}
	}

	return details, nil
}

// GetAdminDetailsByID retrieves full administrative product details (including soft-deleted),
// hydrating brand, category, thumbnail, variants, and warehouse inventory stock breakdown over a single database connection.
func (s *ProductService) GetAdminDetailsByID(
	ctx context.Context,
	productID uuid.UUID,
) (*ProductDetails, error) {
	var details *ProductDetails
	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		prod, err := s.productRepo.GetByIDWithDeleted(ctx, exec, productID)
		if err != nil {
			return err
		}
		var popErr error
		details, popErr = s.populateProductDetails(ctx, exec, prod, true)
		return popErr
	})

	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(mappedError, database.ErrNotFound):
			return nil, apierr.NotFoundError("Product not found").
				WithCode(errcode.CodeProductNotFound)

		default:
			return nil, apierr.InternalServerError("Failed to fetch product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return details, nil
}

// GetBySlug retrieves public storefront details by product SEO slug over a single database connection.
func (s *ProductService) GetBySlug(
	ctx context.Context,
	slug string,
) (*ProductDetails, error) {
	cleanSlug := strings.TrimSpace(slug)
	var details *ProductDetails
	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		prod, err := s.productRepo.GetBySlug(ctx, exec, cleanSlug)
		if err != nil {
			return err
		}
		var popErr error
		details, popErr = s.populateProductDetails(ctx, exec, prod, false)
		return popErr
	})

	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(mappedError, database.ErrNotFound):
			return nil, apierr.NotFoundError("Product not found").
				WithCode(errcode.CodeProductNotFound)

		default:
			return nil, apierr.InternalServerError("Failed to fetch product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return details, nil
}

// List executes a storefront product card query and hydrates review ratings asynchronously.
func (s *ProductService) List(
	ctx context.Context,
	q *ListProductsQuery,
	includeDeleted bool,
) (*pagination.PagedResult[ProductCardReadModel], error) {
	var res *pagination.PagedResult[ProductCardReadModel]
	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		res, err = s.productRepo.List(ctx, exec, q, includeDeleted)
		return err
	})

	if err != nil {
		return nil, apierr.InternalServerError("Failed to list products").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	for i := range res.Items {
		item := res.Items[i]
		if item != nil && item.Thumbnail != nil && s.storageProvider != nil {
			item.Thumbnail.PublicURL = s.storageProvider.GetPublicURL(item.Thumbnail.Bucket, item.Thumbnail.Key)
		}
	}

	if len(res.Items) > 0 && s.reviewReader != nil {
		productIDs := make([]uuid.UUID, 0, len(res.Items))
		for _, item := range res.Items {
			if item != nil {
				productIDs = append(productIDs, item.ID)
			}
		}

		var ratingsMap map[uuid.UUID]*model.RatingSummary
		_ = s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
			var err error
			ratingsMap, err = s.reviewReader.GetRatingSummariesByProductIDs(ctx, exec, productIDs)
			return err
		})
		if ratingsMap != nil {
			for _, item := range res.Items {
				if item != nil {
					if rating, ok := ratingsMap[item.ID]; ok {
						item.Rating = rating
					}
				}
			}
		}
	}

	return res, nil
}

// PublishProduct transitions a product from draft to published.
// Verifies that the product has at least one active variant.
func (s *ProductService) PublishProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, exec *database.Tx) error {
		count, err := s.variantReader.CountByProductID(ctx, exec, productID)
		if err != nil {
			return err
		}
		if count == 0 {
			return apierr.BadRequestError("Product must have at least one variant before publishing").
				WithCode(errcode.CodePublishFailed)
		}
		return s.productRepo.UpdateStatus(ctx, exec, productID, model.PublicationStatusPublished)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.InternalServerError("Failed to publish product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// ArchiveProduct sets a product's publication status to archived, hiding it from public search.
func (s *ProductService) ArchiveProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, exec *database.Tx) error {
		return s.productRepo.UpdateStatus(ctx, exec, productID, model.PublicationStatusArchived)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.InternalServerError("Failed to archive product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// SoftDeleteProduct marks a product and its associated variants as soft-deleted within a single transaction.
func (s *ProductService) SoftDeleteProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		if err := s.productRepo.SoftDeleteByID(ctx, tx, productID); err != nil {
			return err
		}

		// Cascade soft-deletion to variants of this product
		_, err := tx.Exec(ctx, `
			UPDATE product_variants
			SET deleted_at = now(), updated_at = now()
			WHERE product_id = $1 AND deleted_at IS NULL
		`, productID)
		if err != nil {
			return fmt.Errorf("cascade soft-delete variants: %w", err)
		}

		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.InternalServerError("Failed to soft-delete product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// SetDefaultVariant designates a variant as default for the given product.
func (s *ProductService) SetDefaultVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) error {
	if productID == uuid.Nil || variantID == uuid.Nil {
		return apierr.BadRequestError("Product ID and Variant ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		v, err := s.variantReader.GetByID(ctx, tx, variantID)
		if err != nil {
			return err
		}
		if v.ProductID != productID {
			return apierr.BadRequestError("Variant does not belong to this product").
				WithCode(errcode.CodeVariantProductMismatch)
		}

		prod, err := s.productRepo.GetByID(ctx, tx, productID)
		if err != nil {
			return err
		}
		if prod.ProductType == model.ProductTypeSimple && prod.DefaultVariantID != nil && *prod.DefaultVariantID != variantID {
			return apierr.BadRequestError("Default variant for a simple product cannot be changed").
				WithCode(errcode.CodeSimpleProductVariantLocked)
		}

		return s.productRepo.SetDefaultVariant(ctx, tx, productID, &variantID)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Product or variant not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)
		default:
			return apierr.InternalServerError("Failed to set default variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// ReplaceProductTags replaces tag associations for a product atomically through the tag domain service.
func (s *ProductService) ReplaceProductTags(
	ctx context.Context,
	productID uuid.UUID,
	tagIDs []uuid.UUID,
) error {
	return s.tagWriter.ReplaceProductTags(ctx, productID, tagIDs)
}

// UploadProductThumbnail handles multipart thumbnail upload, size and MIME type validation,
// S3/MinIO upload, and product table attachment, safely deleting any previously uploaded thumbnail.
func (ps *ProductService) UploadProductThumbnail(
	ctx context.Context,
	productID uuid.UUID,
	fileHeader *multipart.FileHeader,
) (*model.Object, error) {
	if productID == uuid.Nil {
		return nil, apierr.BadRequestError("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, apierr.BadRequestError("Failed to open uploaded file").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}
	defer func() {
		_ = file.Close()
	}()

	const maxFileSize = 10 * 1024 * 1024
	if fileHeader.Size > maxFileSize {
		return nil, apierr.BadRequestError("File size exceeds 10MB limit").
			WithCode(apierr.CodeFileTooLarge)
	}

	detectedContentType, err := fileUtil.DetectContentType(file)
	if err != nil {
		return nil, apierr.BadRequestError("Failed to detect file content type").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, apierr.BadRequestError("Failed to reset file reader").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	if !fileUtil.IsAllowedImageMime(detectedContentType) {
		return nil, apierr.BadRequestError(
			"Only JPEG, PNG, WebP, and AVIF images are allowed",
		).WithCode(apierr.CodeUnsupportedMediaType)
	}

	product, err := ps.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	bucket := "product-thumbnails"
	key := uuid.New().String()

	cleanCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := ps.storageProvider.Upload(cleanCtx, bucket, key, file, fileHeader.Size, detectedContentType); err != nil {
		return nil, apierr.InternalServerError("Failed to upload file to storage").
			WithCode(apierr.CodeStorageError).
			Wrap(err).
			WithStack()
	}

	obj := &model.Object{
		ID:          uuid.New(),
		Bucket:      bucket,
		Key:         key,
		ContentType: detectedContentType,
		FileSize:    fileHeader.Size,
		Status:      model.ObjectStatusUploaded,
	}

	oldThumbnailObjectID := product.ThumbnailObjectID
	product.ThumbnailObjectID = &obj.ID
	var oldObj *model.Object
	txErr := ps.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		if err := ps.objectRepo.Create(ctx, tx, obj); err != nil {
			return err
		}
		if oldThumbnailObjectID != nil && *oldThumbnailObjectID != obj.ID {
			oldObj, _ = ps.objectRepo.GetByID(ctx, tx, *oldThumbnailObjectID)
			if oldObj != nil {
				_ = ps.objectRepo.MarkDeletingByKey(ctx, tx, oldObj.Bucket, oldObj.Key)
			}
		}
		return ps.productRepo.Update(ctx, tx, product)
	})

	if txErr != nil {
		_ = ps.storageProvider.Delete(ctx, bucket, key)
		return nil, ps.handleError(txErr, "attach thumbnail to product")
	}

	if oldObj != nil && ps.storageProvider != nil {
		if delErr := ps.storageProvider.Delete(ctx, oldObj.Bucket, oldObj.Key); delErr == nil {
			_ = ps.dbRunner.WithTx(ctx, func(ctx context.Context, exec *database.Tx) error {
				return ps.objectRepo.MarkDeletedByKey(ctx, exec, oldObj.Bucket, oldObj.Key)
			})
		}
	}

	obj.PublicURL = ps.storageProvider.GetPublicURL(obj.Bucket, obj.Key)
	return obj, nil
}

// DeleteProductThumbnail unlinks the product's thumbnail and removes the storage object.
func (ps *ProductService) DeleteProductThumbnail(ctx context.Context, productID uuid.UUID) error {
	if productID == uuid.Nil {
		return apierr.BadRequestError("Product ID is required").WithCode(apierr.CodeInvalidInput)
	}

	product, err := ps.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	if product.ThumbnailObjectID == nil {
		return nil
	}

	var oldObj *model.Object
	oldID := *product.ThumbnailObjectID
	product.ThumbnailObjectID = nil
	err = ps.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		if ps.objectRepo != nil {
			oldObj, _ = ps.objectRepo.GetByID(ctx, tx, oldID)
			if oldObj != nil {
				_ = ps.objectRepo.MarkDeletingByKey(ctx, tx, oldObj.Bucket, oldObj.Key)
			}
		}
		return ps.productRepo.Update(ctx, tx, product)
	})
	if err != nil {
		return apierr.InternalServerError("Failed to remove product thumbnail").Wrap(err)
	}

	if oldObj != nil && ps.storageProvider != nil {
		if delErr := ps.storageProvider.Delete(ctx, oldObj.Bucket, oldObj.Key); delErr == nil {
			_ = ps.dbRunner.WithTx(ctx, func(ctx context.Context, exec *database.Tx) error {
				return ps.objectRepo.MarkDeletedByKey(ctx, exec, oldObj.Bucket, oldObj.Key)
			})
		}
	}
	return nil
}

// ListProductReviews retrieves paginated reviews for a product using the review repository.
func (s *ProductService) ListProductReviews(
	ctx context.Context,
	q *pagination.ListQuery,
	productID uuid.UUID,
	status *model.ReviewStatus,
) (*pagination.PagedResult[model.Review], error) {
	var items []model.Review
	var total int
	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		items, total, err = s.reviewReader.List(ctx, exec, q, &productID, nil, status)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "list reviews")
	}

	resultItems := make([]*model.Review, len(items))
	for i := range items {
		resultItems[i] = &items[i]
	}

	return pagination.NewPagedResult(resultItems, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// GetProductRatingSummary returns aggregated star rating statistics for a product.
func (s *ProductService) GetProductRatingSummary(
	ctx context.Context,
	productID uuid.UUID,
) (*model.RatingSummary, error) {
	var summary *model.RatingSummary
	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		summary, err = s.reviewReader.GetRatingSummaryByProductID(ctx, exec, productID)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "get rating summary")
	}
	return summary, nil
}
