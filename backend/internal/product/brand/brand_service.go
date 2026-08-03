package brand

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type BrandService struct {
	brepo     *BrandRepository
	qexecuter database.Runner
}

func NewService(
	r database.Runner,
	br *BrandRepository,
) *BrandService {
	return &BrandService{
		qexecuter: r,
		brepo:     br,
	}
}

type CreateBrandInput struct {
	Name string  `json:"name"`
	Link *string `json:"link,omitempty"`
}

type UpdateBrandInput struct {
	Name                *string    `json:"name,omitempty"`
	Link                *string    `json:"link,omitempty"`
	LogoStorageObjectID *uuid.UUID `json:"logo_storage_object_id,omitempty"`
}

// CreateBrand validates inputs and inserts a new product brand.
func (bs *BrandService) CreateBrand(
	ctx context.Context,
	in *CreateBrandInput,
) (*model.ProductBrand, error) {
	if in == nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_PAYLOAD").
			WithMessage("Request body is required")
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("VALIDATION_FAILED").
			WithMessage("Validation error").
			WithFields(api.FieldError{
				Field:   "name",
				Message: "brand name cannot be empty",
			})
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	brand := &model.ProductBrand{
		ID:        uuid.New(),
		Name:      name,
		Link:      in.Link,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brepo.Create(ctx, db, brand)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, security.NewSecureError(http.StatusConflict).
				WithCode("BRAND_ALREADY_EXISTS").
				WithMessage("A brand with this name already exists").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, security.NewSecureError(http.StatusBadRequest).
				WithCode("LOGO_STORAGE_OBJECT_NOT_FOUND").
				WithMessage("Referenced logo storage object does not exist").
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "logo_storage_object_id",
					Message: "storage object reference is invalid",
				})

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to create brand").
				Wrap(err).
				WithStack()
		}
	}

	return brand, nil
}

// GetBrandByID retrieves an active brand by its UUID.
func (bs *BrandService) GetBrandByID(
	ctx context.Context,
	brandID uuid.UUID,
) (*model.ProductBrand, error) {
	if brandID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Brand ID is required")
	}

	var brand *model.ProductBrand
	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		brand, repoErr = bs.brepo.GetByID(ctx, db, brandID)
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, security.NewSecureError(http.StatusNotFound).
				WithCode("BRAND_NOT_FOUND").
				WithMessage("Brand not found").
				Wrap(err)

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to fetch brand").
				Wrap(err).
				WithStack()
		}
	}

	return brand, nil
}

// UpdateBrand updates specific fields on an existing active brand.
func (bs *BrandService) UpdateBrand(
	ctx context.Context,
	brandID uuid.UUID,
	in UpdateBrandInput,
) error {
	if brandID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Brand ID is required")
	}

	fields := UpdateBrandFields{
		Link:                in.Link,
		LogoStorageObjectID: in.LogoStorageObjectID,
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return security.NewSecureError(http.StatusBadRequest).
				WithCode("VALIDATION_FAILED").
				WithMessage("Validation error").
				WithFields(api.FieldError{
					Field:   "name",
					Message: "brand name cannot be empty",
				})
		}
		fields.Name = &name
	}

	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brepo.Update(ctx, db, brandID, fields)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("BRAND_NOT_FOUND").
				WithMessage("Brand not found").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrConflict):
			return security.NewSecureError(http.StatusConflict).
				WithCode("BRAND_ALREADY_EXISTS").
				WithMessage("A brand with this name already exists").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return security.NewSecureError(http.StatusBadRequest).
				WithCode("LOGO_STORAGE_OBJECT_NOT_FOUND").
				WithMessage("Referenced logo storage object does not exist").
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "logo_storage_object_id",
					Message: "storage object reference is invalid",
				})

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to update brand").
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// ListCategories handles public and admin brand listing pipeline.
func (bs *BrandService) ListBrands(
	ctx context.Context,
	q *api.ListQuery,
	includeDeleted bool,
) (*api.PagedResult[model.ProductBrand], error) {
	if q == nil {
		q = &api.ListQuery{}
	}

	var result *api.PagedResult[model.ProductBrand]
	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = bs.brepo.List(ctx, db, ListBrandOptions{
			Query:          q,
			IncludeDeleted: includeDeleted,
		})
		return repoErr
	})

	if err != nil {
		return nil, security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to list brands").
			Wrap(err).
			WithStack()
	}

	return result, nil
}

func (bs *BrandService) List(
	ctx context.Context,
	q *api.ListQuery,
) (*api.PagedResult[model.ProductBrand], error) {
	return bs.ListBrands(ctx, q, false)
}

func (bs *BrandService) AdminList(
	ctx context.Context,
	q *api.ListQuery,
) (*api.PagedResult[model.ProductBrand], error) {
	return bs.ListBrands(ctx, q, true)
}

// DeleteBrandByID performs a soft-delete on an active brand.
func (bs *BrandService) DeleteBrandByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	if id == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Brand ID is required")
	}

	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brepo.Delete(ctx, db, id)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("BRAND_NOT_FOUND").
				WithMessage("Brand not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to delete brand").
				Wrap(err).
				WithStack()
		}
	}

	return nil
}
