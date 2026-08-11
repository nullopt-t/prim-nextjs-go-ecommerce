package product

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/brand"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/category"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/tag"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/variant"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	fileUtil "github.com/m-mahmoud-alsaid/prim-backend/pkg/file"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
)

type ObjectService interface {
	UploadObject(ctx context.Context, contentType string, size int64, bucket string, file io.Reader) (*model.Object, error)
	GetObjectByID(ctx context.Context, id uuid.UUID) (*model.Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	GetObjectURL(ctx context.Context, bucket, key string) string
}

func (s *ProductService) handleError(err error, defaultMsg string) error {
	if err == nil {
		return nil
	}
	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.ErrNotFound(defaultMsg + " not found").Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ErrConflict(defaultMsg + " conflict").Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.ErrBadRequest("Invalid reference for " + defaultMsg).Wrap(err)
	default:
		s.logger.Error("internal server error in product service", log.Meta{"error": err.Error()})
		return apierr.ErrInternalError("Failed to " + defaultMsg).Wrap(err).WithStack()
	}
} 


type ProductService struct {
	dbRunner        database.Runner
	logger          log.Logger
	objectService   ObjectService
	productRepo     *ProductRepository
	brandService    *brand.BrandService
	categoryService *category.CategoryService
	tagService      *tag.TagService
	variantService  *variant.VariantService
}

func NewService(
	r database.Runner,
	logger log.Logger,
	productRepo *ProductRepository,
	objectService ObjectService,
	brandService *brand.BrandService,
	categoryService *category.CategoryService,
	tagService *tag.TagService,
	variantService *variant.VariantService,
) *ProductService {
	return &ProductService{
		dbRunner:        r,
		logger:          logger,
		objectService:   objectService,
		productRepo:     productRepo,
		brandService:    brandService,
		categoryService: categoryService,
		tagService:      tagService,
		variantService:  variantService,
	}
}

type CreateProductInput struct {
	BrandID     *uuid.UUID
	CategoryID  uuid.UUID
	Title       string
	Description *string
	ProductType string
}

type UpdateProductInput struct {
	BrandID     *uuid.UUID
	CategoryID  *uuid.UUID
	Title       *string
	Description *string
	Highlights  []string
	ProductType *string
}

type CreateProductVariantInput struct {
	Title           string
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	Attributes      map[string]any
	IsDefault       bool
}

type ProductDetails struct {
	Product  *model.Product          `json:"product"`
	Brand    *model.ProductBrand     `json:"brand,omitempty"`
	Category *model.ProductCategory  `json:"category,omitempty"`
	Variants []*model.ProductVariant `json:"variants,omitempty"`
	Tags     []*model.ProductTag     `json:"tags,omitempty"`
}

func (s *ProductService) CreateProductAsDraft(
	ctx context.Context,
	input CreateProductInput,
) (*model.Product, error) {
	pt, err := model.ParseProductType(input.ProductType)
	if err != nil {
		return nil, apierr.ErrBadRequest("Invalid product type").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	}

	product := &model.Product{
		ID:          uuid.New(),
		Slug:        utils.Slugify(input.Title),
		BrandID:     input.BrandID,
		CategoryID:  input.CategoryID,
		Title:       input.Title,
		Description: input.Description,
		Status:      model.PublicationStatusDraft,
		ProductType: pt,
	}

	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.Create(ctx, db, product)
	})
	if err != nil {
		return nil, s.handleError(err, "create product")
	}

	return product, nil
}

