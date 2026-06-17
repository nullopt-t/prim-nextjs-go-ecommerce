package product

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type ProductService struct {
	dbExecuter  database.Runner
	productRepo *ProductRepository
}

func NewService(r database.Runner,
	productRepo *ProductRepository,
) *ProductService {
	return &ProductService{
		dbExecuter:  r,
		productRepo: productRepo,
	}
}

func (s *ProductService) Create(
	ctx context.Context,
	req CreateProductRequest,
) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			p := model.NewProduct(
				req.Title,
				req.ShortDescription,
				req.Description,
				req.SKU,
				req.Slug,
				req.Status,
				req.Price,
				req.Currency,
			)

			cid, err := s.productRepo.Create(ctx, db, p)
			if err != nil {
				mappedErr := database.MapError(err)
				switch {
				case errors.Is(
					mappedErr,
					database.ErrConflict,
				):
					return security.NewSecureError(
						http.StatusConflict,
						security.CodeConflict,
						"product already exists",
						err,
					)
				default:
					return security.NewSecureError(
						http.StatusInternalServerError,
						security.CodeInternal,
						"failed to create a new product",
						err,
					)
				}
			}
			id = cid
			return nil
		})
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *ProductService) get(
	ctx context.Context,
	filter Filter,
) (*model.Product, error) {
	var p *model.Product
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			prod, err := s.productRepo.Get(ctx, db, filter)
			if err != nil {
				mappedError := database.MapError(err)
				switch {
				case errors.Is(
					mappedError,
					database.ErrNotFound,
				):
					return security.NewSecureError(
						http.StatusNotFound,
						security.CodeNotFound,
						"product not found",
						err,
					)
				default:
					return security.NewSecureError(
						http.StatusInternalServerError,
						security.CodeInternal,
						"failed to fetch a product",
						err,
					)

				}
			}
			p = prod
			return nil
		},
	)
	return p, err
}

func (s *ProductService) GetByID(
	ctx context.Context,
	productID uuid.UUID,
) (*model.Product, error) {
	return s.get(ctx, Filter{
		ID: productID,
	})
}

func (s *ProductService) GetAll(
	ctx context.Context,
	q api.PageQuery,
) ([]*model.Product, *api.Page, error) {
	var products []*model.Product
	var page *api.Page
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			products, page, err = s.productRepo.GetAll(ctx, db, q)
			return err
		})
	if err != nil {
		return nil, nil, err
	}
	return products, page, nil
}
