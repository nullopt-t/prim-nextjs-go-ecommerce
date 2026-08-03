package tag

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type TagService struct {
	qexecuter database.Runner
	trepo     *TagRepository
}

func NewService(
	qexecuter database.Runner,
	r *TagRepository,
) *TagService {
	return &TagService{
		qexecuter: qexecuter,
		trepo:     r,
	}
}

type CreateTagInput struct {
	Name string `json:"name"`
}

func (ts *TagService) CreateTag(
	ctx context.Context,
	in *CreateTagInput,
) (*model.ProductTag, error) {
	if in == nil {
		return nil, apierr.New(http.StatusBadRequest, "Request body cannot be empty").
			WithCode(apierr.CodeInvalidPayload)
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apierr.New(http.StatusBadRequest, "Validation error").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{
				Field:   "name",
				Message: "tag name is required and cannot be empty",
			})
	}

	now := time.Now().UTC()
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
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, apierr.New(http.StatusConflict, "A tag with this name already exists").
				WithCode(errcode.CodeTagAlreadyExists).
				Wrap(err)

		default:
			return nil, apierr.New(http.StatusInternalServerError, "Failed to create product tag").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return tag, nil
}

func (ts *TagService) GetTagByID(
	ctx context.Context,
	tagID uuid.UUID,
) (*model.ProductTag, error) {
	if tagID == uuid.Nil {
		return nil, apierr.New(http.StatusBadRequest, "Tag ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var tag *model.ProductTag
	err := ts.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		tag, repoErr = ts.trepo.Get(ctx, db, &Filter{ID: &tagID})
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.New(http.StatusNotFound, "Product tag not found").
				WithCode(errcode.CodeTagNotFound).
				Wrap(err)

		default:
			return nil, apierr.New(http.StatusInternalServerError, "Failed to fetch product tag").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return tag, nil
}

type UpdateTagInput struct {
	Name *string
}

func (ts *TagService) UpdateTagByID(
	ctx context.Context,
	tagID uuid.UUID,
	in *UpdateTagInput,
) error {
	if tagID == uuid.Nil {
		return apierr.New(http.StatusBadRequest, "Tag ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if in == nil {
		return apierr.New(http.StatusBadRequest, "Update payload cannot be empty").
			WithCode(apierr.CodeInvalidPayload)
	}

	fields := UpdateTagFields{}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return apierr.New(http.StatusBadRequest, "Validation error").
				WithCode(apierr.CodeValidationFailed).
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
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.New(http.StatusNotFound, "Product tag not found").
				WithCode(errcode.CodeTagNotFound).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrConflict):
			return apierr.New(http.StatusConflict, "A tag with this name already exists").
				WithCode(errcode.CodeTagAlreadyExists).
				Wrap(err)

		default:
			return apierr.New(http.StatusInternalServerError, "Failed to update product tag").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

type ListTagsInput struct {
	Query          *api.ListQuery
	IncludeDeleted bool
}

func (ts *TagService) ListTags(
	ctx context.Context,
	in ListTagsInput,
) (*api.PagedResult[model.ProductTag], error) {
	q := in.Query
	if q == nil {
		q = &api.ListQuery{}
	}

	var result *api.PagedResult[model.ProductTag]

	err := ts.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = ts.trepo.List(ctx, db, ListTagOptions{
			Query:          q,
			IncludeDeleted: in.IncludeDeleted,
		})
		return repoErr
	})

	if err != nil {
		return nil, apierr.New(http.StatusInternalServerError, "Failed to list product tags").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return result, nil
}

func (ts *TagService) AdminList(
	ctx context.Context,
	q *api.ListQuery,
) (*api.PagedResult[model.ProductTag], error) {
	return ts.ListTags(ctx, ListTagsInput{
		Query:          q,
		IncludeDeleted: true,
	})
}

func (ts *TagService) DeleteTagByID(
	ctx context.Context,
	tagID uuid.UUID,
) error {
	if tagID == uuid.Nil {
		return apierr.New(http.StatusBadRequest, "Tag ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := ts.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return ts.trepo.Delete(ctx, db, tagID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.New(http.StatusNotFound, "Product tag not found").
				WithCode(errcode.CodeTagNotFound).
				Wrap(err)

		default:
			return apierr.New(http.StatusInternalServerError, "Failed to delete product tag").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (ts *TagService) ReplaceProductTags(
	ctx context.Context,
	productID uuid.UUID,
	tagIDs []uuid.UUID,
) error {
	if productID == uuid.Nil {
		return apierr.New(http.StatusBadRequest, "Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := ts.qexecuter.WithTx(ctx, func(tx database.QueryExecutor) error {
		return ts.trepo.ReplaceTagsForProduct(ctx, tx, productID, tagIDs)
	})

	if err != nil {
		return apierr.New(http.StatusInternalServerError, "Failed to update product tags").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}
