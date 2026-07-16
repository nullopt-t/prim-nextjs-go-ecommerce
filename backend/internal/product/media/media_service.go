package media

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type ProductMediaService struct {
	qexecutor database.Runner
	mr        *MediaRepository
}

func NewProductMediaService(
	qexecutor database.Runner,
	mr *MediaRepository,
) *ProductMediaService {
	return &ProductMediaService{
		qexecutor: qexecutor,
		mr:        mr,
	}
}

type CreateMediaInput struct {
	Alt       string
	Type      model.MediaType
	MimeType  string
	FileSize  int64
	Checksum  string
	Width     int
	Height    int
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

func (ms *ProductMediaService) CreateMedia(
	ctx context.Context,
	input CreateMediaInput,
) error {

	now := time.Now()
	media := &model.Media{
		ID:        uuid.New(),
		Alt:       input.Alt,
		Type:      input.Type,
		MimeType:  input.MimeType,
		FileSize:  input.FileSize,
		Checksum:  input.Checksum,
		Width:     input.Width,
		Height:    input.Height,
		CreatedBy: input.CreatedBy,
		UpdatedBy: input.UpdatedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := ms.qexecutor.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return ms.mr.Create(
				ctx,
				db,
				media,
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
			return security.NewSecureError(
				http.StatusConflict,
				security.CodeConflict,
				"conflict",
				err,
			)
		default:
			return security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"internal server error",
				err,
			)
		}
	}
	return nil
}

func (ms *ProductMediaService) GetByID(
	ctx context.Context,
	mediaID uuid.UUID,
) ([]*model.Media, error) {
	var media []*model.Media
	err := ms.qexecutor.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			var err error
			media, err = ms.mr.GetByID(
				ctx,
				db,
				mediaID,
			)
			if err != nil {
				return err
			}
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
				"not found",
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

	return media, nil
}

func (ms *ProductMediaService) GetByChecksum(
	ctx context.Context,
	checksum string,
) (*model.Media, error) {
	var media *model.Media
	err := ms.qexecutor.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			var err error
			media, err = ms.mr.GetByChecksum(
				ctx,
				db,
				checksum,
			)
			if err != nil {
				return err
			}
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
				"not found",
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

	return media, nil
}

type UpdateMediaInput struct {
	Alt       *string
	Type      *string
	Checksum  *string
	MimeType  *string
	Width     *int
	Height    *int
	FileSize  *int64
	UpdatedBy uuid.UUID
}

func (ms *ProductMediaService) Update(
	ctx context.Context,
	mediaID uuid.UUID,
	input UpdateMediaInput,
) error {
	media := UpdateMediaFields{
		Alt:       input.Alt,
		Type:      input.Type,
		Checksum:  input.Checksum,
		MimeType:  input.MimeType,
		Width:     input.Width,
		Height:    input.Height,
		FileSize:  input.FileSize,
		UpdatedBy: input.UpdatedBy,
		UpdatedAt: time.Now().UTC(),
	}
	err := ms.qexecutor.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return ms.mr.Update(
				ctx,
				db,
				mediaID,
				media,
			)
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (ms *ProductMediaService) Delete(
	ctx context.Context,
	mediaID uuid.UUID,
) error {
	err := ms.qexecutor.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return ms.mr.Delete(
				ctx,
				db,
				mediaID,
			)
		},
	)
	if err != nil {
		return err
	}
	return nil
}
