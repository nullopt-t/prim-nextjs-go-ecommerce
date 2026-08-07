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
)

type ObjectProvider interface {
	UploadObject(ctx context.Context, contentType string, size int64, bucket string, file io.Reader) (*model.Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	GetObjectURL(ctx context.Context, bucket, key string) string
}

type ProductService struct {
	dr              database.Runner
	logger          log.Logger
	objectService   ObjectProvider
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
	objectService ObjectProvider,
	brandService *brand.BrandService,
	categoryService *category.CategoryService,
	tagService *tag.TagService,
	variantService *variant.VariantService,
) *ProductService {
	return &ProductService{
		dr:              r,
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
	Description string
	Highlights  []string
}

type UpdateProductInput struct {
	BrandID     *uuid.UUID
	CategoryID  *uuid.UUID
	Title       *string
	Description *string
	Highlights  []string
	Status      *model.PublicationStatus
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
	Media    []*model.ProductMedia   `json:"media,omitempty"`
	Variants []*model.ProductVariant `json:"variants,omitempty"`
	Tags     []*model.ProductTag     `json:"tags,omitempty"`
}

func (s *ProductService) CreateProductAsDraft(
	ctx context.Context,
	input CreateProductInput,
) (*model.Product, error) {
	product := &model.Product{
		ID:          uuid.New(),
		BrandID:     input.BrandID,
		CategoryID:  input.CategoryID,
		PublicID:    uuid.NewString(),
		Title:       input.Title,
		Description: input.Description,
		Status:      model.PublicationStatusDraft,
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.Create(ctx, db, product)
	})
	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, apierr.ErrConflict("Product with this public_id already exists").
				WithCode(errcode.CodeProductAlreadyExists).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.ErrBadRequest("Invalid brand or category reference").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)

		default:
			return nil, apierr.ErrInternalError("Failed to create product").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return product, nil
}

func (s *ProductService) UpdateProduct(
	ctx context.Context,
	productID uuid.UUID,
	input UpdateProductInput,
) error {
	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		product, err := s.productRepo.GetByID(ctx, tx, productID)
		if err != nil {
			return err
		}

		if input.Title != nil {
			product.Title = *input.Title
		}

		if input.Description != nil {
			product.Description = *input.Description
		}

		if input.BrandID != nil {
			product.BrandID = input.BrandID
		}

		if input.CategoryID != nil && *input.CategoryID != uuid.Nil {
			product.CategoryID = *input.CategoryID
		}

		if input.Highlights != nil {
			product.Highlights = input.Highlights
		}

		if input.Status != nil {
			product.Status = *input.Status
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
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
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
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
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

	if productDetails.Product.CategoryID != uuid.Nil {
		catObj, err := s.categoryService.GetCategoryByID(ctx, productDetails.Product.CategoryID)
		if err == nil {
			productDetails.Category = catObj
		}
	}

	mediaList, err := s.GetProductMedia(ctx, productDetails.Product.ID)
	if err == nil {
		productDetails.Media = mediaList
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

func (s *ProductService) GetByPublicID(
	ctx context.Context,
	publicID string,
) (*ProductDetails, error) {
	cleanPID := strings.TrimSpace(publicID)
	productDetails := &ProductDetails{}
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		prod, err := s.productRepo.GetByPublicID(ctx, db, cleanPID)
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

	if productDetails.Product.CategoryID != uuid.Nil {
		catObj, err := s.categoryService.GetCategoryByID(ctx, productDetails.Product.CategoryID)
		if err == nil {
			productDetails.Category = catObj
		}
	}

	mediaList, err := s.GetProductMedia(ctx, productDetails.Product.ID)
	if err == nil {
		productDetails.Media = mediaList
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
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
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

	return res, nil
}
func (s *ProductService) List(
	ctx context.Context,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[PublicProductListReadModel], error) {
	var res *pagination.PagedResult[PublicProductListReadModel]
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
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

	err = s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
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
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
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
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
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

func (ps *ProductService) UploadProductMedia(
	ctx context.Context,
	productID uuid.UUID,
	fileHeader *multipart.FileHeader,
) (*model.ProductMedia, error) {
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
	defer file.Close()

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

	mediaType, err := model.ParseMediaType(detectedContentType)
	if err != nil {
		return nil, apierr.ErrBadRequest("Unsupported media format").
			WithCode(apierr.CodeUnsupportedMediaType).
			Wrap(err)
	}

	_, err = ps.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	bucket := "product-media"

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

	var media *model.ProductMedia
	err = ps.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		maxOrder, err := ps.productRepo.GetMaxMediaSortOrder(ctx, tx, productID)
		if err != nil {
			return err
		}

		var repoErr error
		media, repoErr = ps.productRepo.AddMedia(ctx, tx, CreateProductMediaInput{
			ProductID:       productID,
			StorageObjectID: obj.ID,
			MediaType:       mediaType,
			SortOrder:       maxOrder + 1,
		})
		return repoErr
	})

	if err != nil {
		_ = ps.objectService.DeleteObject(ctx, obj.Bucket, obj.Key)
		return nil, apierr.ErrInternalError("Failed to attach media to product").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return media, nil
}


func (ps *ProductService) GetProductMedia(
	ctx context.Context,
	productID uuid.UUID,
) ([]*model.ProductMedia, error) {
	if productID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var mediaList []*model.ProductMedia
	err := ps.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		mediaList, err = ps.productRepo.ListMediaByProductID(ctx, db, productID)
		return err
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to fetch product media").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	for _, m := range mediaList {
		if m.Object != nil && m.Object.PublicURL == "" && m.Object.Bucket != "" && m.Object.Key != "" {
			url := ps.objectService.GetObjectURL(
				ctx,
				m.Object.Bucket,
				m.Object.Key,
			)
			if url != "" {
				m.Object.PublicURL = url
			}
		}
	}

	return mediaList, nil
}

func (ps *ProductService) DetachMedia(
	ctx context.Context,
	productID uuid.UUID,
	mediaID uuid.UUID,
) error {
	if productID == uuid.Nil || mediaID == uuid.Nil {
		return apierr.ErrBadRequest("Product ID and Media ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	var objectBucket, objectKey string
	err := ps.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		mediaList, err := ps.productRepo.ListMediaByProductID(ctx, tx, productID)
		if err != nil {
			return err
		}
		for _, m := range mediaList {
			if m.ID == mediaID && m.Object != nil {
				objectBucket = m.Object.Bucket
				objectKey = m.Object.Key
				break
			}
		}

		if err := ps.productRepo.RemoveMedia(ctx, tx, productID, mediaID); err != nil {
			return err
		}

		if objectKey != "" {
			// Instead of marking it deleting here in tx, we just do nothing in tx,
			// and let objectService.DeleteObject handle marking and deleting outside tx.
		}
		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Product media relationship not found").
				WithCode(errcode.CodeMediaNotFound).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to detach media").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	if objectKey != "" {
		_ = ps.objectService.DeleteObject(ctx, objectBucket, objectKey)
	}

	return nil
}

func (ps *ProductService) ReorderMedia(
	ctx context.Context,
	productID uuid.UUID,
	orderedMediaIDs []uuid.UUID,
) error {
	if productID == uuid.Nil {
		return apierr.ErrBadRequest("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if len(orderedMediaIDs) == 0 {
		return nil
	}

	err := ps.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return ps.productRepo.ReorderMedia(ctx, tx, productID, orderedMediaIDs)
	})

	if err != nil {
		return apierr.ErrInternalError("Failed to reorder product media").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}

