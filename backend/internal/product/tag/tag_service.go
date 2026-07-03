package tag

import (
	"context"
	"errors"
	"net/http"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
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
) (*model.Tag, error) {
	tag := model.NewTag(in.Name)
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
