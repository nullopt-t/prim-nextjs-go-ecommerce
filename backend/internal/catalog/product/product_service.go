// Package product provides business logic orchestration, transaction boundaries,
// publication workflows, thumbnail uploading, and cross-subdomain integration for products.
package product

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/brand"
	"github.com/nullopt-t/prim-backend/internal/catalog/category"
	"github.com/nullopt-t/prim-backend/internal/catalog/errcode"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/review"
	"github.com/nullopt-t/prim-backend/internal/catalog/tag"
	"github.com/nullopt-t/prim-backend/internal/catalog/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	fileUtil "github.com/nullopt-t/prim-backend/pkg/file"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/utils"
)

// ObjectService defines storage operations for uploading, retrieving, and generating presigned URLs for media.
type ObjectService interface {
	UploadObject(ctx context.Context, contentType string, size int64, bucket string, file io.Reader) (*model.Object, error)
	GetObjectByID(ctx context.Context, id uuid.UUID) (*model.Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	DeleteObjectByID(ctx context.Context, id uuid.UUID) error
	GetObjectURL(ctx context.Context, bucket, key string) string
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
	dbRunner         database.Runner
	logger           log.Logger
	objectService    ObjectService
	productRepo      *ProductRepository
	brandService     *brand.BrandService
	categoryService  *category.CategoryService
	tagService       *tag.TagService
	variantService   *variant.VariantService
	inventoryService *inventory.InventoryService
	reviewService    *review.ReviewService
}

// NewService constructs a new ProductService with all required sub-domain services injected.
func NewService(
	r database.Runner,
	logger log.Logger,
	productRepo *ProductRepository,
	objectService ObjectService,
	brandService *brand.BrandService,
	categoryService *category.CategoryService,
	tagService *tag.TagService,
	variantService *variant.VariantService,
	inventoryService *inventory.InventoryService,
	reviewService *review.ReviewService,
) *ProductService {
	return &ProductService{
		dbRunner:         r,
		logger:           logger,
		objectService:    objectService,
		productRepo:      productRepo,
		brandService:     brandService,
		categoryService:  categoryService,
		tagService:       tagService,
		variantService:   variantService,
		inventoryService: inventoryService,
		reviewService:    reviewService,
	}
}

// CreateProductInput defines the creation parameters for a new draft product.
type CreateProductInput struct {
	BrandID     *uuid.UUID
	CategoryID  uuid.UUID
	Slug        *string
	Title       string
	Description *string
	ProductType string
}

// UpdateProductInput defines mutable fields for updating an existing product.
type UpdateProductInput struct {
	BrandID     *uuid.UUID
	CategoryID  *uuid.UUID
	Title       *string
	Description *string
	Highlights  []string
	ProductType *string
}

// CreateProductVariantInput defines parameters for creating a variant under an existing product.
type CreateProductVariantInput struct {
	Title           string
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	Attributes      map[string]any
	IsDefault       bool
	InitialStock    *int
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

// CreateProductAsDraft validates category and brand references, computes an initial slug from title,
// resolves any slug collisions across all products (including soft-deleted),
// and inserts the product into the database with PublicationStatusDraft.
func (s *ProductService) CreateProductAsDraft(
	ctx context.Context,
	input CreateProductInput,
) (*model.Product, error) {
	pt, err := model.ParseProductType(input.ProductType)
	if err != nil {
		return nil, apierr.BadRequestError("Invalid product type").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	}

	if input.CategoryID == uuid.Nil {
		return nil, apierr.BadRequestError("Category ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	// Validate category existence and active status
	if s.categoryService != nil {
		if _, err := s.categoryService.GetCategoryByID(ctx, input.CategoryID); err != nil {
			return nil, apierr.BadRequestError("Referenced category does not exist").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)
		}
	}

	// Validate brand existence and active status if specified
	if input.BrandID != nil && *input.BrandID != uuid.Nil && s.brandService != nil {
		if _, err := s.brandService.GetBrandByID(ctx, *input.BrandID); err != nil {
			return nil, apierr.BadRequestError("Referenced brand does not exist").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)
		}
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
	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		finalSlug := baseSlug
		// Check for slug collisions across all products (including soft-deleted)
		exists, err := s.productRepo.SlugExists(ctx, db, finalSlug)
		if err != nil {
			return err
		}
		if exists {
			for i := 1; i <= 10; i++ {
				candidate := fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:8])
				candidateExists, err := s.productRepo.SlugExists(ctx, db, candidate)
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
			ID:          uuid.New(),
			Slug:        finalSlug,
			BrandID:     input.BrandID,
			CategoryID:  input.CategoryID,
			Title:       input.Title,
			Description: input.Description,
			Status:      model.PublicationStatusDraft,
			ProductType: pt,
		}

		return s.productRepo.Create(ctx, db, product)
	})
	if err != nil {
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
	if input.CategoryID != nil && *input.CategoryID != uuid.Nil && s.categoryService != nil {
		if _, err := s.categoryService.GetCategoryByID(ctx, *input.CategoryID); err != nil {
			return apierr.BadRequestError("Referenced category does not exist").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)
		}
	}

	if input.BrandID != nil && *input.BrandID != uuid.Nil && s.brandService != nil {
		if _, err := s.brandService.GetBrandByID(ctx, *input.BrandID); err != nil {
			return apierr.BadRequestError("Referenced brand does not exist").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)
		}
	}

	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
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

		if input.Highlights != nil {
			product.Highlights = input.Highlights
		}

		if input.ProductType != nil {
			pt, err := model.ParseProductType(*input.ProductType)
			if err != nil {
				return apierr.BadRequestError("Invalid product type").
					WithCode(apierr.CodeInvalidInput).
					Wrap(err)
			}
			product.ProductType = pt
		}

		return s.productRepo.Update(ctx, tx, product)
	})

	if err != nil {
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

// GetByID retrieves a product by its UUID without hydrating associations.
func (s *ProductService) GetByID(
	ctx context.Context,
	productID uuid.UUID,
) (*model.Product, error) {
	var product *model.Product
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		product, err = s.productRepo.GetByID(ctx, db, productID)
		return err
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

	return product, nil
}

// GetAdminDetailsByID retrieves full administrative product details (including soft-deleted),
// hydrating brand, category, thumbnail, variants, and warehouse inventory stock breakdown.
func (s *ProductService) GetAdminDetailsByID(
	ctx context.Context,
	productID uuid.UUID,
) (*ProductDetails, error) {
	productDetails := &ProductDetails{}
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		prod, err := s.productRepo.GetByIDWithDeleted(ctx, db, productID)
		if err != nil {
			return err
		}
		productDetails.Product = prod
		return nil
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

	if productDetails.Product.BrandID != nil {
		brandObj, err := s.brandService.GetBrandByID(ctx, *productDetails.Product.BrandID)
		if err == nil {
			productDetails.Brand = brandObj
		}
	}

	if productDetails.Product.ThumbnailObjectID != nil {
		thumbObj, err := s.objectService.GetObjectByID(ctx, *productDetails.Product.ThumbnailObjectID)
		if err == nil {
			thumbObj.PublicURL = s.objectService.GetObjectURL(ctx, thumbObj.Bucket, thumbObj.Key)
			productDetails.Product.Thumbnail = thumbObj
		}
	}

	if productDetails.Product.CategoryID != uuid.Nil {
		catObj, err := s.categoryService.GetCategoryByID(ctx, productDetails.Product.CategoryID)
		if err == nil {
			productDetails.Category = catObj
		}
	}

	variantRes, err := s.variantService.ListVariantsByProductID(ctx, productDetails.Product.ID, &pagination.ListQuery{PageSize: 100}, true)
	if err == nil && variantRes != nil {
		productDetails.Variants = variantRes.Items
		if len(productDetails.Variants) > 0 && s.inventoryService != nil {
			vIDs := make([]uuid.UUID, 0, len(productDetails.Variants))
			for _, v := range productDetails.Variants {
				vIDs = append(vIDs, v.ID)
			}
			stocks, err := s.inventoryService.GetStockForVariants(ctx, vIDs)
			if err == nil {
				productDetails.Stock = stocks
			}
		}
	}

	tagsList, err := s.tagService.GetTagsByProductID(ctx, productDetails.Product.ID)
	if err == nil {
		productDetails.Tags = tagsList
	}

	return productDetails, nil
}

// GetBySlug retrieves public storefront details by product SEO slug.
func (s *ProductService) GetBySlug(
	ctx context.Context,
	slug string,
) (*ProductDetails, error) {
	cleanSlug := strings.TrimSpace(slug)
	productDetails := &ProductDetails{}
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		prod, err := s.productRepo.GetBySlug(ctx, db, cleanSlug)
		if err != nil {
			return err
		}
		productDetails.Product = prod
		return nil
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

	if productDetails.Product.BrandID != nil {
		brandObj, err := s.brandService.GetBrandByID(ctx, *productDetails.Product.BrandID)
		if err == nil {
			productDetails.Brand = brandObj
		}
	}

	if productDetails.Product.ThumbnailObjectID != nil {
		thumbObj, err := s.objectService.GetObjectByID(ctx, *productDetails.Product.ThumbnailObjectID)
		if err == nil {
			thumbObj.PublicURL = s.objectService.GetObjectURL(ctx, thumbObj.Bucket, thumbObj.Key)
			productDetails.Product.Thumbnail = thumbObj
		}
	}

	if productDetails.Product.CategoryID != uuid.Nil {
		catObj, err := s.categoryService.GetCategoryByID(ctx, productDetails.Product.CategoryID)
		if err == nil {
			productDetails.Category = catObj
		}
	}

	variantRes, err := s.variantService.ListVariantsByProductID(ctx, productDetails.Product.ID, &pagination.ListQuery{PageSize: 100}, false)
	if err == nil && variantRes != nil {
		productDetails.Variants = variantRes.Items
		if len(productDetails.Variants) > 0 && s.inventoryService != nil {
			vIDs := make([]uuid.UUID, 0, len(productDetails.Variants))
			for _, v := range productDetails.Variants {
				vIDs = append(vIDs, v.ID)
			}
			stocks, err := s.inventoryService.GetStockForVariants(ctx, vIDs)
			if err == nil {
				productDetails.Stock = stocks
			}
		}
	}

	tagsList, err := s.tagService.GetTagsByProductID(ctx, productDetails.Product.ID)
	if err == nil {
		productDetails.Tags = tagsList
	}

	return productDetails, nil
}

// AdminList retrieves paginated products for administrative management, resolving thumbnail URLs.
func (s *ProductService) AdminList(
	ctx context.Context,
	q *ListProductsQuery,
	includeDeleted bool,
) (*pagination.PagedResult[model.Product], error) {
	var res *pagination.PagedResult[model.Product]
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		res, err = s.productRepo.AdminList(ctx, db, q, includeDeleted)
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
		if item != nil && item.Thumbnail != nil {
			item.Thumbnail.PublicURL = s.objectService.GetObjectURL(ctx, item.Thumbnail.Bucket, item.Thumbnail.Key)
		}
	}

	return res, nil
}

