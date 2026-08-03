package category

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
	Name     string     `json:"name"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

func (cs *CategoryService) CreateCategory(
	ctx context.Context,
	in *CreateCategoryInput,
) (*model.ProductCategory, error) {
	if in == nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_PAYLOAD").
			WithMessage("Request body is missing")
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("VALIDATION_FAILED").
			WithMessage("Invalid category parameters").
			WithFields(api.FieldError{
				Field:   "name",
				Message: "name is required and cannot be empty",
			})
	}

	now := time.Now().UTC()
	category := &model.ProductCategory{
		ID:        uuid.New(),
		ParentID:  in.ParentID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return cs.crepository.Create(ctx, db, category)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, security.NewSecureError(http.StatusConflict).
				WithCode("CATEGORY_ALREADY_EXISTS").
				WithMessage("A category with this name already exists").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation), errors.Is(mappedErr, database.ErrNotFound):
			return nil, security.NewSecureError(http.StatusBadRequest).
				WithCode("PARENT_CATEGORY_NOT_FOUND").
				WithMessage("The specified parent_id category does not exist").
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "parent_id",
					Message: "referenced category does not exist",
				})

		default:
			// Attach stack trace for internal server errors
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("An unexpected database error occurred").
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
	if categoryID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Category ID cannot be empty").
			WithFields(api.FieldError{
				Field:   "id",
				Message: "valid UUID is required",
			})
	}

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
			return nil, security.NewSecureError(http.StatusNotFound).
				WithCode("CATEGORY_NOT_FOUND").
				WithMessage("Category not found").
				Wrap(err)

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to fetch category").
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
		// Traverse up from newParentID
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
					return nil // Let the update foreign key constraint return PARENT_CATEGORY_NOT_FOUND
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
	input *UpdateCategoryInput,
) error {
	// 1. Guard against nil IDs or payloads
	if categoryID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Category ID is required")
	}

	if input == nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_PAYLOAD").
			WithMessage("Update payload cannot be empty")
	}

	fields := UpdateCategoryFields{}

	// 2. Validate & sanitize Name if provided
	if input.Name != nil {
		trimmedName := strings.TrimSpace(*input.Name)
		if trimmedName == "" {
			return security.NewSecureError(http.StatusBadRequest).
				WithCode("VALIDATION_FAILED").
				WithMessage("Validation failed").
				WithFields(api.FieldError{
					Field:   "name",
					Message: "category name cannot be empty",
				})
		}
		fields.Name = &trimmedName
	}

	// 3. Hierarchy & Cycle Validation for ParentID
	if input.ParentID != nil {
		parentID := *input.ParentID

		// Guard against direct self-referencing parent
		if parentID == categoryID {
			return security.NewSecureError(http.StatusBadRequest).
				WithCode("INVALID_HIERARCHY").
				WithMessage("A category cannot be set as its own parent").
				WithFields(api.FieldError{
					Field:   "parent_id",
					Message: "category cannot reference itself as parent",
				})
		}

		// Cycle Check: Ensure new parent is not a child/descendant of current category
		if parentID != uuid.Nil {
			isDescendant, err := cs.isDescendant(ctx, categoryID, parentID)
			if err != nil {
				return security.NewSecureError(http.StatusInternalServerError).
					WithCode("INTERNAL_ERROR").
					WithMessage("Failed to validate category hierarchy").
					Wrap(err).
					WithStack()
			}
			if isDescendant {
				return security.NewSecureError(http.StatusBadRequest).
					WithCode("CYCLIC_HIERARCHY").
					WithMessage("Cannot set a descendant category as parent").
					WithFields(api.FieldError{
						Field:   "parent_id",
						Message: "causes a circular reference in the category tree",
					})
			}
		}

		fields.ParentID = input.ParentID
	}

	// 4. Execute Update
	err := cs.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return cs.crepository.Update(ctx, db, categoryID, fields)
	})

	// 5. Comprehensive Error Mapping
	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("CATEGORY_NOT_FOUND").
				WithMessage("Category not found").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrConflict):
			return security.NewSecureError(http.StatusConflict).
				WithCode("CATEGORY_ALREADY_EXISTS").
				WithMessage("A category with this name already exists").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return security.NewSecureError(http.StatusBadRequest).
				WithCode("PARENT_CATEGORY_NOT_FOUND").
				WithMessage("The specified parent category does not exist").
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "parent_id",
					Message: "referenced category does not exist",
				})

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to update category").
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

type ListCategoriesInput struct {
	Query          *api.ListQuery
	IncludeDeleted bool // set true for admin caller context
}

// ListCategories is the consolidated handler for public and admin category queries.
func (cs *CategoryService) ListCategories(
	ctx context.Context,
	in ListCategoriesInput,
) (*api.PagedResult[model.ProductCategory], error) {
	q := in.Query
	if q == nil {
		q = &api.ListQuery{}
	}

	var result *api.PagedResult[model.ProductCategory]

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
			return nil, security.NewSecureError(http.StatusBadRequest).
				WithCode("INVALID_QUERY_PARAMETER").
				WithMessage("Invalid search or sort filter provided").
				Wrap(err)

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to list categories").
				Wrap(err).
				WithStack()
		}
	}

	return result, nil
}

// Customer-facing list endpoint (Soft-deleted records hidden)
func (cs *CategoryService) List(
	ctx context.Context,
	q *api.ListQuery,
) (*api.PagedResult[model.ProductCategory], error) {
	return cs.ListCategories(ctx, ListCategoriesInput{
		Query:          q,
		IncludeDeleted: false,
	})
}

// Admin-facing list endpoint (Includes soft-deleted records)
func (cs *CategoryService) AdminList(
	ctx context.Context,
	q *api.ListQuery,
) (*api.PagedResult[model.ProductCategory], error) {
	return cs.ListCategories(ctx, ListCategoriesInput{
		Query:          q,
		IncludeDeleted: true,
	})
}
