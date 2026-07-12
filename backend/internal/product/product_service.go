package product

import (
	"context"
	"errors"
	"net/http"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"

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

type CreateProductInput struct {
	BrandID          uuid.UUID
	Title            string
	ShortDescription string
	Description      string
	Slug             string
	Status           model.ProductStatus
}

func (s *ProductService) Create(
	ctx context.Context,
	input CreateProductInput,
) (*model.Product, error) {
	userID := ctx.Value("userID").(uuid.UUID)
	product := &model.Product{
		ID:               uuid.New(),
		BrandID:          input.BrandID,
		Title:            input.Title,
		ShortDescription: input.ShortDescription,
		Description:      input.Description,
		Slug:             input.Slug,
		Status:           input.Status,
		CreatedBy:        userID,
		UpdatedBy:        userID,
	}

	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			err := s.productRepo.Create(ctx, db, product)
			if err != nil {
				return err
			}
			return nil
		},
	)

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(
			mappedErr,
			database.ErrConflict,
		):
			return nil, security.NewSecureError(
				http.StatusConflict,
				security.CodeConflict,
				"product already exists",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to create a new product",
				err,
			)
		}
	}

	return product, nil
}

func (s *ProductService) GetByID(
	ctx context.Context,
	productID uuid.UUID,
) (*model.Product, error) {
	var p *model.Product
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			prod, err := s.productRepo.Get(ctx, db, Filter{
				ID: &productID,
			})
			if err != nil {
				return err
			}
			p = prod
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
				"product not found",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to fetch a product",
				err,
			)
		}
	}
	return p, err
}

func (s *ProductService) GetBySlug(
	ctx context.Context,
	slug string,
) (*model.Product, error) {
	var p *model.Product
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			prod, err := s.productRepo.Get(ctx, db, Filter{
				Slug: &slug,
			})
			if err != nil {
				return err
			}
			p = prod
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
				"product not found",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to fetch a product",
				err,
			)
		}
	}
	return p, err
}

func (s *ProductService) List(
	ctx context.Context,
	q *api.ListQuery,
) ([]*ProductListItem, *api.Page, error) {
	var products []*ProductListItem
	var page *api.Page
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			products, page, err = s.productRepo.List(ctx, db, q)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return nil, nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch products",
			err,
		)
	}
	return products, page, nil
}
