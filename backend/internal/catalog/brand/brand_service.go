package brand

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	fileUtil "github.com/m-mahmoud-alsaid/prim-backend/pkg/file"
)

type ObjectService interface {
	UploadObject(ctx context.Context, contentType string, size int64, bucket string, file io.Reader) (*model.Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	DeleteObjectByID(ctx context.Context, id uuid.UUID) error
	GetObjectByID(ctx context.Context, id uuid.UUID) (*model.Object, error)
	GetObjectsByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Object, error)
	GetObjectURL(ctx context.Context, bucket, key string) string
}

type BrandService struct {
	brandRepo     *BrandRepository
	runner        database.Runner
	objectService ObjectService
}

func NewService(
	runner database.Runner,
	brandRepo *BrandRepository,
	objectService ObjectService,
) *BrandService {
	return &BrandService{
		runner:        runner,
		brandRepo:     brandRepo,
		objectService: objectService,
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
		return apierr.ErrNotFound("Brand not found").
			WithCode(errcode.CodeBrandNotFound)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ErrConflict("A brand with this name already exists").
			WithCode(errcode.CodeBrandAlreadyExists)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.ErrBadRequest("Referenced logo storage object does not exist").
			WithCode(errcode.CodeLogoStorageObjectNotFound).
			WithFields(api.FieldError{
				Field:   "logo_object_id",
				Message: "storage object reference is invalid",
			})
	default:
		return apierr.ErrInternalError(fallbackMsg).
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
	LogoObjectID *uuid.UUID `json:"logo_object_id,omitempty"`
}

func (bs *BrandService) CreateBrand(
	ctx context.Context,
	in *CreateBrandInput,
) (*model.ProductBrand, error) {
	if in == nil {
		panic("CreateBrand: input must not be nil")
	}

	brand := &model.ProductBrand{
		ID:       uuid.New(),
		PublicID: uuid.NewString(),
		Name:     in.Name,
		Link:     in.Link,
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
		return repoErr
	})

	if err != nil {
		return nil, bs.handleError(err, "Failed to fetch brand")
	}

	return brand, nil
}

func (bs *BrandService) UpdateBrand(
	ctx context.Context,
	brandID uuid.UUID,
	in UpdateBrandInput,
) error {
	fields := UpdateBrandFields{
		Name: in.Name,
		Link: in.Link,
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return apierr.ErrBadRequest("Validation error").
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

	var result *pagination.PagedResult[model.ProductBrand]
	err := bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = bs.brandRepo.List(ctx, db, ListBrandOptions{
			Query:          q,
			IncludeDeleted: opts.includeDeleted,
		})
		if repoErr != nil {
			return repoErr
		}
		if opts.populateURL {
			return bs.populateLogoURLs(ctx, result.Items)
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
		populateURL:    false,
	})
}

func (bs *BrandService) DeleteBrandByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	if id == uuid.Nil {
		return apierr.ErrBadRequest("Brand ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brandRepo.DeleteByID(ctx, db, id)
	})

	if err != nil {
		if errors.Is(err, database.ErrConflict) {
			return apierr.ErrConflict("Brand is currently in use and cannot be deleted").
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
		return nil, apierr.ErrBadRequest("Brand ID is required").WithCode(apierr.CodeInvalidInput)
	}
	if file == nil {
		return nil, apierr.ErrBadRequest("File is required").WithCode(apierr.CodeInvalidInput)
	}

	src, err := file.Open()
	if err != nil {
		return nil, apierr.ErrInternalError("Failed to open file").Wrap(err)
	}
	defer func() { _ = src.Close() }()

	contentType, err := fileUtil.DetectContentType(src)
	if err != nil {
		return nil, apierr.ErrInternalError("Failed to detect content type").Wrap(err)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return nil, apierr.ErrBadRequest("Only image files are allowed").WithCode(apierr.CodeInvalidInput)
	}

	// Fetch existing brand to check for old logo
	var oldLogoID *uuid.UUID
	err = bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		brand, err := bs.brandRepo.GetByID(ctx, db, brandID)
		if err != nil {
			return err
		}
		oldLogoID = brand.LogoObjectID
		return nil
	})
	if err != nil {
		return nil, bs.handleError(err, "Failed to get brand")
	}

	obj, err := bs.objectService.UploadObject(ctx, contentType, file.Size, "catalog", src)
	if err != nil {
		return nil, apierr.ErrInternalError("Failed to upload logo").Wrap(err)
	}

	// Update the brand
	err = bs.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brandRepo.Update(ctx, db, brandID, UpdateBrandFields{
			LogoObjectID: &obj.ID,
		})
	})
	if err != nil {
		_ = bs.objectService.DeleteObjectByID(ctx, obj.ID) // cleanup
		return nil, bs.handleError(err, "Failed to update brand logo")
	}

	// Delete old logo if existed
	if oldLogoID != nil {
		_ = bs.objectService.DeleteObjectByID(ctx, *oldLogoID)
	}

	url := bs.objectService.GetObjectURL(ctx, obj.Bucket, obj.Key)
	return &url, nil
}

func (bs *BrandService) populateLogoURLs(ctx context.Context, brands []*model.ProductBrand) error {
	var objectIDs []uuid.UUID
	for _, brand := range brands {
		if brand.LogoObjectID != nil {
			objectIDs = append(objectIDs, *brand.LogoObjectID)
		}
	}
	if len(objectIDs) == 0 {
		return nil
	}

	objects, err := bs.objectService.GetObjectsByIDs(ctx, objectIDs)
	if err != nil {
		return err
	}

	objMap := make(map[uuid.UUID]*model.Object)
	for _, obj := range objects {
		objMap[obj.ID] = obj
	}

	for _, brand := range brands {
		if brand.LogoObjectID != nil {
			if obj, ok := objMap[*brand.LogoObjectID]; ok {
				url := bs.objectService.GetObjectURL(ctx, obj.Bucket, obj.Key)
				brand.LogoURL = &url
			}
		}
	}
	return nil
}
