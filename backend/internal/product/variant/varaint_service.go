package variant

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
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_PAYLOAD").
			WithMessage("Request body is required")
	}

	if in.ProductID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("VALIDATION_FAILED").
			WithMessage("Validation error").
			WithFields(api.FieldError{
				Field:   "product_id",
				Message: "product_id is required",
			})
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("VALIDATION_FAILED").
			WithMessage("Validation error").
			WithFields(api.FieldError{
				Field:   "title",
				Message: "title cannot be empty",
			})
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	variant := &model.ProductVariant{
		ID:              uuid.New(),
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

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		if variant.IsDefault {
			if err := vs.vr.ClearDefaultFlags(ctx, tx, variant.ProductID); err != nil {
				return err
			}
		}
		return vs.vr.Create(ctx, tx, variant)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, security.NewSecureError(http.StatusBadRequest).
				WithCode("PRODUCT_NOT_FOUND").
				WithMessage("Referenced product does not exist").
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "product_id",
					Message: "product reference is invalid",
				})

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to create variant").
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
	if variantID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Variant ID is required")
	}

	var variant *model.ProductVariant
	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		variant, repoErr = vs.vr.Get(ctx, db, &Filter{ID: &variantID})
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, security.NewSecureError(http.StatusNotFound).
				WithCode("VARIANT_NOT_FOUND").
				WithMessage("Variant not found").
				Wrap(err)

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to fetch variant").
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
	if variantID == uuid.Nil {
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Variant ID is required")
	}

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
			return security.NewSecureError(http.StatusBadRequest).
				WithCode("VALIDATION_FAILED").
				WithMessage("Validation error").
				WithFields(api.FieldError{
					Field:   "title",
					Message: "title cannot be empty",
				})
		}
		fields.Title = &title
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		if in.IsDefault != nil && *in.IsDefault {
			existing, err := vs.vr.Get(ctx, tx, &Filter{ID: &variantID})
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
			return security.NewSecureError(http.StatusNotFound).
				WithCode("VARIANT_NOT_FOUND").
				WithMessage("Variant not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to update variant").
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (vs *VariantService) ListVariantsByProductID(
	ctx context.Context,
	productID uuid.UUID,
	q *api.ListQuery,
	includeDeleted bool,
) (*api.PagedResult[model.ProductVariant], error) {
	if productID == uuid.Nil {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Product ID is required")
	}

	if q == nil {
		q = &api.ListQuery{}
	}

	var result *api.PagedResult[model.ProductVariant]
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
		return nil, security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to list variants").
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
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Variant ID is required")
	}

	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return vs.vr.Delete(ctx, db, variantID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("VARIANT_NOT_FOUND").
				WithMessage("Variant not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to delete variant").
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
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Variant ID and Storage Object ID are required")
	}

	mediaType := strings.TrimSpace(in.MediaType)
	if mediaType == "" {
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("VALIDATION_FAILED").
			WithMessage("Validation error").
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
			return nil, security.NewSecureError(http.StatusConflict).
				WithCode("MEDIA_ALREADY_ATTACHED").
				WithMessage("This storage object is already attached to this variant").
				Wrap(err)

		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, security.NewSecureError(http.StatusBadRequest).
				WithCode("INVALID_REFERENCE").
				WithMessage("Referenced variant or storage object does not exist").
				Wrap(err)

		default:
			return nil, security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to attach media to variant").
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
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Variant ID and Media ID are required")
	}

	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return vs.vr.RemoveMedia(ctx, db, variantID, mediaID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return security.NewSecureError(http.StatusNotFound).
				WithCode("MEDIA_NOT_FOUND").
				WithMessage("Variant media relationship not found").
				Wrap(err)

		default:
			return security.NewSecureError(http.StatusInternalServerError).
				WithCode("INTERNAL_ERROR").
				WithMessage("Failed to detach media from variant").
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
		return security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Variant ID is required")
	}

	if len(orderedMediaIDs) == 0 {
		return nil
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return vs.vr.ReorderMedia(ctx, tx, variantID, orderedMediaIDs)
	})

	if err != nil {
		return security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to reorder variant media").
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
		return nil, security.NewSecureError(http.StatusBadRequest).
			WithCode("INVALID_INPUT").
			WithMessage("Variant ID is required")
	}

	var mediaList []*model.VariantMedia
	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		mediaList, repoErr = vs.vr.ListMediaByVariantID(ctx, db, variantID)
		return repoErr
	})

	if err != nil {
		return nil, security.NewSecureError(http.StatusInternalServerError).
			WithCode("INTERNAL_ERROR").
			WithMessage("Failed to list variant media").
			Wrap(err).
			WithStack()
	}

	for _, m := range mediaList {
		if m.Object == nil {
			continue
		}

		// Fallback to presigned URL if PublicURL is not set
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
