package brand

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	fileUtil "github.com/nullopt-t/prim-backend/pkg/file"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

type ObjectRepository interface {
	Create(ctx context.Context, qe database.QueryExecutor, object *model.Object) error
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Object, error)
	GetByIDs(ctx context.Context, qe database.QueryExecutor, ids []uuid.UUID) ([]*model.Object, error)
	MarkDeletingByKey(ctx context.Context, qe database.QueryExecutor, bucket, key string) error
	MarkDeletedByKey(ctx context.Context, qe database.QueryExecutor, bucket, key string) error
}

type BrandService struct {
	runner          database.Runner
	brandRepo       *BrandRepository
	objectRepo      ObjectRepository
	storageProvider storage.StorageProvider
}

func NewService(
	runner database.Runner,
	brandRepo *BrandRepository,
	objectRepo ObjectRepository,
	storageProvider storage.StorageProvider,
) *BrandService {
	return &BrandService{
		runner:          runner,
		brandRepo:       brandRepo,
		objectRepo:      objectRepo,
		storageProvider: storageProvider,
	}
}

// handleError maps a repository error to a standard API error.
func (bs *BrandService) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError("Brand not found").
			WithCode(errcode.CodeBrandNotFound)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("A brand with this name already exists").
			WithCode(errcode.CodeBrandAlreadyExists)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced logo storage object does not exist").
			WithCode(errcode.CodeLogoStorageObjectNotFound).
			WithFields(api.FieldError{
				Field:   "logo_object_id",
				Message: "storage object reference is invalid",
			})
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

type CreateBrandInput struct {
	Name string  `json:"name"`
	Link *string `json:"link,omitempty"`
}

type UpdateBrandInput struct {
	Name         *string    `json:"name,omitempty"`
	Link         *string    `json:"link,omitempty"`
	LogoObjectID *uuid.UUID `json:"logoObjectId,omitempty"`
}

func (bs *BrandService) CreateBrand(
	ctx context.Context,
	in *CreateBrandInput,
) (*model.ProductBrand, error) {
	if in == nil {
		panic("CreateBrand: input must not be nil")
	}

	brand := &model.ProductBrand{
		ID:   uuid.New(),
		Name: in.Name,
		Link: in.Link,
	}

	err := bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brandRepo.Create(ctx, db, brand)
	})

	if err != nil {
		return nil, bs.handleError(err, "Failed to create brand")
	}

	return brand, nil
}

func (bs *BrandService) GetBrandByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.ProductBrand, error) {
	var brand *model.ProductBrand
	err := bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		brand, repoErr = bs.brandRepo.GetByID(ctx, db, id)
		if repoErr != nil {
			return repoErr
		}
		return bs.populateLogoURLs(ctx, db, []*model.ProductBrand{brand})
	})

	if err != nil {
		return nil, bs.handleError(err, "Failed to fetch brand")
	}

	return brand, nil
}

func (bs *BrandService) GetBrandByIDWithTx(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.ProductBrand, error) {
	brand, repoErr := bs.brandRepo.GetByID(ctx, qe, id)
	if repoErr != nil {
		return nil, bs.handleError(repoErr, "Failed to fetch brand")
	}
	if err := bs.populateLogoURLs(ctx, qe, []*model.ProductBrand{brand}); err != nil {
		return nil, bs.handleError(err, "Failed to populate brand logo")
	}
	return brand, nil
}

func (bs *BrandService) UpdateBrand(
	ctx context.Context,
	brandID uuid.UUID,
	in UpdateBrandInput,
) error {
	fields := UpdateBrandFields{
		Name:         in.Name,
		Link:         in.Link,
		LogoObjectID: in.LogoObjectID,
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return apierr.BadRequestError("Validation error").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{
					Field:   "name",
					Message: "brand name cannot be empty",
				})
		}
		fields.Name = &name
	}

	err := bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brandRepo.Update(ctx, db, brandID, fields)
	})

	if err != nil {
		return bs.handleError(err, "Failed to update brand")
	}

	return nil
}

type listBrandsOptions struct {
	query          *pagination.ListQuery
	includeDeleted bool
	populateURL    bool
}

func (bs *BrandService) listBrands(
	ctx context.Context,
	opts listBrandsOptions,
) (*pagination.PagedResult[model.ProductBrand], error) {
	q := opts.query
	if q == nil {
		q = &pagination.ListQuery{}
	}

	cursor, err := q.Keyset()
	if err != nil {
		return nil, apierr.BadRequestError("Invalid pagination cursor").Wrap(err)
	}

	var result *pagination.PagedResult[model.ProductBrand]
	err = bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = bs.brandRepo.List(ctx, db, ListBrandOptions{
			Query:          q,
			IncludeDeleted: opts.includeDeleted,
			Cursor:         cursor,
		})
		if repoErr != nil {
			return repoErr
		}
		if opts.populateURL {
			return bs.populateLogoURLs(ctx, db, result.Items)
		}
		return nil
	})

	if err != nil {
		return nil, bs.handleError(err, "Failed to list brands")
	}

	return result, nil
}

func (bs *BrandService) List(
	ctx context.Context,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductBrand], error) {
	return bs.listBrands(ctx, listBrandsOptions{
		query:          q,
		includeDeleted: false,
		populateURL:    true,
	})
}

