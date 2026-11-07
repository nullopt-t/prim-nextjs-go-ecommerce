package variant

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/category/attribute"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	fileUtil "github.com/nullopt-t/prim-backend/pkg/file"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// ObjectRepository defines database persistence methods for media storage records.
type ObjectRepository interface {
	Create(ctx context.Context, object *model.Object) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Object, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Object, error)
	MarkDeletingByKey(ctx context.Context, bucket, key string) error
	MarkDeletedByKey(ctx context.Context, bucket, key string) error
}

// InventoryReader is the read-only subset of inventory persistence the
// variant domain needs for stock projections and ledger history.
type InventoryReader interface {
	GetStock(ctx context.Context, variantID uuid.UUID) (*model.InventoryStock, error)
	GetStockForVariants(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID]*model.InventoryStock, error)
	ListVariantLedgers(ctx context.Context, variantID uuid.UUID, q *pagination.ListQuery) (*pagination.PagedResult[model.InventoryLedger], error)
}

// InventoryWriter owns stock mutation invariants: appending ledger entries
// and transitioning holds. Write rules (locking, on-hand vs reserved checks)
// live here; the caller decides the transaction boundary — existing ctx
// joins the ambient transaction, standalone calls start their own.
type InventoryWriter interface {
	// AdjustStock records one adjustment as a standalone use case.
	AdjustStock(ctx context.Context, in inventory.AdjustStockInput) (*model.InventoryStock, error)
	// ApplyStockAdjustment appends a signed ledger entry participating in the
	// caller's transaction, committing atomically with the owning aggregate.
	ApplyStockAdjustment(ctx context.Context, in inventory.AdjustStockInput) error
	// CommitReservationsForOrder turns an order's holds into sale deductions.
	CommitReservationsForOrder(ctx context.Context, orderID uuid.UUID) error
	// ReleaseReservationsForOrder releases an order's holds, reporting the count.
	ReleaseReservationsForOrder(ctx context.Context, orderID uuid.UUID) (int64, error)
	// ExpiredReservationOrderIDs lists orders whose checkout hold has elapsed.
	ExpiredReservationOrderIDs(ctx context.Context) ([]uuid.UUID, error)
}

// errSimpleProductLocked signals that a lifecycle change was rejected because the
// parent product is a simple product whose default variant is fixed.
var errSimpleProductLocked = errors.New("simple product variant is locked")

// errVariationAxesRequired signals that a second child cannot be created until
// the parent product declares which attributes vary.
var errVariationAxesRequired = errors.New("variation axes are required for multiple variants")

// variationAxesRequiredError translates errVariationAxesRequired into a structured API error.
func variationAxesRequiredError() error {
	return apierr.ConflictError("Set variation axes before adding another variant. Products without variation axes have a single child.").
		WithCode(errcode.CodeVariationAxesRequired)
}

// simpleProductLockedError translates errSimpleProductLocked into a structured API error.
func simpleProductLockedError() error {
	return apierr.ConflictError("Products without variation axes have a single fixed child. Set variation axes to manage multiple variants.").
		WithCode(errcode.CodeSimpleProductVariantLocked)
}

// variationValidationError builds a field-scoped 400 for invalid variation data.
func variationValidationError(code, field, message string) error {
	return apierr.BadRequestError(message).
		WithCode(code).
		WithFields(api.FieldError{Field: field, Message: message})
}

// variationValueString converts a JSON scalar into a trimmed string. Arrays,
// objects, and empty values are rejected because variation axes must be scalar.
func variationValueString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		s := strings.TrimSpace(v)
		return s, s != ""
	case json.Number:
		s := strings.TrimSpace(v.String())
		return s, s != ""
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), true
	case int:
		return strconv.Itoa(v), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint:
		return strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	case bool:
		return strconv.FormatBool(v), true
	default:
		return "", false
	}
}

// canonicalAllowedValues decodes a category attribute's allowed_values JSON
// array into comparable strings.
func canonicalAllowedValues(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var values []any
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		if s, ok := variationValueString(v); ok {
			out = append(out, s)
		}
	}
	return out
}