// List executes a storefront product card query and hydrates review ratings asynchronously.
func (s *ProductService) List(
	ctx context.Context,
	q *ListProductsQuery,
	includeDeleted bool,
) (*pagination.PagedResult[ProductCardReadModel], error) {
	var res *pagination.PagedResult[ProductCardReadModel]
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		res, err = s.productRepo.List(ctx, db, q, includeDeleted)
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
		if item != nil && item.Thumbnail != nil {
			item.Thumbnail.PublicURL = s.objectService.GetObjectURL(ctx, item.Thumbnail.Bucket, item.Thumbnail.Key)
		}
	}

	if len(res.Items) > 0 {
		productIDs := make([]uuid.UUID, 0, len(res.Items))
		for _, item := range res.Items {
			if item != nil {
				productIDs = append(productIDs, item.ID)
			}
		}

		ratingsMap, err := s.reviewService.GetRatingSummariesByProductIDs(ctx, productIDs)
		if err == nil && ratingsMap != nil {
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
	variants, err := s.variantService.ListVariantsByProductID(ctx, productID, nil, false)
	if err != nil {
		return err
	}
	if len(variants.Items) == 0 {
		return apierr.BadRequestError("Product must have at least one variant before publishing").
			WithCode(errcode.CodePublishFailed)
	}

	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.UpdateStatus(ctx, db, productID, model.PublicationStatusPublished)
	})

	if err != nil {
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
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.UpdateStatus(ctx, db, productID, model.PublicationStatusArchived)
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
	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
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

// CreateProductVariant creates a variant under the given product and records initial stock if specified.
func (s *ProductService) CreateProductVariant(
	ctx context.Context,
	productID uuid.UUID,
	input CreateProductVariantInput,
) (*model.ProductVariant, error) {
	product, err := s.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	createdVariant, err := s.variantService.CreateVariant(ctx, &variant.CreateVariantInput{
		ProductID:       product.ID,
		Title:           input.Title,
		Price:           input.Price,
		CrossedOutPrice: input.CrossedOutPrice,
		Currency:        input.Currency,
		Attributes:      input.Attributes,
		IsDefault:       input.IsDefault,
	})
	if err != nil {
		return nil, err
	}

	if input.InitialStock != nil && *input.InitialStock > 0 && s.inventoryService != nil {
		_, err = s.inventoryService.AdjustStock(ctx, inventory.AdjustStockInput{
			VariantID: createdVariant.ID,
			Quantity:  *input.InitialStock,
			Reason:    string(model.InventoryReasonRestock),
		})
		if err != nil {
			// Roll back created variant on failure to maintain consistency
			_ = s.variantService.DeleteVariantByID(ctx, createdVariant.ID)
			return nil, err
		}
	}

	return createdVariant, nil
}

// GetProductVariants retrieves active variants for a product.
func (s *ProductService) GetProductVariants(
	ctx context.Context,
	productID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductVariant], error) {
	return s.variantService.ListVariantsByProductID(ctx, productID, q, false)
}

