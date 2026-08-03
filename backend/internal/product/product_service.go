package product

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/object"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/brand"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/category"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/tag"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/variant"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	fileUtil "github.com/m-mahmoud-alsaid/prim-backend/pkg/file"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/minio/minio-go/v7"
)

type ProductService struct {
	dr              database.Runner
	logger          log.Logger
	minioClient     *minio.Client
	objectService   *object.ObjectService
	productRepo     *ProductRepository
	brandService    *brand.BrandService
	categoryService *category.CategoryService
	tagService      *tag.TagService
	variantService  *variant.VariantService
	minCfg          *config.MinioConfig
}

func NewService(
	r database.Runner,
	logger log.Logger,
	minioClient *minio.Client,
	productRepo *ProductRepository,
	objectService *object.ObjectService,
	brandService *brand.BrandService,
	categoryService *category.CategoryService,
	tagService *tag.TagService,
	variantService *variant.VariantService,
	minCfg *config.MinioConfig,
) *ProductService {
	return &ProductService{
		dr:              r,
		minCfg:          minCfg,
		logger:          logger,
		minioClient:     minioClient,
		objectService:   objectService,
		productRepo:     productRepo,
		brandService:    brandService,
		categoryService: categoryService,
		tagService:      tagService,
		variantService:  variantService,
	}
}

// --- DTO Inputs ---

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
	Product *model.Product      `json:"product"`
	Brand   *model.ProductBrand `json:"brand,omitempty"`
}

// --- Product Core Operations ---

func (s *ProductService) CreateProductAsDraft(
	ctx context.Context,
	input CreateProductInput,
) (*model.Product, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("VALIDATION_FAILED").
			WithMessage("Title is required")
	}

	if input.CategoryID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("VALIDATION_FAILED").
			WithMessage("CategoryID is required")
	}

	product := &model.Product{
		ID:          uuid.New(),
		BrandID:     input.BrandID,
		CategoryID:  input.CategoryID,
		PublicID:    uuid.NewString(),
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Highlights:  input.Highlights,
		Status:      model.PublicationStatusDraft,
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.Create(ctx, db, product)
	})
	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, security.NewSecureError(http.StatusConflict).
				WithCode("PRODUCT_ALREADY_EXISTS").
				WithMessage("Product with this public_id already exists").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, security.NewSecureError(http.StatusBadRequest).
				WithCode("INVALID_REFERENCE").
				WithMessage("Invalid brand or category reference").
				Wrap(err)

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to create product").
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
	if productID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		product, err := s.productRepo.Get(ctx, tx, Filter{ID: &productID})
		if err != nil {
			return err
		}

		if input.Title != nil {
			title := strings.TrimSpace(*input.Title)
			if title == "" {
				return security.NewSecureError(http.StatusBadRequest).
					WithCode("VALIDATION_FAILED").
					WithMessage("Title cannot be empty")
			}
			product.Title = title
		}

		if input.Description != nil {
			product.Description = strings.TrimSpace(*input.Description)
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
			return security.NewSecureError(http.StatusNotFound).
				WithCode("PRODUCT_NOT_FOUND").
				WithMessage("Product not found").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return security.NewSecureError(http.StatusBadRequest).
				WithCode("INVALID_REFERENCE").
				WithMessage("Referenced brand or category does not exist").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to update product").
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
	if productID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	var product *model.Product
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		product, err = s.productRepo.Get(ctx, db, Filter{ID: &productID})
		return err
	})

	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(mappedError, database.ErrNotFound):
			return nil, security.NewSecureError(http.StatusNotFound).
				WithCode("PRODUCT_NOT_FOUND").
				WithMessage("Product not found")

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to fetch product").
				Wrap(err).
				WithStack()
		}
	}

	return product, nil
}

func (s *ProductService) GetByPID(
	ctx context.Context,
	ppid string,
) (*ProductDetails, error) {
	cleanPID := strings.TrimSpace(ppid)
	if cleanPID == "" {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Public ID is required")
	}

	productDetails := &ProductDetails{}
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		prod, err := s.productRepo.Get(ctx, db, Filter{PublicID: &cleanPID})
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
			return nil, security.NewSecureError(http.StatusNotFound).
				WithCode("PRODUCT_NOT_FOUND").
				WithMessage("Product not found")

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to fetch product").
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

	return productDetails, nil
}

func (s *ProductService) List(
	ctx context.Context,
	q *api.ListQuery,
	includeDeleted bool,
) (*api.PagedResult[ProductListItem], error) {
	if q == nil {
		q = &api.ListQuery{}
	}

	var res *api.PagedResult[ProductListItem]
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		res, err = s.productRepo.List(ctx, db, q, includeDeleted)
		return err
	})

	if err != nil {
		return nil, security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to list products").
			Wrap(err).
			WithStack()
	}

	return res, nil
}

