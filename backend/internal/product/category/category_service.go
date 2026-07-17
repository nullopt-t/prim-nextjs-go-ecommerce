package category

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type CategoryService struct {
	crepository *CategoryRepository
	qexecuter   database.Runner
}

func NewService(
	qexecuter database.Runner,
	r *CategoryRepository,
) *CategoryService {
	return &CategoryService{
		crepository: r,
		qexecuter:   qexecuter,
	}
}

type CreateCategoryInput struct {
	Name      string
	ParentID  *uuid.UUID
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

func (cs *CategoryService) CreateCategory(
	ctx context.Context,
	in *CreateCategoryInput,
) (*model.ProductCategory, error) {

	now := time.Now().UTC()
	category := &model.ProductCategory{
		ID:        uuid.New(),
		Name:      in.Name,
		ParentID:  in.ParentID,
		CreatedBy: in.CreatedBy,
		UpdatedBy: in.UpdatedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := cs.qexecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return cs.crepository.Create(
				ctx,
				db,
				category,
			)
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
				"resource already exists",
				err,
			)
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to create a new resource",
				err,
			)
		}
	}
	return category, err
}

func (cs *CategoryService) GetCategoryByID(
	ctx context.Context,
	categoryID uuid.UUID,
) (*model.ProductCategory, error) {
	var category *model.ProductCategory
	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		c, err := cs.crepository.Get(
			ctx,
			db,
			Filter{
				ID: &categoryID,
			},
		)

		if err != nil {
			return err
		}

		category = c
		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(
			mappedErr,
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
				"failed to fetch the resource",
				err,
			)
		}
	}

	return category, nil
}

type UpdateCategoryInput struct {
	Name      *string
	ParentID  *uuid.UUID
	UpdatedBy uuid.UUID
}

func (cs *CategoryService) UpdateCategory(
	ctx context.Context,
	categoryID uuid.UUID,
	input *UpdateCategoryInput,
) error {
	fields := UpdateCategoryFields{
		Name:      input.Name,
		ParentID:  input.ParentID,
		UpdatedBy: input.UpdatedBy,
	}
	err := cs.qexecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			return cs.crepository.Update(
				ctx,
				db,
				categoryID,
				fields,
			)
		},
	)
	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(
			mappedErr,
			database.ErrNotFound,
		):
			return security.NewSecureError(
				http.StatusNotFound,
				security.CodeNotFound,
				"resource not found",
				err,
			)
		default:
			return security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to fetch the resource",
				err,
			)
		}
	}

	return nil
}

func (cs *CategoryService) PutProductCategories(
	ctx context.Context,
	productID uuid.UUID,
	categoryIDs []uuid.UUID,
) error {
	err := cs.qexecuter.WithTx(
		ctx,
		func(tx database.QueryExecutor) error {
			return cs.crepository.PutProductCategories(
				ctx,
				tx,
				productID,
				categoryIDs,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("put product categories: %w", err)
	}
	return nil
}

func (cs *CategoryService) ListProductCategories(
	ctx context.Context,
	productID uuid.UUID,
) ([]*model.ProductCategory, error) {
	var res []*model.ProductCategory
	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		categories, err := cs.crepository.ListProductCategories(
			ctx,
			db,
			productID,
		)
		if err != nil {
			return err
		}
		res = categories
		return nil
	})
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch the product categories",
			err,
		)
	}
	return res, nil
}

func (cs *CategoryService) List(
	ctx context.Context,
	q *api.ListQuery,
) ([]*model.ProductCategory, *api.Page, error) {
	var res []*model.ProductCategory
	var page *api.Page
	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		categories, p, err := cs.crepository.List(
			ctx,
			db,
			q,
		)
		if err != nil {
			return err
		}
		res = categories
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
