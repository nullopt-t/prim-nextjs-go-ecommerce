package tag

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

type TagService struct {
	qexecuter database.Runner
	trepo     *TagRepository
}

func NewService(
	r database.Runner,
	tr *TagRepository,
) *TagService {
	return &TagService{
		qexecuter: r,
		trepo:     tr,
	}
}

type CreateTagInput struct {
	Name string
}

type UpdateTagInput struct {
	Name *string
}

// CreateTag validates and creates a new product tag.
func (ts *TagService) CreateTag(
	ctx context.Context,
	in *CreateTagInput,
) (*model.ProductTag, error) {
	if in == nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_PAYLOAD").
			WithMessage("Request body is required")
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("VALIDATION_FAILED").
			WithMessage("Validation error").
			WithFields(api.FieldError{
				Field:   "name",
				Message: "tag name cannot be empty",
			})
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	tag := &model.ProductTag{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := ts.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return ts.trepo.Create(ctx, db, tag)
	})

	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(mappedError, database.ErrConflict):
			return nil, security.NewSecureError(http.StatusConflict).
				WithCode("TAG_ALREADY_EXISTS").
				WithMessage("A tag with this name already exists").
				Wrap(err)

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to create tag").
				Wrap(err).
				WithStack()
		}
	}

	return tag, nil
}

// GetTagByID retrieves an active tag by UUID.
func (ts *TagService) GetTagByID(
	ctx context.Context,
	tagID uuid.UUID,
) (*model.ProductTag, error) {
	if tagID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Tag ID is required")
	}

	var tag *model.ProductTag
	err := ts.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		tag, repoErr = ts.trepo.Get(ctx, db, &Filter{ID: &tagID})
		return repoErr
	})

	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(mappedError, database.ErrNotFound):
			return nil, security.NewSecureError(http.StatusNotFound).
				WithCode("TAG_NOT_FOUND").
				WithMessage("Tag not found").
				Wrap(err)

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to fetch tag").
				Wrap(err).
				WithStack()
		}
	}

	return tag, nil
}

// UpdateByID modifies an active tag.
func (ts *TagService) UpdateByID(
	ctx context.Context,
	tagID uuid.UUID,
	input UpdateTagInput,
) error {
	if tagID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Tag ID is required")
	}

	fields := UpdateTagFields{}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return security.NewSecureError(http.StatusBadRequest).
				WithCode("VALIDATION_FAILED").
				WithMessage("Validation error").
				WithFields(api.FieldError{
					Field:   "name",
					Message: "tag name cannot be empty",
				})
		}
		fields.Name = &name
	}

	err := ts.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return ts.trepo.Update(ctx, db, tagID, fields)
	})

	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(mappedError, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("TAG_NOT_FOUND").
				WithMessage("Tag not found").
				Wrap(err)

		case errors.Is(mappedError, database.ErrConflict):
			return security.NewSecureError(http.StatusConflict).
				WithCode("TAG_ALREADY_EXISTS").
				WithMessage("A tag with this name already exists").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to update tag").
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// DeleteByID soft-deletes an active tag.
func (ts *TagService) DeleteByID(
	ctx context.Context,
	tagID uuid.UUID,
) error {
	if tagID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Tag ID is required")
	}

	err := ts.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return ts.trepo.Delete(ctx, db, tagID)
	})

	if err != nil {
		mappedError := database.MapError(err)
		switch {
		case errors.Is(mappedError, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("TAG_NOT_FOUND").
				WithMessage("Tag not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to delete tag").
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// List handles unified paginated list operations for public and admin callers.
func (ts *TagService) List(
	ctx context.Context,
	q *api.ListQuery,
	includeDeleted bool,
) (*api.PagedResult[model.ProductTag], error) {
	if q == nil {
		q = &api.ListQuery{}
	}

	var result *api.PagedResult[model.ProductTag]
	err := ts.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = ts.trepo.List(ctx, db, ListTagOptions{
			Query:          q,
			IncludeDeleted: includeDeleted,
		})
		return repoErr
	})

	if err != nil {
		return nil, security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to list tags").
			Wrap(err).
			WithStack()
	}

	return result, nil
}

func (ts *TagService) ListTags(
	ctx context.Context,
	q *api.ListQuery,
) (*api.PagedResult[model.ProductTag], error) {
	return ts.List(ctx, q, false)
}

func (ts *TagService) AdminListTags(
	ctx context.Context,
	q *api.ListQuery,
) (*api.PagedResult[model.ProductTag], error) {
	return ts.List(ctx, q, true)
}