func (s *ProductService) PublishProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	if productID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.UpdateStatus(ctx, db, productID, model.PublicationStatusPublished)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("PRODUCT_NOT_FOUND").
				WithMessage("Product not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to publish product").
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
	if productID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.UpdateStatus(ctx, db, productID, model.PublicationStatusArchived)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("PRODUCT_NOT_FOUND").
				WithMessage("Product not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to archive product").
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
	if productID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.SoftDelete(ctx, db, Filter{ID: &productID})
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("PRODUCT_NOT_FOUND").
				WithMessage("Product not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to soft-delete product").
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// --- Cross-Domain Delegations ---

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
) (*api.PagedResult[model.ProductVariant], error) {
	return s.variantService.ListVariantsByProductID(ctx, productID, nil, false)
}

// --- Media Management ---

func (ps *ProductService) GenObjectKey(
	productID uuid.UUID,
	contentType string,
) string {
	return fmt.Sprintf(
		"products/%s/%s%s",
		productID,
		uuid.NewString(),
		fileUtil.MimeExtension(contentType),
	)
}

func (ps *ProductService) UploadProductMedia(
	ctx context.Context,
	productID uuid.UUID,
	fileHeader *multipart.FileHeader,
) (*model.ProductMedia, error) {
	if productID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_FILE").
			WithMessage("Failed to open uploaded file").
			Wrap(err)
	}
	defer file.Close()

	const maxFileSize = 10 * 1024 * 1024 // 10 MB limit
	if fileHeader.Size > maxFileSize {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("FILE_TOO_LARGE").
			WithMessage("File size exceeds 10MB limit")
	}

	detectedContentType, err := fileUtil.DetectContentType(file)
	if err != nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_FILE").
			WithMessage("Failed to detect file content type").
			Wrap(err)
	}

	mediaType, err := model.ParseMediaType(detectedContentType)
	if err != nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("UNSUPPORTED_MEDIA_TYPE").
			WithMessage("Unsupported media format").
			Wrap(err)
	}

	// Verify product exists
	product, err := ps.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	objectKey := ps.GenObjectKey(product.ID, detectedContentType)
	bucket := "product-media"

	cleanCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	info, err := ps.minioClient.PutObject(
		cleanCtx,
		bucket,
		objectKey,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{
			ContentType: detectedContentType,
		},
	)
	if err != nil {
		return nil, security.NewSecureError(http.StatusInternalServerError).
			WithCode("STORAGE_ERROR").
			WithMessage("Failed to upload file to MinIO").
			Wrap(err).
			WithStack()
	}

	var media *model.ProductMedia
	err = ps.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		obj, err := ps.objectService.CreateObjectWithTx(
			ctx,
			tx,
			info.Bucket,
			objectKey,
			info.Size,
			detectedContentType,
			model.ObjectStatusUploaded,
		)
		if err != nil {
			return err
		}

		var repoErr error
		media, repoErr = ps.productRepo.AddMedia(ctx, tx, CreateProductMediaInput{
			ProductID:       productID,
			StorageObjectID: obj.ID,
			MediaType:       mediaType,
			SortOrder:       0,
		})
		return repoErr
	})

	if err != nil {
		// Roll back MinIO upload if DB insert fails
		_ = ps.minioClient.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
		return nil, security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to attach media to product").
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
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	var mediaList []*model.ProductMedia
	err := ps.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		mediaList, err = ps.productRepo.ListMediaByProductID(ctx, db, productID)
		return err
	})

	if err != nil {
		return nil, security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to fetch product media").
			Wrap(err).
			WithStack()
	}

	// Resolve presigned URLs for media items
	for _, m := range mediaList {
		if m.Object != nil && m.Object.PublicURL == "" && m.Object.Bucket != "" && m.Object.Key != "" {
			presignedGET, err := ps.minioClient.PresignedGetObject(
				ctx,
				m.Object.Bucket,
				m.Object.Key,
				1*time.Hour,
				nil,
			)
			if err == nil {
				m.Object.PublicURL = presignedGET.String()
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
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID and Media ID are required")
	}

	err := ps.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return ps.productRepo.RemoveMedia(ctx, db, productID, mediaID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("MEDIA_NOT_FOUND").
				WithMessage("Product media relationship not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to detach media").
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (ps *ProductService) ReorderMedia(
	ctx context.Context,
	productID uuid.UUID,
	orderedMediaIDs []uuid.UUID,
) error {
	if productID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	if len(orderedMediaIDs) == 0 {
		return nil
	}

	err := ps.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return ps.productRepo.ReorderMedia(ctx, tx, productID, orderedMediaIDs)
	})

	if err != nil {
		return security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to reorder product media").
			Wrap(err).
			WithStack()
	}

	return nil
}
