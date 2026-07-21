package brand

import (
	"context"
	"errors"
	"net/http"

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
	LogoURL *string
	LogoAlt *string
}

func (bs *BrandService) CreateBrand(
	ctx context.Context,
	in *CreateBrandInput,
) (*model.ProductBrand, error) {

	brand := &model.ProductBrand{
		ID:      uuid.New(),
		Name:    in.Name,
		LogoURL: in.LogoURL,
		LogoAlt: in.LogoAlt,
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
				Filter{
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

type UpdateBrandInput struct {
	Name    *string
}

func (bs *BrandService) UpdateBrand(
	ctx context.Context,
	brandID uuid.UUID,
	input UpdateBrandInput,
) error {
	err := bs.qexecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return bs.brepo.Update(
				ctx,
				db,
				brandID,
				UpdateBrandFields{
					Name:    input.Name,
				},
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
) (*BrandList, error) {
	var brandList *BrandList
	err := bs.qexecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			brandList, err = bs.brepo.List(
				ctx,
				db,
				q,
			)
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch the categories",
			err,
		)
	}
	return brandList, nil
}


func (bs *BrandService) AdminList(
	ctx context.Context,
	q *api.ListQuery,
) (*BrandList, error) {
	var brandList *BrandList
	err := bs.qexecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			brandList, err = bs.brepo.AdminList(
				ctx,
				db,
				q,
			)
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch the categories",
			err,
		)
	}
	return brandList, nil
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
				Filter{
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