func (bs *BrandService) AdminList(
	ctx context.Context,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductBrand], error) {
	return bs.listBrands(ctx, listBrandsOptions{
		query:          q,
		includeDeleted: true,
		populateURL:    true,
	})
}

func (bs *BrandService) DeleteBrandByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	if id == uuid.Nil {
		return apierr.BadRequestError("Brand ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	// Execute within a transaction so brand check and soft-delete are atomic
	err := bs.runner.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		return bs.brandRepo.DeleteByID(ctx, tx, id)
	})

	if err != nil {
		if errors.Is(err, database.ErrConflict) {
			return apierr.ConflictError("Brand is currently in use and cannot be deleted").
				WithCode(errcode.CodeBrandInUse)
		}
		return bs.handleError(err, "Failed to delete brand")
	}

	return nil
}

func (bs *BrandService) UploadBrandLogo(
	ctx context.Context,
	brandID uuid.UUID,
	file *multipart.FileHeader,
) (*string, error) {
	if brandID == uuid.Nil {
		return nil, apierr.BadRequestError("Brand ID is required").WithCode(apierr.CodeInvalidInput)
	}
	if file == nil {
		return nil, apierr.BadRequestError("File is required").WithCode(apierr.CodeInvalidInput)
	}

	// Brand logos are raster images only; cap size at 2MB.
	const maxBrandLogoSize = 2 * 1024 * 1024
	if file.Size > maxBrandLogoSize {
		return nil, apierr.BadRequestError("File size exceeds 2MB limit").
			WithCode(apierr.CodeFileTooLarge)
	}

	src, err := file.Open()
	if err != nil {
		return nil, apierr.InternalServerError("Failed to open file").Wrap(err)
	}
	defer func() { _ = src.Close() }()

	contentType, err := fileUtil.DetectContentType(src)
	if err != nil {
		return nil, apierr.InternalServerError("Failed to detect content type").Wrap(err)
	}
	if !fileUtil.IsAllowedImageMime(contentType) {
		return nil, apierr.BadRequestError(
			"Only JPEG, PNG, WebP, and AVIF images are allowed",
		).WithCode(apierr.CodeUnsupportedMediaType)
	}

	if bs.storageProvider == nil || bs.objectRepo == nil {
		return nil, apierr.InternalServerError("Storage provider or repository unavailable")
	}

	key := uuid.New().String()
	const bucket = "catalog"

	if err := bs.storageProvider.Upload(ctx, bucket, key, src, file.Size, contentType); err != nil {
		return nil, apierr.InternalServerError("Failed to upload logo").Wrap(err)
	}

	newObj := &model.Object{
		ID:          uuid.New(),
		Bucket:      bucket,
		Key:         key,
		ContentType: contentType,
		FileSize:    file.Size,
		Status:      model.ObjectStatusUploaded,
	}

	var oldLogoObj *model.Object

	txErr := bs.runner.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		brand, bErr := bs.brandRepo.GetByID(ctx, tx, brandID)
		if bErr != nil {
			return bErr
		}

		if brand.LogoObjectID != nil {
			oldLogoObj, _ = bs.objectRepo.GetByID(ctx, tx, *brand.LogoObjectID)
		}

		if err := bs.objectRepo.Create(ctx, tx, newObj); err != nil {
			return err
		}

		if err := bs.brandRepo.Update(ctx, tx, brandID, UpdateBrandFields{
			LogoObjectID: &newObj.ID,
		}); err != nil {
			return err
		}

		if oldLogoObj != nil {
			_ = bs.objectRepo.MarkDeletingByKey(ctx, tx, oldLogoObj.Bucket, oldLogoObj.Key)
		}

		return nil
	})

	if txErr != nil {
		_ = bs.storageProvider.Delete(ctx, bucket, key)
		return nil, bs.handleError(txErr, "Failed to update brand logo")
	}

	if oldLogoObj != nil {
		if delErr := bs.storageProvider.Delete(ctx, oldLogoObj.Bucket, oldLogoObj.Key); delErr == nil {
			_ = bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
				return bs.objectRepo.MarkDeletedByKey(ctx, db, oldLogoObj.Bucket, oldLogoObj.Key)
			})
		}
	}

	url := bs.storageProvider.GetPublicURL(bucket, key)
	return &url, nil
}

func (bs *BrandService) populateLogoURLs(
	ctx context.Context,
	qe database.QueryExecutor,
	brands []*model.ProductBrand,
) error {
	var objectIDs []uuid.UUID
	for _, brand := range brands {
		if brand.LogoObjectID != nil {
			objectIDs = append(objectIDs, *brand.LogoObjectID)
		}
	}
	if len(objectIDs) == 0 || bs.objectRepo == nil {
		return nil
	}

	objects, err := bs.objectRepo.GetByIDs(ctx, qe, objectIDs)
	if err != nil {
		return err
	}

	objMap := make(map[uuid.UUID]*model.Object, len(objects))
	for _, obj := range objects {
		objMap[obj.ID] = obj
	}

	for _, brand := range brands {
		if brand.LogoObjectID != nil {
			if obj, ok := objMap[*brand.LogoObjectID]; ok {
				if bs.storageProvider != nil {
					url := bs.storageProvider.GetPublicURL(obj.Bucket, obj.Key)
					brand.LogoURL = &url
				}
			}
		}
	}
	return nil
}

