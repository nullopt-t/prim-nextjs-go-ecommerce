package variant

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/errcode"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// ObjectService defines storage URL resolution methods for product and variant media.
type ObjectService interface {
	GetObjectURL(ctx context.Context, bucket, key string) string
}

// VariantService coordinates business logic, transaction handling, and cross-domain inventory integration
// for product variants (SKUs), media galleries, and stock availability.
type VariantService struct {
	logger           log.Logger
	dr               database.Runner
	objectService    ObjectService
	vr               *VariantRepository
	inventoryService *inventory.InventoryService
}

// NewService constructs a new VariantService with injected database runner, media object service,
// variant repository, and inventory service.
func NewService(
	logger log.Logger,
	r database.Runner,
	objectService ObjectService,
	vr *VariantRepository,
	inventoryService *inventory.InventoryService,
) *VariantService {
	return &VariantService{
		logger:           logger,
		dr:               r,
		objectService:    objectService,
		vr:               vr,
		inventoryService: inventoryService,
	}
}

// CreateVariantInput represents the parameters required to create a new product variant.
type CreateVariantInput struct {
	// ProductID is the UUID of the parent product.
	ProductID         uuid.UUID
	// SKU is an optional custom SKU string; if omitted, a new UUID string is generated.
	SKU               *string
	// Title is the required display title of the variant.
	Title             string
	// Price is the sales price in minor currency units (cents).
	Price             *int64
	// CrossedOutPrice is the original MSRP price in minor currency units (cents).
	CrossedOutPrice   *int64
	// Currency is the ISO-4217 currency code (e.g. USD).
	Currency          *string
	// Attributes contains multi-attribute key-value specifications (e.g. {"color":"black","size":"XL"}).
	Attributes        map[string]any
	// IsDefault indicates whether this variant is the initial storefront selection.
	IsDefault         bool
	// ThumbnailObjectID is an optional storage object UUID for the variant's thumbnail image.
	ThumbnailObjectID *uuid.UUID
}

// UpdateVariantInput represents the parameters for updating an existing product variant.
type UpdateVariantInput struct {
	SKU               *string
	Title             *string
	Price             *int64
	CrossedOutPrice   *int64
	Currency          *string
	Attributes        map[string]any
	IsDefault         *bool
	ThumbnailObjectID *uuid.UUID
}

// AttachMediaInput defines the parameters to link a storage object to a variant.
type AttachMediaInput struct {
	VariantID       uuid.UUID
	StorageObjectID uuid.UUID
	MediaType       string
	SortOrder       int
}

// CreateVariant creates a new product variant. If this is the product's first variant,
// it is automatically designated as default. If explicitly marked as default, any existing default
// flags for the product's variants are cleared first.
func (vs *VariantService) CreateVariant(
	ctx context.Context,
	in *CreateVariantInput,
) (*model.ProductVariant, error) {
	if in == nil {
		return nil, apierr.BadRequestError("Request body is required").
			WithCode(apierr.CodeInvalidPayload)
	}

	if in.ProductID == uuid.Nil {
		return nil, apierr.BadRequestError("Validation error").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{
				Field:   "product_id",
				Message: "product_id is required",
			})
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, apierr.BadRequestError("Validation error").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{
				Field:   "title",
				Message: "title cannot be empty",
			})
	}

	sku := uuid.NewString()
	if in.SKU != nil {
		trimmedSKU := strings.TrimSpace(*in.SKU)
		if trimmedSKU != "" {
			sku = trimmedSKU
		}
	}

	currency := "EGP"
	if in.Currency != nil && strings.TrimSpace(*in.Currency) != "" {
		currency = strings.TrimSpace(*in.Currency)
	}

	variant := &model.ProductVariant{
		ID:                uuid.New(),
		SKU:               sku,
		ProductID:         in.ProductID,
		Title:             title,
		Price:             in.Price,
		CrossedOutPrice:   in.CrossedOutPrice,
		Currency:          &currency,
		Attributes:        in.Attributes,
		IsDefault:         in.IsDefault,
		ThumbnailObjectID: in.ThumbnailObjectID,
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		// Acquire an exclusive row-level lock on the parent product to prevent concurrent creation races
		var parentProductID uuid.UUID
		err := tx.QueryRow(ctx, `SELECT id FROM products WHERE id = $1 FOR UPDATE`, variant.ProductID).Scan(&parentProductID)
		if err != nil {
			return err
		}

		// If this is the first variant of the product, make it default automatically
		count, err := vs.vr.CountByProductID(ctx, tx, variant.ProductID)
		if err != nil {
			return err
		}
		if count == 0 {
			variant.IsDefault = true
		}

		if variant.IsDefault && count > 0 {
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
			return nil, apierr.BadRequestError("Referenced product or storage object does not exist").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err).
				WithFields(api.FieldError{
					Field:   "product_id",
					Message: "product reference is invalid",
				})

		case errors.Is(mappedErr, database.ErrConflict):
			return nil, apierr.ConflictError("Variant with this SKU already exists").
				WithCode(errcode.CodeVariantAlreadyExists).
				Wrap(err)

		default:
			return nil, apierr.InternalServerError("Failed to create variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return variant, nil
}

// GetVariantByID retrieves a variant by UUID and populates presigned thumbnail and media gallery URLs.
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
			return nil, apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		default:
			return nil, apierr.InternalServerError("Failed to fetch variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	if variant.Thumbnail != nil && vs.objectService != nil {
		variant.Thumbnail.PublicURL = vs.objectService.GetObjectURL(
			ctx,
			variant.Thumbnail.Bucket,
			variant.Thumbnail.Key,
		)
	}

	mediaList, err := vs.ListVariantMedia(ctx, variantID)
	if err == nil {
		variant.Media = mediaList
	}

	return variant, nil
}

