package brand

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
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

func (bs *BrandService) CreateBrand(
	ctx context.Context,
	in *CreateBrandInput,
) (*model.ProductBrand, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	brand := &model.ProductBrand{
		ID:        uuid.New(),
		PublicID:  uuid.NewString(),
		Name:      in.Name,
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
			return nil, apierr.ErrConflict("A brand with this name already exists").
				WithCode(errcode.CodeBrandAlreadyExists).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.ErrBadRequest("Referenced logo storage object does not exist").
				WithCode(errcode.CodeLogoStorageObjectNotFound).
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "logo_storage_object_id",
					Message: "storage object reference is invalid",
				})

		default:
			return nil, apierr.ErrInternalError("Failed to create brand").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return brand, nil
}

func (bs *BrandService) GetBrandByID(
	ctx context.Context,
	brandID uuid.UUID,
) (*model.ProductBrand, error) {
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
			return nil, apierr.ErrNotFound("Brand not found").
				WithCode(errcode.CodeBrandNotFound).
				Wrap(err)

		default:
			return nil, apierr.ErrInternalError("Failed to fetch brand").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return brand, nil
}

func (bs *BrandService) UpdateBrand(
	ctx context.Context,
	brandID uuid.UUID,
	in UpdateBrandInput,
) error {
	fields := UpdateBrandFields{
		Link:                in.Link,
		LogoStorageObjectID: in.LogoStorageObjectID,
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

	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brepo.Update(ctx, db, brandID, fields)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Brand not found").
				WithCode(errcode.CodeBrandNotFound).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrConflict):
			return apierr.ErrConflict("A brand with this name already exists").
				WithCode(errcode.CodeBrandAlreadyExists).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return apierr.ErrBadRequest("Referenced logo storage object does not exist").
				WithCode(errcode.CodeLogoStorageObjectNotFound).
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "logo_storage_object_id",
					Message: "storage object reference is invalid",
				})

		default:
			return apierr.ErrInternalError("Failed to update brand").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (bs *BrandService) ListBrands(
	ctx context.Context,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[model.ProductBrand], error) {
	if q == nil {
		q = &pagination.ListQuery{}
	}

	var result *pagination.PagedResult[model.ProductBrand]
	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = bs.brepo.List(ctx, db, ListBrandOptions{
			Query:          q,
			IncludeDeleted: includeDeleted,
		})
		return repoErr
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to list brands").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return result, nil
}

func (bs *BrandService) List(
	ctx context.Context,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductBrand], error) {
	return bs.ListBrands(ctx, q, false)
}

func (bs *BrandService) AdminList(
	ctx context.Context,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductBrand], error) {
	return bs.ListBrands(ctx, q, true)
}

func (bs *BrandService) DeleteBrandByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	if id == uuid.Nil {
		return apierr.ErrBadRequest("Brand ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return bs.brepo.Delete(ctx, db, id)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Brand not found").
				WithCode(errcode.CodeBrandNotFound).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to delete brand").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}
