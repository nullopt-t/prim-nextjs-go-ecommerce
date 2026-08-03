package product

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/object"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/brand"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/category"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/tag"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/variant"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
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

func (s *ProductService) CreateProductAsDraft(
	ctx context.Context,
	input CreateProductInput,
) (*model.Product, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, apierr.New(http.StatusBadRequest, "Title is required").
			WithCode(apierr.CodeValidationFailed)
	}

	if input.CategoryID == uuid.Nil {
		return nil, apierr.New(http.StatusBadRequest, "CategoryID is required").
			WithCode(apierr.CodeValidationFailed)
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
			return nil, apierr.New(http.StatusConflict, "Product with this public_id already exists").
				WithCode(errcode.CodeProductAlreadyExists).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.New(http.StatusBadRequest, "Invalid brand or category reference").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)

		default:
			return nil, apierr.New(http.StatusInternalServerError, "Failed to create product").
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
	if productID == uuid.Nil {
		return apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		product, err := s.productRepo.Get(ctx, tx, Filter{ID: &productID})
		if err != nil {
			return err
		}

		if input.Title != nil {
			title := strings.TrimSpace(*input.Title)
			if title == "" {
				return apierr.New(http.StatusBadRequest, "Title cannot be empty").
					WithCode(apierr.CodeValidationFailed)
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
			return apierr.New(http.StatusNotFound, "Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return apierr.New(http.StatusBadRequest, "Referenced brand or category does not exist").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)

		default:
			return apierr.New(http.StatusInternalServerError, "Failed to update product").
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
	if productID == uuid.Nil {
		return nil, apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
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
			return nil, apierr.New(http.StatusNotFound, "Product not found").
				WithCode(errcode.CodeProductNotFound)

		default:
			return nil, apierr.New(http.StatusInternalServerError, "Failed to fetch product").
				WithCode(apierr.CodeInternalError).
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
		return nil, apierr.New(http.StatusBadRequest, "Public ID is required").
			WithCode(apierr.CodeInvalidInput)
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
			return nil, apierr.New(http.StatusNotFound, "Product not found").
				WithCode(errcode.CodeProductNotFound)

		default:
			return nil, apierr.New(http.StatusInternalServerError, "Failed to fetch product").
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
		return nil, apierr.New(http.StatusInternalServerError, "Failed to list products").
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
	if productID == uuid.Nil {
		return apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	variants, err := s.variantService.ListVariantsByProductID(ctx, productID, nil, false)
	if err != nil {
		return err
	}
	if len(variants.Items) == 0 {
		return apierr.New(http.StatusBadRequest, "Product must have at least one variant before publishing").
			WithCode(errcode.CodePublishFailed)
	}

	err = s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.UpdateStatus(ctx, db, productID, model.PublicationStatusPublished)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.New(http.StatusNotFound, "Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.New(http.StatusInternalServerError, "Failed to publish product").
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
	if productID == uuid.Nil {
		return apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.UpdateStatus(ctx, db, productID, model.PublicationStatusArchived)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.New(http.StatusNotFound, "Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.New(http.StatusInternalServerError, "Failed to archive product").
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
	if productID == uuid.Nil {
		return apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.productRepo.SoftDelete(ctx, db, Filter{ID: &productID})
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.New(http.StatusNotFound, "Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)

		default:
			return apierr.New(http.StatusInternalServerError, "Failed to soft-delete product").
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
	q *api.ListQuery,
) (*api.PagedResult[model.ProductVariant], error) {
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
		return nil, apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, apierr.New(http.StatusBadRequest, "Failed to open uploaded file").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}
	defer file.Close()

	const maxFileSize = 10 * 1024 * 1024
	if fileHeader.Size > maxFileSize {
		return nil, apierr.New(http.StatusBadRequest, "File size exceeds 10MB limit").
			WithCode(apierr.CodeFileTooLarge)
	}

	detectedContentType, err := fileUtil.DetectContentType(file)
	if err != nil {
		return nil, apierr.New(http.StatusBadRequest, "Failed to detect file content type").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, apierr.New(http.StatusBadRequest, "Failed to reset file reader").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	mediaType, err := model.ParseMediaType(detectedContentType)
	if err != nil {
		return nil, apierr.New(http.StatusBadRequest, "Unsupported media format").
			WithCode(apierr.CodeUnsupportedMediaType).
			Wrap(err)
	}

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
		return nil, apierr.New(http.StatusInternalServerError, "Failed to upload file to MinIO").
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
			SortOrder:       maxOrder + 1,
		})
		return repoErr
	})

	if err != nil {
		_ = ps.minioClient.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
		return nil, apierr.New(http.StatusInternalServerError, "Failed to attach media to product").
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
		return nil, apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var mediaList []*model.ProductMedia
	err := ps.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		mediaList, err = ps.productRepo.ListMediaByProductID(ctx, db, productID)
		return err
	})

	if err != nil {
		return nil, apierr.New(http.StatusInternalServerError, "Failed to fetch product media").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

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
		return apierr.New(http.StatusBadRequest, "Product ID and Media ID are required").
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
			if err := ps.objectService.MarkDeletingByKey(ctx, tx, objectBucket, objectKey); err != nil {
				ps.logger.Warn("failed to mark storage object for deletion", log.Meta{"key": objectKey, "error": err})
			}
		}
		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.New(http.StatusNotFound, "Product media relationship not found").
				WithCode(errcode.CodeMediaNotFound).
				Wrap(err)

		default:
			return apierr.New(http.StatusInternalServerError, "Failed to detach media").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	if objectKey != "" {
		_ = ps.minioClient.RemoveObject(ctx, objectBucket, objectKey, minio.RemoveObjectOptions{})
	}

	return nil
}

func (ps *ProductService) ReorderMedia(
	ctx context.Context,
	productID uuid.UUID,
	orderedMediaIDs []uuid.UUID,
) error {
	if productID == uuid.Nil {
		return apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if len(orderedMediaIDs) == 0 {
		return nil
	}

	err := ps.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return ps.productRepo.ReorderMedia(ctx, tx, productID, orderedMediaIDs)
	})

	if err != nil {
		return apierr.New(http.StatusInternalServerError, "Failed to reorder product media").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}
