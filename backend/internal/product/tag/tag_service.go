package tag

import (
	"context"
	"errors"
	"net/http"
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

func (ts *TagService) CreateTag(
	ctx context.Context,
	in CreateTagInput,
) (*model.ProductTag, error) {
	userID := ctx.Value("userID").(uuid.UUID)

	now := time.Now()
	tag := &model.ProductTag{
		ID:        uuid.New(),
		Name:      in.Name,
		CreatedBy: userID,
		UpdatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := ts.qexecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return ts.trepo.Create(
				ctx,
				db,
				tag,
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
				"resource already existed",
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

	return tag, nil
}

func (ts *TagService) GetTagByID(
	ctx context.Context,
	tagID uuid.UUID,
) (*model.ProductTag, error) {
	var tag *model.ProductTag
	err := ts.qexecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			t, err := ts.trepo.Get(
				ctx,
				db,
				&Filter{
					ID: &tagID,
				},
			)
			if err != nil {
				return err
			}
			tag = t
			return nil
		})
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

	return tag, nil
}

func (ts *TagService) PutProductTags(
	ctx context.Context,
	productID uuid.UUID,
	tagsIDs []uuid.UUID,
) error {
	err := ts.qexecuter.WithTx(ctx,
		func(tx database.QueryExecutor) error {
			return ts.trepo.PutProductTags(
				ctx,
				tx,
				productID,
				tagsIDs,
			)
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"internal server error",
			err,
		)
	}
	return nil
}

func (ts *TagService) ListProductTags(
	ctx context.Context,
	productID uuid.UUID,
) ([]*model.ProductTag, error) {
	var tags []*model.ProductTag
	err := ts.qexecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			t, err := ts.trepo.ListProductTags(
				ctx,
				db,
				productID,
			)
			if err != nil {
				return err
			}
			tags = t
			return nil
		})
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"internal server error",
			err,
		)
	}
	return tags, nil
}

func (ts *TagService) ListTags(
	ctx context.Context,
	q *api.ListQuery,
) ([]*model.ProductTag, *api.Page, error) {
	var tags []*model.ProductTag
	var page *api.Page
	err := ts.qexecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			ts, p, err := ts.trepo.List(
				ctx,
				db,
				q,
			)
			if err != nil {
				return err
			}
			tags = ts
			page = p
			return nil
		},
	)
	if err != nil {
		return nil, nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch the categories",
			err,
		)
	}
	return tags, page, nil
}