func (s *ProductService) UpdateProduct(
	ctx context.Context,
	productID uuid.UUID,
	input UpdateProductInput,
) error {
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

		if input.Title != nil {
			product.Title = *input.Title
		}

		if input.Highlights != nil {
			product.Highlights = input.Highlights
		}

		if input.ProductType != nil {
			pt, err := model.ParseProductType(*input.ProductType)
			if err != nil {
				return apierr.ErrBadRequest("Invalid product type").
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
			return apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return apierr.ErrBadRequest("Referenced brand or category does not exist").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to update product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

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
			return nil, apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound)

		default:
			return nil, apierr.ErrInternalError("Failed to fetch product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return product, nil
}

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
			return nil, apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound)

		default:
			return nil, apierr.ErrInternalError("Failed to fetch product").
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
	}

	tagsList, err := s.tagService.GetTagsByProductID(ctx, productDetails.Product.ID)
	if err == nil {
		productDetails.Tags = tagsList
	}

	return productDetails, nil
}

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
			return nil, apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound)

		default:
			return nil, apierr.ErrInternalError("Failed to fetch product").
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
	}

	tagsList, err := s.tagService.GetTagsByProductID(ctx, productDetails.Product.ID)
	if err == nil {
		productDetails.Tags = tagsList
	}

	return productDetails, nil
}
func (s *ProductService) AdminList(
	ctx context.Context,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[model.Product], error) {
	var res *pagination.PagedResult[model.Product]
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		res, err = s.productRepo.AdminList(ctx, db, q, includeDeleted)
		return err
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to list products").
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
func (s *ProductService) List(
	ctx context.Context,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[ProductCardReadModel], error) {
	var res *pagination.PagedResult[ProductCardReadModel]
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		res, err = s.productRepo.List(ctx, db, q, includeDeleted)
		return err
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to list products").
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

func (s *ProductService) PublishProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	variants, err := s.variantService.ListVariantsByProductID(ctx, productID, nil, false)
	if err != nil {
		return err
	}
	if len(variants.Items) == 0 {
		return apierr.ErrBadRequest("Product must have at least one variant before publishing").
			WithCode(errcode.CodePublishFailed)
	}

	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.UpdateStatus(ctx, db, productID, model.PublicationStatusPublished)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to publish product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

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
			return apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to archive product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (s *ProductService) SoftDeleteProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.SoftDeleteByID(ctx, db, productID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to soft-delete product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (s *ProductService) CreateProductVariant(
	ctx context.Context,
	productID uuid.UUID,
	input CreateProductVariantInput,
) (*model.ProductVariant, error) {
	product, err := s.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return s.variantService.CreateVariant(ctx, &variant.CreateVariantInput{
		ProductID:       product.ID,
		Title:           input.Title,
		Price:           input.Price,
		CrossedOutPrice: input.CrossedOutPrice,
		Currency:        input.Currency,
		Attributes:      input.Attributes,
		IsDefault:       input.IsDefault,
	})
}

func (s *ProductService) GetProductVariants(
	ctx context.Context,
	productID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductVariant], error) {
	return s.variantService.ListVariantsByProductID(ctx, productID, q, false)
}

func (s *ProductService) SetDefaultVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) error {
	return s.variantService.SetDefaultVariant(ctx, productID, variantID)
}

func (s *ProductService) ReplaceProductTags(
	ctx context.Context,
	productID uuid.UUID,
	tagIDs []uuid.UUID,
) error {
	return s.tagService.ReplaceProductTags(ctx, productID, tagIDs)
}

func (ps *ProductService) UploadProductThumbnail(
	ctx context.Context,
	productID uuid.UUID,
	fileHeader *multipart.FileHeader,
) (*model.Object, error) {
	if productID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, apierr.ErrBadRequest("Failed to open uploaded file").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}
	defer func() {
		_ = file.Close()
	}()

	const maxFileSize = 10 * 1024 * 1024
	if fileHeader.Size > maxFileSize {
		return nil, apierr.ErrBadRequest("File size exceeds 10MB limit").
			WithCode(apierr.CodeFileTooLarge)
	}

	detectedContentType, err := fileUtil.DetectContentType(file)
	if err != nil {
		return nil, apierr.ErrBadRequest("Failed to detect file content type").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, apierr.ErrBadRequest("Failed to reset file reader").
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
		return nil, apierr.ErrInternalError("Failed to upload file to storage").
			WithCode(apierr.CodeStorageError).
			Wrap(err).
			WithStack()
	}

	product.ThumbnailObjectID = &obj.ID
	err = ps.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return ps.productRepo.Update(ctx, tx, product)
	})

	if err != nil {
		_ = ps.objectService.DeleteObject(ctx, obj.Bucket, obj.Key)
		return nil, apierr.ErrInternalError("Failed to attach thumbnail to product").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	obj.PublicURL = ps.objectService.GetObjectURL(ctx, obj.Bucket, obj.Key)
	return obj, nil
}