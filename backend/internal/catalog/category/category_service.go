package category

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

// TODO(ai): create handlerError method for the errors

// TODO(ai): make the names descriptive
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

// TODO(ai): service should know nothing about respresentaion
type CreateCategoryInput struct {
	Name     string     `json:"name"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

func (cs *CategoryService) CreateCategory(
	ctx context.Context,
	in *CreateCategoryInput,
) (*model.ProductCategory, error) {
	if in == nil {
		return nil, apierr.ErrBadRequest("Request body is missing").
			WithCode(apierr.CodeInvalidPayload)
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apierr.ErrBadRequest("Invalid category parameters").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{
				Field:   "name",
				Message: "name is required and cannot be empty",
			})
	}

	category := &model.ProductCategory{
		ID:       uuid.New(),
		PublicID: uuid.NewString(),
		ParentID: in.ParentID,
		Name:     name,
	}

	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return cs.crepository.Create(ctx, db, category)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, apierr.ErrConflict("A category with this name already exists").
				WithCode(errcode.CodeCategoryAlreadyExists).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation), errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.ErrBadRequest("The specified parent_id category does not exist").
				WithCode(errcode.CodeParentCategoryNotFound).
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "parent_id",
					Message: "referenced category does not exist",
				})

		default:
			return nil, apierr.ErrInternalError("An unexpected database error occurred").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return category, nil
}

func (cs *CategoryService) GetCategoryByID(
	ctx context.Context,
	categoryID uuid.UUID,
) (*model.ProductCategory, error) {
	var category *model.ProductCategory
	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		category, repoErr = cs.crepository.GetByID(ctx, db, categoryID)
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.ErrNotFound("Category not found").
				WithCode(errcode.CodeCategoryNotFound).
				Wrap(err)

		default:
			return nil, apierr.ErrInternalError("Failed to fetch category").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return category, nil
}

func (cs *CategoryService) isDescendant(
	ctx context.Context,
	categoryID uuid.UUID,
	newParentID uuid.UUID,
) (bool, error) {
	var isChild bool
	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		currentParent := &newParentID
		for currentParent != nil && *currentParent != uuid.Nil {
			if *currentParent == categoryID {
				isChild = true
				return nil
			}
			parentCategory, err := cs.crepository.GetByID(ctx, db, *currentParent)
			if err != nil {
				mappedErr := database.MapError(err)
				if errors.Is(mappedErr, database.ErrNotFound) {
					return nil
				}
				return err
			}
			currentParent = parentCategory.ParentID
		}
		return nil
	})
	return isChild, err
}

type UpdateCategoryInput struct {
	Name     *string
	ParentID *uuid.UUID
}

func (cs *CategoryService) UpdateCategory(
	ctx context.Context,
	categoryID uuid.UUID,
	// TODO(ai): this shouldn't be optional
	input *UpdateCategoryInput,
) error {
	if categoryID == uuid.Nil {
		return apierr.ErrBadRequest("Category ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if input == nil {
		return apierr.ErrBadRequest("Update payload cannot be empty").
			WithCode(apierr.CodeInvalidPayload)
	}

	fields := UpdateCategoryFields{}

	if input.Name != nil {
		trimmedName := strings.TrimSpace(*input.Name)
		if trimmedName == "" {
			return apierr.ErrBadRequest("Validation failed").
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
			return apierr.ErrBadRequest("A category cannot be set as its own parent").
				WithCode(errcode.CodeInvalidCategoryHierarchy).
				WithFields(api.FieldError{
					Field:   "parent_id",
					Message: "category cannot reference itself as parent",
				})
		}

		if parentID != uuid.Nil {
			isDescendant, err := cs.isDescendant(ctx, categoryID, parentID)
			if err != nil {
				return apierr.ErrInternalError("Failed to validate category hierarchy").
					WithCode(apierr.CodeInternalError).
					Wrap(err).
					WithStack()
			}
			if isDescendant {
				return apierr.ErrBadRequest("Cannot set a descendant category as parent").
					WithCode(errcode.CodeCyclicCategoryHierarchy).
					WithFields(api.FieldError{
						Field:   "parent_id",
						Message: "causes a circular reference in the category tree",
					})
			}
		}

		fields.ParentID = input.ParentID
	}

	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return cs.crepository.Update(ctx, db, categoryID, fields)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Category not found").
				WithCode(errcode.CodeCategoryNotFound).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrConflict):
			return apierr.ErrConflict("A category with this name already exists").
				WithCode(errcode.CodeCategoryAlreadyExists).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return apierr.ErrBadRequest("The specified parent category does not exist").
				WithCode(errcode.CodeParentCategoryNotFound).
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "parent_id",
					Message: "referenced category does not exist",
				})

		default:
			return apierr.ErrInternalError("Failed to update category").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

type ListCategoriesInput struct {
	Query          *pagination.ListQuery
	IncludeDeleted bool
}

func (cs *CategoryService) ListCategories(
	ctx context.Context,
	in ListCategoriesInput,
) (*pagination.PagedResult[model.ProductCategory], error) {
	// TODO(ai): query shouldn't be optional by all meaning it could contain defaults but not nil by itself
	q := in.Query
	if q == nil {
		q = &pagination.ListQuery{}
	}

	var result *pagination.PagedResult[model.ProductCategory]

	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = cs.crepository.List(ctx, db, ListCategoryOptions{
			ListQuery:      q,
			IncludeDeleted: in.IncludeDeleted,
		})
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrInvalidInput):
			return nil, apierr.ErrBadRequest("Invalid search or sort filter provided").
				WithCode(apierr.CodeInvalidQueryParameter).
				Wrap(err)

		default:
			return nil, apierr.ErrInternalError("Failed to list categories").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return result, nil
}

func (cs *CategoryService) List(
	ctx context.Context,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductCategory], error) {
	return cs.ListCategories(ctx, ListCategoriesInput{
		Query:          q,
		IncludeDeleted: false,
	})
}

func (cs *CategoryService) AdminList(
	ctx context.Context,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.ProductCategory], error) {
	return cs.ListCategories(ctx, ListCategoriesInput{
		Query:          q,
		IncludeDeleted: true,
	})
}
