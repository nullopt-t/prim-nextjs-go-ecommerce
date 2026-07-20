package brand

import (
	"context"
	"errors"
	"net/http"
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
	Name    string
	Slug    string
	LogoURL string
	LogoAlt string
}

func (bs *BrandService) CreateBrand(
	ctx context.Context,
	in *CreateBrandInput,
) (*model.ProductBrand, error) {

	now := time.Now()
	brand := &model.ProductBrand{
		ID:        uuid.New(),
		Name:      in.Name,
		Slug:      in.Slug,
		LogoURL:   in.LogoURL,
		LogoAlt:   in.LogoAlt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := bs.qexecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return bs.brepo.Create(
				ctx,
				db,
				brand,
			)
		},
	)
	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(
			mappedError,
			database.ErrConflict,
		):
			return nil, security.NewSecureError(
				http.StatusConflict,
				security.CodeConflict,
				"resourse alread existed",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"internal server error",
				err,
			)
		}
	}

	return brand, nil
}

func (bs *BrandService) GetBrandByID(
	ctx context.Context,
	brandID uuid.UUID,
) (*model.ProductBrand, error) {
	var brand *model.ProductBrand
	err := bs.qexecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			b, err := bs.brepo.Get(
				ctx,
				db,
				&Filter{
					ID: &brandID,
				},
			)
			if err != nil {
				return err
			}
			brand = b
			return nil
		},
	)
	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(
			mappedError,
			database.ErrNotFound,
		):
			return nil, security.NewSecureError(
				http.StatusNotFound,
				security.CodeNotFound,
				"resource not found",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"internal server error",
				err,
			)
		}
	}
	return brand, nil
}

func (bs *BrandService) GetBrandBySlug(
	ctx context.Context,
	slug string,
) (*model.ProductBrand, error) {
	var brand *model.ProductBrand
	err := bs.qexecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			b, err := bs.brepo.Get(
				ctx,
				db,
				&Filter{
					Slug: &slug,
				},
			)
			if err != nil {
				return err
			}
			brand = b
			return nil
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch the brand",
			err,
		)
	}
	return brand, nil
}

type UpdateBrandInput struct {
	Name    *string
	LogoURL *string
	LogoAlt *string
}

func (bs *BrandService) UpdateBrand(
	ctx context.Context,
	brandID uuid.UUID,
	input *UpdateBrandInput,
) error {
	fields := &UpdateBrandFields{
		Name:    input.Name,
		LogoURL: input.LogoURL,
		LogoAlt: input.LogoAlt,
	}
	err := bs.qexecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return bs.brepo.Update(
				ctx,
				db,
				brandID,
				fields,
			)
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to update the brand",
			err,
		)
	}
	return nil
}

func (bs *BrandService) List(
	ctx context.Context,
	q *api.ListQuery,
) ([]*model.ProductBrand, *api.Page, error) {
	var res []*model.ProductBrand
	var page *api.Page
	err := bs.qexecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			brands, p, err := bs.brepo.List(
				ctx,
				db,
				q,
			)
			if err != nil {
				return err
			}
			res = brands
			page = p
			return nil
		})
	if err != nil {
		return nil, nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch the categories",
			err,
		)
	}
	return res, page, nil
}

func (bs *BrandService) DeleteBrandByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := bs.qexecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			return bs.brepo.Delete(
				ctx,
				db,
				&Filter{
					ID: &id,
				},
			)
		})
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to delete the brand",
			err,
		)
	}
	return nil
}

func (bs *BrandService) DeleteBrandBySlug(
	ctx context.Context,
	slug string,
) error {
	err := bs.qexecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			return bs.brepo.Delete(
				ctx,
				db,
				&Filter{
					Slug: &slug,
				},
			)
		})
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to delete the brand",
			err,
		)
	}
	return nil
}