// validateVariantAttributes enforces that a variant's attributes exactly match
// the parent product's chosen variation axes, that each value is non-empty and
// allowed by the category attribute definition (when one exists), and that the
// resulting combination is unique among the product's active children.
func (vs *VariantService) validateVariantAttributes(
	ctx context.Context,
	productID uuid.UUID,
	attrs map[string]any,
	forUpdate bool,
	excludeVariantID *uuid.UUID,
) (map[string]any, *ProductVariationContext, error) {
	context, err := vs.vr.GetVariationContext(ctx, productID, forUpdate)
	if err != nil {
		return nil, nil, err
	}
	if attrs == nil {
		attrs = map[string]any{}
	}
	if len(context.VariationAxes) == 0 {
		return attrs, context, nil
	}

	defs, err := attribute.NewRepository().ListByCategory(ctx, context.CategoryID)
	if err != nil {
		return nil, nil, err
	}
	defByKey := make(map[string]attribute.AttributeItem, len(defs))
	for _, def := range defs {
		defByKey[def.Key] = def
	}

	axisSet := make(map[string]struct{}, len(context.VariationAxes))
	for _, axis := range context.VariationAxes {
		axisSet[axis] = struct{}{}
	}

	normalized := make(map[string]any, len(context.VariationAxes))
	for _, axis := range context.VariationAxes {
		raw, ok := attrs[axis]
		if !ok {
			return nil, nil, variationValidationError(
				errcode.CodeVariationAxisMismatch,
				"attributes."+axis,
				"Missing value for variation axis "+axis,
			)
		}
		value, ok := variationValueString(raw)
		if !ok {
			return nil, nil, variationValidationError(
				errcode.CodeVariationValueInvalid,
				"attributes."+axis,
				"Variation value for "+axis+" must be a non-empty scalar",
			)
		}

		if def, exists := defByKey[axis]; exists {
			allowed := canonicalAllowedValues(def.AllowedValues)
			if len(allowed) > 0 {
				canonical := ""
				for _, candidate := range allowed {
					if strings.EqualFold(candidate, value) {
						canonical = candidate
						break
					}
				}
				if canonical == "" {
					return nil, nil, variationValidationError(
						errcode.CodeVariationValueInvalid,
						"attributes."+axis,
						"Value "+value+" is not allowed for "+axis,
					)
				}
				value = canonical
			}
		}
		normalized[axis] = value
	}

	for key := range attrs {
		if _, ok := axisSet[key]; !ok {
			return nil, nil, variationValidationError(
				errcode.CodeVariationAxisMismatch,
				"attributes."+key,
				"Attribute "+key+" is not a variation axis for this product",
			)
		}
	}

	exists, err := vs.vr.CombinationExists(ctx, productID, normalized, excludeVariantID)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, apierr.ConflictError("A variant with this variation combination already exists").
			WithCode(errcode.CodeVariantCombinationExists).
			WithFields(api.FieldError{
				Field:   "attributes",
				Message: "combination already exists",
			})
	}

	return normalized, context, nil
}

// buildVariantTitle derives a readable child title from the parent title when
// the caller does not supply one (for example, matrix bulk creation).
func buildVariantTitle(productTitle string, axes []string, attrs map[string]any) string {
	parts := make([]string, 0, len(axes))
	for _, axis := range axes {
		if v, ok := attrs[axis]; ok {
			if s, ok := variationValueString(v); ok {
				parts = append(parts, s)
			}
		}
	}
	if len(parts) == 0 {
		return productTitle
	}
	return productTitle + " - " + strings.Join(parts, ", ")
}

// VariantService coordinates business logic, transaction handling, and cross-domain inventory integration
// for product variants (SKUs), media galleries, and stock availability.
type VariantService struct {
	logger          log.Logger
	dr              database.Runner
	objectRepo      ObjectRepository
	storageProvider storage.StorageProvider
	vr              *VariantRepository
	inventoryReader InventoryReader
	inventory       InventoryWriter
}

