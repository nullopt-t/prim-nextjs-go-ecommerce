package tag

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
type TagService struct {
	dr   database.Runner
	repo *TagRepository
}

func NewService(
	dr database.Runner,
	repo *TagRepository,
) *TagService {
	return &TagService{
		dr:   dr,
		repo: repo,
	}
}

type CreateTagInput struct {
	Name string
}

func (ts *TagService) CreateTag(
	ctx context.Context,
	in CreateTagInput,
) (*model.ProductTag, error) {
	tag := &model.ProductTag{
		ID:       uuid.New(),
		PublicID: uuid.New(),
		Name:     strings.TrimSpace(in.Name),
	}

	err := ts.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return ts.repo.Create(ctx, db, tag)
	})

	if err != nil {
		return nil, ts.handleError(err, "Failed to create product tag")
	}

	return tag, nil
}

func (ts *TagService) GetTagByID(
	ctx context.Context,
	tagID uuid.UUID,
) (*model.ProductTag, error) {
	var tag *model.ProductTag
	err := ts.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		tag, repoErr = ts.repo.GetByID(ctx, db, tagID)
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.ErrNotFound("Product tag not found").
				WithCode(errcode.CodeTagNotFound).
				Wrap(err)

		default:
			return nil, apierr.ErrInternalError("Failed to fetch product tag").
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

func (ts *TagService) UpdateTag(
	ctx context.Context,
	tagID uuid.UUID,
	in UpdateTagInput,
) error {
	fields := UpdateTagFields{}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return apierr.ErrBadRequest("Validation error").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{
					Field:   "name",
					Message: "tag name cannot be empty",
				})
		}
		fields.Name = &name
	}

	err := ts.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return ts.repo.Update(ctx, db, tagID, fields)
	})

	if err != nil {
		return ts.handleError(err, "Failed to update product tag")
	}

	return nil
}

type ListTagsInput struct {
	Query          pagination.ListQuery
	IncludeDeleted bool
}

func (ts *TagService) ListTags(
	ctx context.Context,
	in ListTagsInput,
) (*pagination.PagedResult[model.ProductTag], error) {
	q := in.Query

	var result *pagination.PagedResult[model.ProductTag]

	err := ts.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = ts.repo.List(ctx, db, ListTagOptions{
			Query:          &q,
			IncludeDeleted: in.IncludeDeleted,
		})
		return repoErr
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to list product tags").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return result, nil
}

func (ts *TagService) AdminList(
	ctx context.Context,
	q pagination.ListQuery,
) (*pagination.PagedResult[model.ProductTag], error) {
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
		return apierr.ErrBadRequest("Tag ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := ts.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return ts.repo.Delete(ctx, db, tagID)
	})

	if err != nil {
		return ts.handleError(err, "Failed to delete product tag")
	}

	return nil
}

func (ts *TagService) ReplaceProductTags(
	ctx context.Context,
	productID uuid.UUID,
	tagIDs []uuid.UUID,
) error {
	if productID == uuid.Nil {
		return apierr.ErrBadRequest("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := ts.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return ts.repo.ReplaceTagsForProduct(ctx, tx, productID, tagIDs)
	})

	if err != nil {
		return ts.handleError(err, "Failed to update product tags")
	}

	return nil
}

func (ts *TagService) GetTagsByProductID(
	ctx context.Context,
	productID uuid.UUID,
) ([]*model.ProductTag, error) {
	if productID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var tags []*model.ProductTag
	err := ts.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		tags, err = ts.repo.GetTagsByProductID(ctx, db, productID)
		return err
	})

	if err != nil {
		return nil, ts.handleError(err, "Failed to fetch product tags")
	}

	return tags, nil
}

func (ts *TagService) handleError(err error, defaultMsg string) error {
	if err == nil {
		return nil
	}
	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.ErrNotFound("Product tag not found").
			WithCode(errcode.CodeTagNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ErrConflict("A tag with this name already exists").
			WithCode(errcode.CodeTagAlreadyExists).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrInvalidInput):
		return apierr.ErrBadRequest("Invalid input provided").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	default:
		return apierr.ErrInternalError(defaultMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}