// SetDefaultVariant designates a variant as default for the given product.
func (s *ProductService) SetDefaultVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) error {
	return s.variantService.SetDefaultVariant(ctx, productID, variantID)
}

// ReplaceProductTags delegates tag association replacement to tagService.
func (s *ProductService) ReplaceProductTags(
	ctx context.Context,
	productID uuid.UUID,
	tagIDs []uuid.UUID,
) error {
	return s.tagService.ReplaceProductTags(ctx, productID, tagIDs)
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

	product, err := ps.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	bucket := "product-thumbnails"

	cleanCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	obj, err := ps.objectService.UploadObject(
		cleanCtx,
		detectedContentType,
		fileHeader.Size,
		bucket,
		file,
	)
	if err != nil {
		return nil, apierr.InternalServerError("Failed to upload file to storage").
			WithCode(apierr.CodeStorageError).
			Wrap(err).
			WithStack()
	}

	oldThumbnailObjectID := product.ThumbnailObjectID
	product.ThumbnailObjectID = &obj.ID
	err = ps.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return ps.productRepo.Update(ctx, tx, product)
	})

	if err != nil {
		_ = ps.objectService.DeleteObjectByID(ctx, obj.ID)
		return nil, apierr.InternalServerError("Failed to attach thumbnail to product").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	if oldThumbnailObjectID != nil && *oldThumbnailObjectID != obj.ID {
		_ = ps.objectService.DeleteObjectByID(ctx, *oldThumbnailObjectID)
	}

	obj.PublicURL = ps.objectService.GetObjectURL(ctx, obj.Bucket, obj.Key)
	return obj, nil
}