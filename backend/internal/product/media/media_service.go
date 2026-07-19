package media

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type MediaService struct {
	dr database.Runner
	mr *MediaRepository
}

func NewService(
	dr database.Runner,
	mr *MediaRepository,
) *MediaService {
	return &MediaService{
		dr: dr,
		mr: mr,
	}
}

type CreateProductMediaInput struct {
	VariantID uuid.UUID
	ObjectID  uuid.UUID
	Type      string
	SortOrder int
}

func (ms *MediaService) CreateProductMedia(
	ctx context.Context,
	input CreateProductMediaInput,
) (*model.ProductMedia, error) {

	now := time.Now().UTC()
	media := &model.ProductMedia{
		ID:        uuid.New(),
		VariantID: input.VariantID,
		ObjectID:  input.ObjectID,
		SortOrder: input.SortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}

	validType, err := model.ValidateMediaType(input.Type)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusBadRequest,
			security.CodeValidation,
			"invalid media type",
			err,
		)
	}

	media.Type = validType

	err = ms.dr.WithDB(ctx,
		func(db database.QueryExecutor) error {
			return ms.mr.Create(
				ctx,
				db,
				media,
			)
		})
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to create a new product media",
			err,
		)
	}

	return media, nil
}

func (ms *MediaService) GetProductMediaByID(
	ctx context.Context,
	mediaID uuid.UUID,
) (*model.ProductMedia, error) {
	var media *model.ProductMedia
	err := ms.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			m, err := ms.mr.GetByID(
				ctx,
				db,
				mediaID,
			)
			if err != nil {
				return err
			}
			media = m
			return nil
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch a media",
			err,
		)
	}
	return media, nil
}

func (ms *MediaService) GetVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
) ([]*model.ProductMedia, error) {
	var media []*model.ProductMedia
	err := ms.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			m, err := ms.mr.GetVariantMedia(
				ctx,
				db,
				variantID,
			)
			if err != nil {
				return err
			}

			media = m
			return nil
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch a media",
			err,
		)
	}
	return media, nil
}
