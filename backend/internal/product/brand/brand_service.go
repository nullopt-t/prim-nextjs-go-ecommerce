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
	Name      string
	LogoURL   string
	LogoLabel string
}

func (bs *BrandService) CreateBrand(
	ctx context.Context,
	in CreateBrandInput,
) (*model.Brand, error) {
	brand := model.NewBrand(
		in.Name,
		in.LogoURL,
		in.LogoLabel,
	)
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
) (*model.Brand, error) {
	var brand *model.Brand
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

func (bs *BrandService) List(
	ctx context.Context,
	q *api.PageQuery,
) ([]*model.Brand, *api.Page, error) {
	var res []*model.Brand
	var page *api.Page
	err := bs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
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
		switch {
		default:
			return nil, nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to fetch the categories",
				err,
			)
		}
	}
	return res, page, nil
}
