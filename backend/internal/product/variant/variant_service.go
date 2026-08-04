package variant

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/minio/minio-go/v7"
)

type VariantService struct {
	logger      log.Logger
	dr          database.Runner
	minioClient *minio.Client
	vr          *VariantRepository
	minCfg      *config.MinioConfig
}

func NewService(
	logger log.Logger,
	r database.Runner,
	minioClient *minio.Client,
	vr *VariantRepository,
	minCfg *config.MinioConfig,
) *VariantService {
	return &VariantService{
		logger:      logger,
		dr:          r,
		minioClient: minioClient,
		vr:          vr,
		minCfg:      minCfg,
	}
}

type CreateVariantInput struct {
	ProductID       uuid.UUID
	Title           string
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	Attributes      map[string]any
	IsDefault       bool
}

type UpdateVariantInput struct {
	Title           *string
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	Attributes      map[string]any
	IsDefault       *bool
}

type AttachMediaInput struct {
	VariantID       uuid.UUID
	StorageObjectID uuid.UUID
	MediaType       string
	SortOrder       int
}

func (vs *VariantService) CreateVariant(
	ctx context.Context,
	in *CreateVariantInput,
) (*model.ProductVariant, error) {
	if in == nil {
		return nil, apierr.ErrBadRequest("Request body is required").
			WithCode(apierr.CodeInvalidPayload)
	}

	if in.ProductID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Validation error").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{
				Field:   "product_id",
				Message: "product_id is required",
			})
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, apierr.ErrBadRequest("Validation error").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{
				Field:   "title",
				Message: "title cannot be empty",
			})
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	variant := &model.ProductVariant{
		ID:              uuid.New(),
		PublicID:        uuid.NewString(),
		ProductID:       in.ProductID,
		Title:           title,
		Price:           in.Price,
		CrossedOutPrice: in.CrossedOutPrice,
		Currency:        in.Currency,
		Attributes:      in.Attributes,
		IsDefault:       in.IsDefault,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	execFunc := vs.dr.WithDB
	if variant.IsDefault {
		execFunc = vs.dr.WithTx
	}

	err := execFunc(ctx, func(db database.QueryExecutor) error {
		if variant.IsDefault {
			if err := vs.vr.ClearDefaultFlags(ctx, db, variant.ProductID); err != nil {
				return err
			}
		}
		return vs.vr.Create(ctx, db, variant)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.ErrBadRequest("Referenced product does not exist").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "product_id",
					Message: "product reference is invalid",
				})

		default:
			return nil, apierr.ErrInternalError("Failed to create variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return variant, nil
}

func (vs *VariantService) GetVariantByID(
	ctx context.Context,
	variantID uuid.UUID,
) (*model.ProductVariant, error) {
	var variant *model.ProductVariant
	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		variant, repoErr = vs.vr.GetByID(ctx, db, variantID)
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.ErrNotFound("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		default:
			return nil, apierr.ErrInternalError("Failed to fetch variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return variant, nil
}

func (vs *VariantService) UpdateVariant(
	ctx context.Context,
	variantID uuid.UUID,
	in UpdateVariantInput,
) error {
	fields := UpdateVariantFields{
		Price:           in.Price,
		CrossedOutPrice: in.CrossedOutPrice,
		Currency:        in.Currency,
		Attributes:      in.Attributes,
		IsDefault:       in.IsDefault,
	}

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return apierr.ErrBadRequest("Validation error").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{
					Field:   "title",
					Message: "title cannot be empty",
				})
		}
		fields.Title = &title
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		if in.IsDefault != nil && *in.IsDefault {
			existing, err := vs.vr.GetByID(ctx, tx, variantID)
			if err != nil {
				return err
			}
			if err := vs.vr.ClearDefaultFlags(ctx, tx, existing.ProductID); err != nil {
				return err
			}
		}
		return vs.vr.Update(ctx, tx, variantID, fields)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to update variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (vs *VariantService) ListVariantsByProductID(
	ctx context.Context,
	productID uuid.UUID,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[model.ProductVariant], error) {
	if productID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if q == nil {
		q = &pagination.ListQuery{}
	}

	var result *pagination.PagedResult[model.ProductVariant]
	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = vs.vr.ListByProductID(ctx, db, ListVariantOptions{
			ProductID:      productID,
			Query:          q,
			IncludeDeleted: includeDeleted,
		})
		return repoErr
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to list variants").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return result, nil
}

func (vs *VariantService) DeleteVariantByID(
	ctx context.Context,
	variantID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return apierr.ErrBadRequest("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return vs.vr.Delete(ctx, db, variantID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to delete variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (vs *VariantService) AttachMedia(
	ctx context.Context,
	in AttachMediaInput,
) (*model.VariantMedia, error) {
	if in.VariantID == uuid.Nil || in.StorageObjectID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Variant ID and Storage Object ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	mediaType := strings.TrimSpace(in.MediaType)
	if mediaType == "" {
		return nil, apierr.ErrBadRequest("Validation error").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{
				Field:   "media_type",
				Message: "media_type is required",
			})
	}

	var media *model.VariantMedia
	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		media, repoErr = vs.vr.AddMedia(ctx, db, CreateVariantMediaInput{
			VariantID:       in.VariantID,
			StorageObjectID: in.StorageObjectID,
			MediaType:       mediaType,
			SortOrder:       in.SortOrder,
		})
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, apierr.ErrConflict("This storage object is already attached to this variant").
				WithCode(errcode.CodeMediaAlreadyAttached).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.ErrBadRequest("Referenced variant or storage object does not exist").
				WithCode(apierr.CodeInvalidReference).
				Wrap(err)

		default:
			return nil, apierr.ErrInternalError("Failed to attach media to variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return media, nil
}

func (vs *VariantService) DetachMedia(
	ctx context.Context,
	variantID uuid.UUID,
	mediaID uuid.UUID,
) error {
	if variantID == uuid.Nil || mediaID == uuid.Nil {
		return apierr.ErrBadRequest("Variant ID and Media ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return vs.vr.RemoveMedia(ctx, db, variantID, mediaID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Variant media relationship not found").
				WithCode(errcode.CodeMediaNotFound).
				Wrap(err)

		default:
			return apierr.ErrInternalError("Failed to detach media from variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (vs *VariantService) ReorderMedia(
	ctx context.Context,
	variantID uuid.UUID,
	orderedMediaIDs []uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return apierr.ErrBadRequest("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if len(orderedMediaIDs) == 0 {
		return nil
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return vs.vr.ReorderMedia(ctx, tx, variantID, orderedMediaIDs)
	})

	if err != nil {
		return apierr.ErrInternalError("Failed to reorder variant media").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}

func (vs *VariantService) ListVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
) ([]*model.VariantMedia, error) {
	if variantID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var mediaList []*model.VariantMedia
	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		mediaList, repoErr = vs.vr.ListMediaByVariantID(ctx, db, variantID)
		return repoErr
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to list variant media").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	for _, m := range mediaList {
		if m.Object == nil {
			continue
		}

		if m.Object.PublicURL == "" && m.Object.Bucket != "" && m.Object.Key != "" {
			presignedGET, err := vs.minioClient.PresignedGetObject(
				ctx,
				m.Object.Bucket,
				m.Object.Key,
				1*time.Hour,
				nil,
			)
			if err == nil {
				m.Object.PublicURL = presignedGET.String()
			}
		}
	}

	return mediaList, nil
}

func (vs *VariantService) SetDefaultVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) error {
	if productID == uuid.Nil || variantID == uuid.Nil {
		return apierr.ErrBadRequest("Product ID and Variant ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		v, err := vs.vr.GetByID(ctx, tx, variantID)
		if err != nil {
			return err
		}
		if v.ProductID != productID {
			return apierr.ErrBadRequest("Variant does not belong to this product").
				WithCode(errcode.CodeVariantProductMismatch)
		}

		if err := vs.vr.ClearDefaultFlags(ctx, tx, productID); err != nil {
			return err
		}
		isDefault := true
		return vs.vr.Update(ctx, tx, variantID, UpdateVariantFields{IsDefault: &isDefault})
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)
		default:
			return err
		}
	}

	return nil
}
