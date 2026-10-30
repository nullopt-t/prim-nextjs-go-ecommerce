package category

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type CategoryService struct {
	repo *CategoryRepository
	dr   database.Runner
}

func NewService(
	dr database.Runner,
	repo *CategoryRepository,
) *CategoryService {
	if dr == nil {
		panic("category: database runner is required")
	}
	if repo == nil {
		panic("category: category repository is required")
	}
	return &CategoryService{
		repo: repo,
		dr:   dr,
	}
}

type CreateCategoryInput struct {
	Name     string
	ParentID *uuid.UUID
}

func (cs *CategoryService) CreateCategory(
	ctx context.Context,
	in CreateCategoryInput,
) (*model.ProductCategory, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apierr.BadRequestError("Invalid category parameters").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{
				Field:   "name",
				Message: "name is required and cannot be empty",
			})
	}

	category := &model.ProductCategory{
		ID:       uuid.New(),
		ParentID: in.ParentID,
		Name:     name,
	}

	err := cs.dr.WithTx(ctx, func(ctx context.Context) error {
		return cs.repo.Create(ctx, category)
	})

	if err != nil {
		return nil, cs.handleError(err, "An unexpected database error occurred")
	}

	return category, nil
}

func (cs *CategoryService) GetCategoryByID(
	ctx context.Context,
	categoryID uuid.UUID,
) (*model.ProductCategory, error) {
	var category *model.ProductCategory
	err := cs.dr.WithExecutor(ctx, func(ctx context.Context) error {
		var repoErr error
		category, repoErr = cs.repo.GetByID(ctx, categoryID)
		return repoErr
	})

	if err != nil {
		return nil, cs.handleError(err, "Failed to fetch category")
	}

	return category, nil
}

func (cs *CategoryService) isDescendant(
	ctx context.Context,
	categoryID uuid.UUID,
	newParentID uuid.UUID,
) (bool, error) {
	visited := make(map[uuid.UUID]struct{})
	currentParent := &newParentID
	for currentParent != nil && *currentParent != uuid.Nil {
		if *currentParent == categoryID {
			return true, nil
		}
		if _, seen := visited[*currentParent]; seen {
			// Cycle detected in existing hierarchy
			return false, errors.New("cyclic category hierarchy detected")
		}
		visited[*currentParent] = struct{}{}

		parentCategory, err := cs.repo.GetByID(ctx, *currentParent)
		if err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrNotFound) {
				return false, nil
			}
			return false, err
		}
		currentParent = parentCategory.ParentID
	}
	return false, nil
}

type UpdateCategoryInput struct {
	Name     *string
	ParentID *uuid.UUID
}

func (cs *CategoryService) UpdateCategory(
	ctx context.Context,
	categoryID uuid.UUID,
	input UpdateCategoryInput,
) error {
	if categoryID == uuid.Nil {
		return apierr.BadRequestError("Category ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	fields := UpdateCategoryFields{}

	if input.Name != nil {
		trimmedName := strings.TrimSpace(*input.Name)
		if trimmedName == "" {
			return apierr.BadRequestError("Validation failed").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{
					Field:   "name",
					Message: "category name cannot be empty",
				})
		}
		fields.Name = &trimmedName
	}

	if input.ParentID != nil {
		parentID := *input.ParentID

		if parentID == categoryID {
			return apierr.BadRequestError("A category cannot be set as its own parent").
				WithCode(errcode.CodeInvalidCategoryHierarchy).
				WithFields(api.FieldError{
					Field:   "parent_id",
					Message: "category cannot reference itself as parent",
				})
		}

		if parentID != uuid.Nil {
			fields.ParentID = input.ParentID
		} else {
			// uuid.Nil signifies clearing/unsetting the parent category to make it a top-level category (NULL in DB)
			nilParent := uuid.Nil
			fields.ParentID = &nilParent
		}
	}

	err := cs.dr.WithTx(ctx, func(ctx context.Context) error {
		if input.ParentID != nil && *input.ParentID != uuid.Nil {
			isDescendant, descErr := cs.isDescendant(ctx, categoryID, *input.ParentID)
			if descErr != nil {
				return apierr.InternalServerError("Failed to validate category hierarchy").
					WithCode(apierr.CodeInternalError).
					Wrap(descErr).
					WithStack()
			}
			if isDescendant {
				return apierr.BadRequestError("Cannot set a descendant category as parent").
					WithCode(errcode.CodeCyclicCategoryHierarchy).
					WithFields(api.FieldError{
						Field:   "parent_id",
						Message: "causes a circular reference in the category tree",
					})
			}
		}

		return cs.repo.Update(ctx, categoryID, fields)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return cs.handleError(err, "Failed to update category")
	}

	return nil
}

type ListCategoriesInput struct {
	Query            *pagination.ListQuery
	IncludeDeleted   bool
	OnlyWithProducts bool
	WithParentName   bool
	WithProductCount bool
}

func (cs *CategoryService) ListCategories(
	ctx context.Context,
	in ListCategoriesInput,
) (*pagination.PagedResult[model.ProductCategory], error) {
	q := in.Query
	if q == nil {
		q = &pagination.ListQuery{}
	}

	cursor, err := q.Keyset()
	if err != nil {
		return nil, apierr.BadRequestError("Invalid pagination cursor").Wrap(err)
	}

	var result *pagination.PagedResult[model.ProductCategory]

	err = cs.dr.WithExecutor(ctx, func(ctx context.Context) error {
		var repoErr error
		result, repoErr = cs.repo.List(ctx, ListCategoryOptions{
			ListQuery:        q,
			IncludeDeleted:   in.IncludeDeleted,
			OnlyWithProducts: in.OnlyWithProducts,
			WithParentName:   in.WithParentName,
			WithProductCount: in.WithProductCount,
			Cursor:           cursor,
		})
		return repoErr
	})

	if err != nil {
		return nil, cs.handleError(err, "Failed to list categories")
	}

	return result, nil
}

func (cs *CategoryService) List(
	ctx context.Context,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductCategory], error) {
	return cs.ListCategories(ctx, ListCategoriesInput{
		Query:            q,
		IncludeDeleted:   false,
		OnlyWithProducts: true,
	})
}

func (cs *CategoryService) AdminList(
	ctx context.Context,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductCategory], error) {
	if q == nil {
		q = &pagination.ListQuery{}
	}
	return cs.ListCategories(ctx, ListCategoriesInput{
		Query:            q,
		IncludeDeleted:   true,
		WithParentName:   true,
		WithProductCount: true,
	})
}

func (cs *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return cs.dr.WithTx(ctx, func(ctx context.Context) error {
		if err := cs.repo.SoftDeleteCategory(ctx, id); err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Category not found").WithCode(apierr.CodeNotFound)
			}
			return err
		}
		return nil
	})
}

func (cs *CategoryService) handleError(err error, defaultMsg string) error {
	if err == nil {
		return nil
	}
	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError("Category not found").
			WithCode(errcode.CodeCategoryNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Category already exists").
			WithCode(errcode.CodeCategoryAlreadyExists).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("The specified parent category does not exist").
			WithCode(errcode.CodeParentCategoryNotFound).
			Wrap(err).
			WithFields(api.FieldError{
				Field:   "parent_id",
				Message: "referenced category does not exist",
			})
	case errors.Is(mappedErr, database.ErrInvalidInput):
		return apierr.BadRequestError("Invalid input provided").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	default:
		return apierr.InternalServerError(defaultMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}