// GetVariantBySKU retrieves a variant by its unique SKU identifier.
func (vs *VariantService) GetVariantBySKU(
	ctx context.Context,
	sku string,
) (*model.ProductVariant, error) {
	cleanSKU := strings.TrimSpace(sku)
	if cleanSKU == "" {
		return nil, apierr.BadRequestError("SKU is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var variant *model.ProductVariant
	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		v, err := vs.vr.GetBySKU(ctx, db, cleanSKU)
		if err != nil {
			return err
		}
		variant = v
		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		default:
			return nil, apierr.InternalServerError("Failed to fetch variant by SKU").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	if variant.Thumbnail != nil && vs.objectService != nil {
		variant.Thumbnail.PublicURL = vs.objectService.GetObjectURL(
			ctx,
			variant.Thumbnail.Bucket,
			variant.Thumbnail.Key,
		)
	}

	mediaList, err := vs.ListVariantMedia(ctx, variant.ID)
	if err == nil {
		variant.Media = mediaList
	}

	return variant, nil
}

// UpdateVariant updates mutable variant fields. If is_default is set to true, existing default
// flags for the product's variants are cleared first. If an existing default variant is unset,
// the first available variant is automatically reassigned as default.
func (vs *VariantService) UpdateVariant(
	ctx context.Context,
	variantID uuid.UUID,
	in UpdateVariantInput,
) error {
	fields := UpdateVariantFields{
		Price:             in.Price,
		CrossedOutPrice:   in.CrossedOutPrice,
		Currency:          in.Currency,
		Attributes:        in.Attributes,
		IsDefault:         in.IsDefault,
		ThumbnailObjectID: in.ThumbnailObjectID,
	}

	if in.SKU != nil {
		sku := strings.TrimSpace(*in.SKU)
		if sku == "" {
			return apierr.BadRequestError("Validation error").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{
					Field:   "sku",
					Message: "sku cannot be empty",
				})
		}
		fields.SKU = &sku
	}

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return apierr.BadRequestError("Validation error").
				WithCode(apierr.CodeValidationFailed).
				WithFields(api.FieldError{
					Field:   "title",
					Message: "title cannot be empty",
				})
		}
		fields.Title = &title
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		existing, err := vs.vr.GetByID(ctx, tx, variantID)
		if err != nil {
			return err
		}

		if in.IsDefault != nil && *in.IsDefault {
			if err := vs.vr.ClearDefaultFlags(ctx, tx, existing.ProductID); err != nil {
				return err
			}
		}

		if err := vs.vr.Update(ctx, tx, variantID, fields); err != nil {
			return err
		}

		// If this variant was default and is now unset, ensure another variant is designated as default
		if in.IsDefault != nil && !*in.IsDefault && existing.IsDefault {
			if err := vs.vr.SetFirstAvailableAsDefault(ctx, tx, existing.ProductID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		case errors.Is(mappedErr, database.ErrConflict):
			return apierr.ConflictError("Variant with this SKU already exists").
				WithCode(errcode.CodeVariantAlreadyExists).
				Wrap(err)

		default:
			return apierr.InternalServerError("Failed to update variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// ListVariantsByProductID returns a paginated list of variants belonging to a specific product,
// populating thumbnail and media gallery presigned URLs.
func (vs *VariantService) ListVariantsByProductID(
	ctx context.Context,
	productID uuid.UUID,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[model.ProductVariant], error) {
	if productID == uuid.Nil {
		return nil, apierr.BadRequestError("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if q == nil {
		q = &pagination.ListQuery{}
	}

	var result *pagination.PagedResult[model.ProductVariant]
	var mediaMap map[uuid.UUID][]*model.VariantMedia

	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = vs.vr.ListByProductID(ctx, db, ListVariantOptions{
			ProductID:      productID,
			Query:          q,
			IncludeDeleted: includeDeleted,
		})
		if repoErr != nil {
			return repoErr
		}

		if len(result.Items) > 0 {
			variantIDs := make([]uuid.UUID, 0, len(result.Items))
			for _, item := range result.Items {
				if item != nil {
					variantIDs = append(variantIDs, item.ID)
				}
			}
			mediaMap, repoErr = vs.vr.ListMediaByVariantIDs(ctx, db, variantIDs)
			if repoErr != nil {
				return repoErr
			}
		}

		return nil
	})

	if err != nil {
		return nil, apierr.InternalServerError("Failed to list variants").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	for _, v := range result.Items {
		if v == nil {
			continue
		}

		if v.Thumbnail != nil && vs.objectService != nil {
			v.Thumbnail.PublicURL = vs.objectService.GetObjectURL(
				ctx,
				v.Thumbnail.Bucket,
				v.Thumbnail.Key,
			)
		}

		if mediaMap != nil {
			if list, ok := mediaMap[v.ID]; ok {
				for _, m := range list {
					if m.Object != nil && vs.objectService != nil {
						m.Object.PublicURL = vs.objectService.GetObjectURL(
							ctx,
							m.Object.Bucket,
							m.Object.Key,
						)
					}
				}
				v.Media = list
			}
		}
	}

	return result, nil
}

// DeleteVariantByID soft-deletes a variant, releasing any active inventory reservations.
// If the deleted variant was the default, the oldest remaining active variant is promoted to default.
func (vs *VariantService) DeleteVariantByID(
	ctx context.Context,
	variantID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		variant, err := vs.vr.GetByID(ctx, tx, variantID)
		if err != nil {
			return err
		}

		if err := vs.vr.Delete(ctx, tx, variantID); err != nil {
			return err
		}

		if variant.IsDefault {
			if err := vs.vr.SetFirstAvailableAsDefault(ctx, tx, variant.ProductID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		default:
			return apierr.InternalServerError("Failed to delete variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// RestoreVariantByID un-deletes a previously soft-deleted product variant.
// If another default variant already exists for the product, the restored variant has is_default reset to false
// to preserve the partial unique index idx_product_variants_default.
func (vs *VariantService) RestoreVariantByID(
	ctx context.Context,
	variantID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		// Fetch the variant including deleted records to inspect its product and default flag
		var (
			productID uuid.UUID
			isDefault bool
		)
		err := tx.QueryRow(ctx,
			`SELECT product_id, is_default FROM product_variants WHERE id = $1 AND deleted_at IS NOT NULL`,
			variantID,
		).Scan(&productID, &isDefault)
		if err != nil {
			return err
		}

		// Lock parent product
		var pID uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT id FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&pID); err != nil {
			return err
		}

		if isDefault {
			// Check if another active variant is currently designated as default
			var defaultExists bool
			err = tx.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM product_variants WHERE product_id = $1 AND is_default = true AND deleted_at IS NULL)`,
				productID,
			).Scan(&defaultExists)
			if err != nil {
				return err
			}

			// If another default already exists, reset is_default to false on the variant being restored
			if defaultExists {
				_, err = tx.Exec(ctx, `UPDATE product_variants SET is_default = false WHERE id = $1`, variantID)
				if err != nil {
					return err
				}
			}
		}

		return vs.vr.Restore(ctx, tx, variantID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Variant not found or not deleted").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		default:
			return apierr.InternalServerError("Failed to restore variant").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// AttachMedia associates an uploaded storage object with a variant.
func (vs *VariantService) AttachMedia(
	ctx context.Context,
	in AttachMediaInput,
) (*model.VariantMedia, error) {
	if in.VariantID == uuid.Nil || in.StorageObjectID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID and Storage Object ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	in.MediaType = strings.TrimSpace(in.MediaType)
	if in.MediaType == "" {
		in.MediaType = "image"
	}

	mediaID := uuid.New()
	var createdMedia *model.VariantMedia

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		var repoErr error
		createdMedia, repoErr = vs.vr.AddMedia(ctx, tx, CreateVariantMediaInput{
			ID:        mediaID,
			VariantID: in.VariantID,
			ObjectID:  in.StorageObjectID,
			MediaType: in.MediaType,
			SortOrder: in.SortOrder,
		})
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.BadRequestError("Referenced variant or storage object does not exist").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)

		default:
			return nil, apierr.InternalServerError("Failed to attach media").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return createdMedia, nil
}

// RemoveMedia disassociates a media item from a variant and resequences remaining media sort orders.
func (vs *VariantService) RemoveMedia(
	ctx context.Context,
	variantID uuid.UUID,
	mediaID uuid.UUID,
) error {
	if variantID == uuid.Nil || mediaID == uuid.Nil {
		return apierr.BadRequestError("Variant ID and Media ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return vs.vr.RemoveMedia(ctx, db, variantID, mediaID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Variant media not found").
				WithCode(errcode.CodeMediaNotFound).
				Wrap(err)

		default:
			return apierr.InternalServerError("Failed to remove variant media").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

// DetachMedia is an alias for RemoveMedia.
func (vs *VariantService) DetachMedia(
	ctx context.Context,
	variantID uuid.UUID,
	mediaID uuid.UUID,
) error {
	return vs.RemoveMedia(ctx, variantID, mediaID)
}

// ReorderMedia applies a new sequence order to all media items attached to a variant.
func (vs *VariantService) ReorderMedia(
	ctx context.Context,
	variantID uuid.UUID,
	orderedMediaIDs []uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if len(orderedMediaIDs) == 0 {
		return apierr.BadRequestError("Ordered media IDs cannot be empty").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		existingMedia, err := vs.vr.ListMediaByVariantID(ctx, tx, variantID)
		if err != nil {
			return err
		}

		if len(existingMedia) != len(orderedMediaIDs) {
			return apierr.BadRequestError("Ordered media list must contain all variant media items").
				WithCode(apierr.CodeInvalidInput)
		}

		existingMap := make(map[uuid.UUID]bool, len(existingMedia))
		for _, m := range existingMedia {
			existingMap[m.ID] = true
		}

		for _, id := range orderedMediaIDs {
			if !existingMap[id] {
				return apierr.BadRequestError("Media ID does not belong to this variant").
					WithCode(apierr.CodeInvalidInput)
			}
		}

		return vs.vr.ReorderMedia(ctx, tx, variantID, orderedMediaIDs)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.InternalServerError("Failed to reorder media").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}

// ListVariantMedia retrieves all gallery media items for a variant and populates presigned CDN URLs.
func (vs *VariantService) ListVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
) ([]*model.VariantMedia, error) {
	if variantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var mediaList []*model.VariantMedia
	err := vs.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		mediaList, repoErr = vs.vr.ListMediaByVariantID(ctx, db, variantID)
		return repoErr
	})

	if err != nil {
		return nil, apierr.InternalServerError("Failed to list variant media").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	for _, m := range mediaList {
		if m.Object == nil {
			continue
		}

		if m.Object.PublicURL == "" && m.Object.Bucket != "" && m.Object.Key != "" && vs.objectService != nil {
			url := vs.objectService.GetObjectURL(
				ctx,
				m.Object.Bucket,
				m.Object.Key,
			)
			if url != "" {
				m.Object.PublicURL = url
			}
		}
	}

	return mediaList, nil
}

// SetDefaultVariant designates a specific variant as the default for its parent product,
// clearing default flags on all other variants of the product.
func (vs *VariantService) SetDefaultVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) error {
	if productID == uuid.Nil || variantID == uuid.Nil {
		return apierr.BadRequestError("Product ID and Variant ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		v, err := vs.vr.GetByID(ctx, tx, variantID)
		if err != nil {
			return err
		}
		if v.ProductID != productID {
			return apierr.BadRequestError("Variant does not belong to this product").
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
			return apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)
		default:
			return err
		}
	}

	return nil
}

// AdjustStock delegates stock adjustment to the injected inventory service.
func (vs *VariantService) AdjustStock(
	ctx context.Context,
	in inventory.AdjustStockInput,
) (*model.InventoryStock, error) {
	if vs.inventoryService == nil {
		return nil, apierr.InternalServerError("inventory service is not configured")
	}
	return vs.inventoryService.AdjustStock(ctx, in)
}

// GetStockByVariantID delegates stock balance lookup to the injected inventory service.
func (vs *VariantService) GetStockByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) (*model.InventoryStock, error) {
	if vs.inventoryService == nil {
		return nil, apierr.InternalServerError("inventory service is not configured")
	}
	return vs.inventoryService.GetStockByVariantID(ctx, variantID)
}

// ListLedgers delegates inventory audit ledger queries to the injected inventory service.
func (vs *VariantService) ListLedgers(
	ctx context.Context,
	variantID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.InventoryLedger], error) {
	if vs.inventoryService == nil {
		return nil, apierr.InternalServerError("inventory service is not configured")
	}
	return vs.inventoryService.ListLedgers(ctx, variantID, q)
}