// NewService constructs a new VariantService with injected database runner, media object repository,
// storage provider, variant repository, inventory read repository, and the inventory writer that
// owns stock mutation invariants.
func NewService(
	logger log.Logger,
	r database.Runner,
	objectRepo ObjectRepository,
	storageProvider storage.StorageProvider,
	vr *VariantRepository,
	inventoryReader InventoryReader,
	inventoryWriter InventoryWriter,
) *VariantService {
	if logger == nil {
		panic("variant: logger is required")
	}
	if r == nil {
		panic("variant: database runner is required")
	}
	if objectRepo == nil {
		panic("variant: object repository is required")
	}
	if storageProvider == nil {
		panic("variant: storage provider is required")
	}
	if vr == nil {
		panic("variant: variant repository is required")
	}
	if inventoryReader == nil {
		panic("variant: inventory reader is required")
	}
	if inventoryWriter == nil {
		panic("variant: inventory writer is required")
	}
	return &VariantService{
		logger:          logger,
		dr:              r,
		objectRepo:      objectRepo,
		storageProvider: storageProvider,
		vr:              vr,
		inventoryReader: inventoryReader,
		inventory:       inventoryWriter,
	}
}

// handleError maps repository errors to standard API errors.
func (vs *VariantService) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError("Variant not found").
			WithCode(errcode.CodeVariantNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Variant combination or SKU already exists").
			WithCode(errcode.CodeVariantCombinationExists).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced record does not exist").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

// CreateVariantInput represents the parameters required to create a new product variant.
type CreateVariantInput struct {
	// ProductID is the UUID of the parent product.
	ProductID uuid.UUID
	// SKU is an optional custom SKU string; if omitted, a new UUID string is generated.
	SKU *string
	// Title is the required display title of the variant.
	Title string
	// Price is the sales price in minor currency units (cents).
	Price *int64
	// CrossedOutPrice is the original MSRP price in minor currency units (cents).
	CrossedOutPrice *int64
	// Currency is the ISO-4217 currency code (e.g. USD).
	Currency *string
	// Attributes contains multi-attribute key-value specifications (e.g. {"color":"black","size":"XL"}).
	Attributes map[string]any
	// IsDefault indicates whether this variant is the initial storefront selection.
	IsDefault bool
	// ThumbnailObjectID is an optional storage object UUID for the variant's thumbnail image.
	ThumbnailObjectID *uuid.UUID
}

// BatchCreateVariantInput represents one child row in a variation-matrix bulk
// creation request.
type BatchCreateVariantInput struct {
	SKU             *string
	Title           *string
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	Attributes      map[string]any
	IsDefault       *bool
	InitialStock    *int
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

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		// Validate against the parent's chosen variation axes while holding an
		// exclusive row lock, so two concurrent requests cannot create the same
		// combination.
		normalizedAttrs, productCtx, err := vs.validateVariantAttributes(
			ctx, variant.ProductID, in.Attributes, true, nil,
		)
		if err != nil {
			return err
		}
		variant.Attributes = normalizedAttrs
		if strings.TrimSpace(variant.Title) == "" {
			variant.Title = buildVariantTitle(productCtx.Title, productCtx.VariationAxes, normalizedAttrs)
		}

		// If this is the first variant of the product, make it default automatically
		count, err := vs.vr.CountByProductID(ctx, variant.ProductID)
		if err != nil {
			return err
		}

		// A product can have multiple children only when it declares a
		// variation theme. This replaces the old product_type lock.
		if len(productCtx.VariationAxes) == 0 && count > 0 {
			return errVariationAxesRequired
		}

		if count == 0 {
			variant.IsDefault = true
		}

		if err := vs.vr.Create(ctx, variant); err != nil {
			return err
		}

		// If the variant carries an image, the product carries it too.
		if variant.ThumbnailObjectID != nil {
			if err := vs.vr.SetProductThumbnailIfMissing(ctx, variant.ProductID, *variant.ThumbnailObjectID); err != nil {
				return err
			}
		}

		if variant.IsDefault {
			if err := vs.vr.SetDefaultVariant(ctx, variant.ProductID, variant.ID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		if errors.Is(err, errVariationAxesRequired) {
			return nil, variationAxesRequiredError()
		}
		if errors.Is(err, errSimpleProductLocked) {
			return nil, simpleProductLockedError()
		}

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

// CreateVariantsBatch creates a full variation matrix in a single transaction.
// Every row is validated against the product's axes, so a partially-created
// matrix is never left behind.
func (vs *VariantService) CreateVariantsBatch(
	ctx context.Context,
	productID uuid.UUID,
	inputs []BatchCreateVariantInput,
) ([]*model.ProductVariant, error) {
	if productID == uuid.Nil {
		return nil, apierr.BadRequestError("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}
	if len(inputs) == 0 {
		return nil, apierr.BadRequestError("At least one variant is required").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "variants", Message: "variants cannot be empty"})
	}
	if len(inputs) > 200 {
		return nil, apierr.BadRequestError("Too many variants in one request").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "variants", Message: "maximum 200 variants per request"})
	}

	created := make([]*model.ProductVariant, 0, len(inputs))

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		productCtx, err := vs.vr.GetVariationContext(ctx, productID, true)
		if err != nil {
			return err
		}

		count, err := vs.vr.CountByProductID(ctx, productID)
		if err != nil {
			return err
		}

		for i, in := range inputs {
			normalized, _, err := vs.validateVariantAttributes(
				ctx, productID, in.Attributes, true, nil,
			)
			if err != nil {
				return err
			}

			if len(productCtx.VariationAxes) == 0 && count > 0 {
				return errVariationAxesRequired
			}

			title := ""
			if in.Title != nil {
				title = strings.TrimSpace(*in.Title)
			}
			if title == "" {
				title = buildVariantTitle(productCtx.Title, productCtx.VariationAxes, normalized)
			}

			sku := uuid.NewString()
			if in.SKU != nil && strings.TrimSpace(*in.SKU) != "" {
				sku = strings.TrimSpace(*in.SKU)
			}

			currency := "EGP"
			if in.Currency != nil && strings.TrimSpace(*in.Currency) != "" {
				currency = strings.TrimSpace(*in.Currency)
			}

			isDefault := count == 0 && i == 0
			if in.IsDefault != nil && *in.IsDefault {
				isDefault = true
			}

			variant := &model.ProductVariant{
				ID:              uuid.New(),
				SKU:             sku,
				ProductID:       productID,
				Title:           title,
				Price:           in.Price,
				CrossedOutPrice: in.CrossedOutPrice,
				Currency:        &currency,
				Attributes:      normalized,
				IsDefault:       isDefault,
			}

			if err := vs.vr.Create(ctx, variant); err != nil {
				return err
			}

			if isDefault {
				if err := vs.vr.SetDefaultVariant(ctx, productID, variant.ID); err != nil {
					return err
				}
			}

			if in.InitialStock != nil && *in.InitialStock > 0 {
				adjustment := inventory.AdjustStockInput{
					VariantID: variant.ID,
					Quantity:  *in.InitialStock,
					Reason:    string(model.InventoryReasonRestock),
				}
				if err := vs.inventory.ApplyStockAdjustment(ctx, adjustment); err != nil {
					return err
				}
			}

			created = append(created, variant)
			count++
		}
		return nil
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		if errors.Is(err, errVariationAxesRequired) {
			return nil, variationAxesRequiredError()
		}

		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.BadRequestError("Referenced product does not exist").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, apierr.ConflictError("A variant SKU or combination already exists").
				WithCode(errcode.CodeVariantCombinationExists).
				Wrap(err)
		default:
			return nil, apierr.InternalServerError("Failed to create variants").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return created, nil
}

// GetVariantByID retrieves a single variant by its unique identifier.
func (vs *VariantService) GetVariantByID(
	ctx context.Context,
	variantID uuid.UUID,
) (*model.ProductVariant, error) {
	var variant *model.ProductVariant
	err := vs.dr.WithExecutor(ctx, func(ctx context.Context) error {
		var repoErr error
		variant, repoErr = vs.getVariantByID(ctx, variantID)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (vs *VariantService) getVariantByID(
	ctx context.Context,
	variantID uuid.UUID,
) (*model.ProductVariant, error) {
	variant, err := vs.vr.GetByID(ctx, variantID)
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

	if variant.Thumbnail != nil {
		variant.Thumbnail.PublicURL = vs.storageProvider.GetPublicURL(
			variant.Thumbnail.Bucket,
			variant.Thumbnail.Key,
		)
	}

	mediaList, mediaErr := vs.listVariantMedia(ctx, variantID)
	if mediaErr == nil {
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
	err := vs.dr.WithExecutor(ctx, func(ctx context.Context) error {
		v, err := vs.vr.GetBySKU(ctx, cleanSKU)
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

	if variant.Thumbnail != nil {
		variant.Thumbnail.PublicURL = vs.storageProvider.GetPublicURL(
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

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		existing, err := vs.vr.GetByID(ctx, variantID)
		if err != nil {
			return err
		}

		productCtx, err := vs.vr.GetVariationContext(ctx, existing.ProductID, true)
		if err != nil {
			return err
		}

		if in.Attributes != nil {
			normalized, _, err := vs.validateVariantAttributes(
				ctx, existing.ProductID, in.Attributes, true, &variantID,
			)
			if err != nil {
				return err
			}
			fields.Attributes = normalized
		}

		if in.IsDefault != nil {
			// A product without variation axes has a single child, so its
			// default selection is fixed. Only no-op writes are allowed.
			if len(productCtx.VariationAxes) == 0 && *in.IsDefault != existing.IsDefault {
				return errSimpleProductLocked
			}
		}

		if err := vs.vr.Update(ctx, variantID, fields); err != nil {
			return err
		}

		// If the variant carries an image, the product carries it too.
		if fields.ThumbnailObjectID != nil {
			if err := vs.vr.SetProductThumbnailIfMissing(ctx, existing.ProductID, *fields.ThumbnailObjectID); err != nil {
				return err
			}
		}

		if in.IsDefault != nil {
			if *in.IsDefault {
				if err := vs.vr.SetDefaultVariant(ctx, existing.ProductID, variantID); err != nil {
					return err
				}
			} else if existing.IsDefault {
				// If this variant was default and is now unset, ensure another variant is designated as default
				if err := vs.vr.SetFirstAvailableAsDefault(ctx, existing.ProductID); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		if errors.Is(err, errSimpleProductLocked) {
			return simpleProductLockedError()
		}

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

	err := vs.dr.WithExecutor(ctx, func(ctx context.Context) error {
		var repoErr error
		result, repoErr = vs.vr.ListByProductID(ctx, ListVariantOptions{
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
			mediaMap, repoErr = vs.vr.ListMediaByVariantIDs(ctx, variantIDs)
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

		if v.Thumbnail != nil && vs.storageProvider != nil {
			v.Thumbnail.PublicURL = vs.storageProvider.GetPublicURL(
				v.Thumbnail.Bucket,
				v.Thumbnail.Key,
			)
		}

		if mediaMap != nil {
			if list, ok := mediaMap[v.ID]; ok {
				for _, m := range list {
					if m.Object != nil && vs.storageProvider != nil {
						m.Object.PublicURL = vs.storageProvider.GetPublicURL(
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

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		variant, err := vs.vr.GetByID(ctx, variantID)
		if err != nil {
			return err
		}

		productType, err := vs.vr.GetProductType(ctx, variant.ProductID)
		if err != nil {
			return err
		}
		// Simple products must keep their single default variant.
		if productType == model.ProductTypeSimple && variant.IsDefault {
			return errSimpleProductLocked
		}

		if err := vs.vr.Delete(ctx, variantID); err != nil {
			return err
		}

		if variant.IsDefault {
			if err := vs.vr.SetFirstAvailableAsDefault(ctx, variant.ProductID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, errSimpleProductLocked) {
			return simpleProductLockedError()
		}

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

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		productID, defaultVariantID, err := vs.vr.GetDeletedVariantInfo(ctx, variantID)
		if err != nil {
			return err
		}

		// Lock parent product
		if err := vs.vr.LockProduct(ctx, productID); err != nil {
			return err
		}

		productType, err := vs.vr.GetProductType(ctx, productID)
		if err != nil {
			return err
		}
		if productType == model.ProductTypeSimple {
			count, err := vs.vr.CountByProductID(ctx, productID)
			if err != nil {
				return err
			}
			// Restoring into a simple product is only valid when it has no active variant,
			// and the restored variant becomes its fixed default.
			if count > 0 {
				return errSimpleProductLocked
			}
			if err := vs.vr.SetDefaultVariant(ctx, productID, variantID); err != nil {
				return err
			}
		} else if defaultVariantID == nil {
			// If no default variant is currently designated, make the restored variant default
			if err := vs.vr.SetDefaultVariant(ctx, productID, variantID); err != nil {
				return err
			}
		}

		return vs.vr.Restore(ctx, variantID)
	})

	if err != nil {
		if errors.Is(err, errSimpleProductLocked) {
			return simpleProductLockedError()
		}

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

	var createdMedia *model.VariantMedia

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		var repoErr error
		createdMedia, repoErr = vs.vr.AddMedia(ctx, CreateVariantMediaInput{
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
	mediaID int64,
) error {
	if variantID == uuid.Nil || mediaID == 0 {
		return apierr.BadRequestError("Variant ID and Media ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		return vs.vr.RemoveMedia(ctx, variantID, mediaID)
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
	mediaID int64,
) error {
	return vs.RemoveMedia(ctx, variantID, mediaID)
}

// ReorderMedia applies a new sequence order to all media items attached to a variant.
func (vs *VariantService) ReorderMedia(
	ctx context.Context,
	variantID uuid.UUID,
	orderedMediaIDs []int64,
) error {
	if variantID == uuid.Nil {
		return apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if len(orderedMediaIDs) == 0 {
		return apierr.BadRequestError("Ordered media IDs cannot be empty").
			WithCode(apierr.CodeInvalidInput)
	}

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		existingMedia, err := vs.vr.ListMediaByVariantID(ctx, variantID)
		if err != nil {
			return err
		}

		if len(existingMedia) != len(orderedMediaIDs) {
			return apierr.BadRequestError("Ordered media list must contain all variant media items").
				WithCode(apierr.CodeInvalidInput)
		}

		existingMap := make(map[int64]bool, len(existingMedia))
		for _, m := range existingMedia {
			existingMap[m.ID] = true
		}

		for _, id := range orderedMediaIDs {
			if !existingMap[id] {
				return apierr.BadRequestError("Media ID does not belong to this variant").
					WithCode(apierr.CodeInvalidInput)
			}
		}

		return vs.vr.ReorderMedia(ctx, variantID, orderedMediaIDs)
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

// ListVariantMedia lists a variant's media, resolving public object URLs.
//
// It checks a connection out of the pool; in-transaction callers use the
// unexported listVariantMedia with their own executor.
func (vs *VariantService) ListVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
) ([]*model.VariantMedia, error) {
	var mediaList []*model.VariantMedia
	err := vs.dr.WithExecutor(ctx, func(ctx context.Context) error {
		var repoErr error
		mediaList, repoErr = vs.listVariantMedia(ctx, variantID)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	return mediaList, nil
}

func (vs *VariantService) listVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
) ([]*model.VariantMedia, error) {
	if variantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	mediaList, err := vs.vr.ListMediaByVariantID(ctx, variantID)
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

		if m.Object.PublicURL == "" && m.Object.Bucket != "" && m.Object.Key != "" && vs.storageProvider != nil {
			url := vs.storageProvider.GetPublicURL(
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

	err := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		v, err := vs.vr.GetByID(ctx, variantID)
		if err != nil {
			return err
		}
		if v.ProductID != productID {
			return apierr.BadRequestError("Variant does not belong to this product").
				WithCode(errcode.CodeVariantProductMismatch)
		}

		productType, err := vs.vr.GetProductType(ctx, productID)
		if err != nil {
			return err
		}
		// A simple product's default variant is fixed and cannot be swapped.
		if productType == model.ProductTypeSimple && !v.IsDefault {
			return errSimpleProductLocked
		}

		return vs.vr.SetDefaultVariant(ctx, productID, variantID)
	})

	if err != nil {
		if errors.Is(err, errSimpleProductLocked) {
			return simpleProductLockedError()
		}

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

// AdjustStock updates variant stock and records the ledger entry. The inventory
// domain owns the mutation (lock, invariant check, ledger append, reload), so the
// variant service delegates instead of reimplementing the stock math.
func (vs *VariantService) AdjustStock(
	ctx context.Context,
	in inventory.AdjustStockInput,
) (*model.InventoryStock, error) {
	return vs.inventory.AdjustStock(ctx, in)
}

// GetStockByVariantID retrieves stock levels for a given variant.
func (vs *VariantService) GetStockByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) (*model.InventoryStock, error) {
	if variantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var stock *model.InventoryStock
	err := vs.dr.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		stock, err = vs.inventoryReader.GetStock(ctx, variantID)
		return err
	})
	if err != nil {
		return nil, vs.handleError(err, "Failed to fetch stock")
	}
	return stock, nil
}

// ListLedgers returns paginated audit log entries for variant inventory movements.
func (vs *VariantService) ListLedgers(
	ctx context.Context,
	variantID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.InventoryLedger], error) {
	if variantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var result *pagination.PagedResult[model.InventoryLedger]
	err := vs.dr.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		result, err = vs.inventoryReader.ListVariantLedgers(ctx, variantID, q)
		return err
	})
	if err != nil {
		return nil, vs.handleError(err, "Failed to list inventory ledgers")
	}
	return result, nil
}

// UploadVariantThumbnail handles multipart thumbnail upload, size and MIME type validation,
// S3/MinIO upload, and variant table attachment, safely deleting any previously uploaded thumbnail.
func (vs *VariantService) UploadVariantThumbnail(
	ctx context.Context,
	variantID uuid.UUID,
	fileHeader *multipart.FileHeader,
) (*model.Object, error) {
	if variantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, apierr.BadRequestError("Failed to open uploaded file").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}
	defer func() {
		_ = file.Close()
	}()

	const maxFileSize = 10 * 1024 * 1024
	if fileHeader.Size > maxFileSize {
		return nil, apierr.BadRequestError("File size exceeds 10MB limit").
			WithCode(apierr.CodeFileTooLarge)
	}

	detectedContentType, err := fileUtil.DetectContentType(file)
	if err != nil {
		return nil, apierr.BadRequestError("Failed to detect file content type").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, apierr.BadRequestError("Failed to reset file reader").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	if !fileUtil.IsAllowedImageMime(detectedContentType) {
		return nil, apierr.BadRequestError(
			"Only JPEG, PNG, WebP, and AVIF images are allowed",
		).WithCode(apierr.CodeUnsupportedMediaType)
	}

	bucket := "product-thumbnails"
	key := uuid.New().String()

	cleanCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := vs.storageProvider.Upload(cleanCtx, bucket, key, file, fileHeader.Size, detectedContentType); err != nil {
		return nil, apierr.InternalServerError("Failed to upload file to storage").
			WithCode(apierr.CodeStorageError).
			Wrap(err).
			WithStack()
	}

	obj := &model.Object{
		ID:          uuid.New(),
		Bucket:      bucket,
		Key:         key,
		ContentType: detectedContentType,
		FileSize:    fileHeader.Size,
		Status:      model.ObjectStatusUploaded,
	}

	var oldThumbnailObjectID *uuid.UUID
	var oldObj *model.Object
	txErr := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		variant, err := vs.vr.GetByID(ctx, variantID)
		if err != nil {
			return err
		}
		oldThumbnailObjectID = variant.ThumbnailObjectID

		if err := vs.objectRepo.Create(ctx, obj); err != nil {
			return err
		}

		if oldThumbnailObjectID != nil && *oldThumbnailObjectID != obj.ID {
			oldObj, _ = vs.objectRepo.GetByID(ctx, *oldThumbnailObjectID)
			if oldObj != nil {
				_ = vs.objectRepo.MarkDeletingByKey(ctx, oldObj.Bucket, oldObj.Key)
			}
		}

		return vs.vr.Update(ctx, variantID, UpdateVariantFields{
			ThumbnailObjectID: &obj.ID,
		})
	})

	if txErr != nil {
		_ = vs.storageProvider.Delete(ctx, bucket, key)
		return nil, vs.handleError(txErr, "Failed to attach thumbnail to variant")
	}

	if oldObj != nil && vs.storageProvider != nil {
		if delErr := vs.storageProvider.Delete(ctx, oldObj.Bucket, oldObj.Key); delErr == nil {
			_ = vs.dr.WithTx(ctx, func(ctx context.Context) error {
				return vs.objectRepo.MarkDeletedByKey(ctx, oldObj.Bucket, oldObj.Key)
			})
		}
	}

	obj.PublicURL = vs.storageProvider.GetPublicURL(obj.Bucket, obj.Key)
	return obj, nil
}

// DeleteVariantThumbnail removes the thumbnail association and deletes the storage object.
func (vs *VariantService) DeleteVariantThumbnail(
	ctx context.Context,
	variantID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var oldObj *model.Object
	txErr := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		variant, err := vs.vr.GetByID(ctx, variantID)
		if err != nil {
			return err
		}
		if variant.ThumbnailObjectID != nil && vs.objectRepo != nil {
			oldObj, _ = vs.objectRepo.GetByID(ctx, *variant.ThumbnailObjectID)
			if oldObj != nil {
				_ = vs.objectRepo.MarkDeletingByKey(ctx, oldObj.Bucket, oldObj.Key)
			}
		}
		return vs.vr.ClearThumbnail(ctx, variantID)
	})
	if txErr != nil {
		return vs.handleError(txErr, "Failed to remove variant thumbnail")
	}

	if oldObj != nil && vs.storageProvider != nil {
		if delErr := vs.storageProvider.Delete(ctx, oldObj.Bucket, oldObj.Key); delErr == nil {
			_ = vs.dr.WithTx(ctx, func(ctx context.Context) error {
				return vs.objectRepo.MarkDeletedByKey(ctx, oldObj.Bucket, oldObj.Key)
			})
		}
	}
	return nil
}

// UploadVariantMedia handles multipart preview image upload and links it as variant media.
func (vs *VariantService) UploadVariantMedia(
	ctx context.Context,
	variantID uuid.UUID,
	fileHeader *multipart.FileHeader,
) (*model.VariantMedia, error) {
	if variantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, apierr.BadRequestError("Failed to open uploaded file").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}
	defer func() {
		_ = file.Close()
	}()

	const maxFileSize = 10 * 1024 * 1024
	if fileHeader.Size > maxFileSize {
		return nil, apierr.BadRequestError("File size exceeds 10MB limit").
			WithCode(apierr.CodeFileTooLarge)
	}

	detectedContentType, err := fileUtil.DetectContentType(file)
	if err != nil {
		return nil, apierr.BadRequestError("Failed to detect file content type").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, apierr.BadRequestError("Failed to reset file reader").
			WithCode(apierr.CodeValidationFailed).
			Wrap(err)
	}

	if !fileUtil.IsAllowedImageMime(detectedContentType) {
		return nil, apierr.BadRequestError(
			"Only JPEG, PNG, WebP, and AVIF images are allowed",
		).WithCode(apierr.CodeUnsupportedMediaType)
	}

	bucket := "catalog"
	key := uuid.New().String()

	cleanCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := vs.storageProvider.Upload(cleanCtx, bucket, key, file, fileHeader.Size, detectedContentType); err != nil {
		return nil, apierr.InternalServerError("Failed to upload file to storage").
			WithCode(apierr.CodeStorageError).
			Wrap(err).
			WithStack()
	}

	obj := &model.Object{
		ID:          uuid.New(),
		Bucket:      bucket,
		Key:         key,
		ContentType: detectedContentType,
		FileSize:    fileHeader.Size,
		Status:      model.ObjectStatusUploaded,
	}

	var createdMedia *model.VariantMedia
	txErr := vs.dr.WithTx(ctx, func(ctx context.Context) error {
		if _, err := vs.vr.GetByID(ctx, variantID); err != nil {
			return err
		}
		if err := vs.objectRepo.Create(ctx, obj); err != nil {
			return err
		}
		var repoErr error
		createdMedia, repoErr = vs.vr.AddMedia(ctx, CreateVariantMediaInput{
			VariantID: variantID,
			ObjectID:  obj.ID,
			MediaType: detectedContentType,
			SortOrder: -1,
		})
		return repoErr
	})

	if txErr != nil {
		_ = vs.storageProvider.Delete(ctx, bucket, key)
		return nil, vs.handleError(txErr, "Failed to attach media to variant")
	}

	obj.PublicURL = vs.storageProvider.GetPublicURL(obj.Bucket, obj.Key)
	createdMedia.Object = obj
	return createdMedia, nil
}
