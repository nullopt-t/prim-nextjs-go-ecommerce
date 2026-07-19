package variant

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/media"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type VariantService struct {
	dr database.Runner
	vr *VariantRepository
	ms *media.MediaService
}

func NewService(
	r database.Runner,
	vr *VariantRepository,
	ms *media.MediaService,
) *VariantService {
	return &VariantService{
		dr: r,
		vr: vr,
		ms: ms,
	}
}

type CreateVariantInput struct {
	ProductID uuid.UUID
	SKU       *string
	Price     int64
	Currency  string
}

func (vs *VariantService) CreateVariant(
	ctx context.Context,
	in CreateVariantInput,
) (*model.ProductVariant, error) {

	now := time.Now().UTC()
	variant := &model.ProductVariant{
		ID:        uuid.New(),
		ProductID: in.ProductID,
		SKU:       in.SKU,
		Price:     in.Price,
		Currency:  in.Currency,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := vs.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return vs.vr.Create(
				ctx,
				db,
				variant,
			)
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to create a new variant",
			err,
		)
	}
	return variant, nil
}

func (vs *VariantService) GetVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
) ([]*model.ProductMedia, error) {
	var media []*model.ProductMedia
	err := vs.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			m, err := vs.ms.GetVariantMedia(
				ctx,
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
			"failed to create a new variant",
			err,
		)
	}
	return media, nil
}

type UpdateVariantInput struct {
	ProductID *uuid.UUID
	SKU       *string
	Price     *int64
	Currency  *string
}

func (vs *VariantService) UpdateVariant(
	ctx context.Context,
	variantID uuid.UUID,
) error {
	err := vs.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			panic("TODO: delegate database variant updating to respository")
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to create a new variant",
			err,
		)
	}
	return nil
}
