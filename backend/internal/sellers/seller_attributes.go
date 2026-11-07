package sellers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/category/attribute"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// SellerAttributeStore is the seller portal's window into seller-scoped
// attribute definitions.
type SellerAttributeStore interface {
	ListBySellerCategory(ctx context.Context, sellerID, categoryID uuid.UUID) ([]attribute.AttributeItem, error)
	// ListByCategory returns the platform definitions that already describe the
	// category, so a merchant key cannot shadow one of them.
	ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]attribute.AttributeItem, error)
	CreateSellerAttribute(ctx context.Context, sellerID uuid.UUID, ai *attribute.AttributeItem) error
	UpdateSellerAttribute(ctx context.Context, sellerID, id uuid.UUID, fields attribute.UpdateAttributeFields) error
	DeleteSellerAttribute(ctx context.Context, sellerID, id uuid.UUID) error
}

var sellerAttributeAllowedTypes = map[string]struct{}{
	"enum": {}, "text": {}, "number": {}, "boolean": {},
}

// ListAttributes returns the merchant's own custom attributes for a category.
// The caller merges them with the platform category attributes.
func (s *Service) ListAttributes(
	ctx context.Context,
	userID, categoryID uuid.UUID,
) ([]attribute.AttributeItem, error) {
	var items []attribute.AttributeItem
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		if categoryID == uuid.Nil {
			return apierr.BadRequestError("Category ID is required").WithCode(apierr.CodeInvalidInput)
		}
		items, err = s.attrs.ListBySellerCategory(ctx, seller.ID, categoryID)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to list attributes")
	}
	return items, nil
}

// CreateAttribute defines a custom attribute for the merchant in a category.
func (s *Service) CreateAttribute(
	ctx context.Context,
	userID uuid.UUID,
	categoryID uuid.UUID,
	key, label, valueType string,
	allowedValues []string,
) (*attribute.AttributeItem, error) {
	code := strings.TrimSpace(strings.ToLower(key))
	if code == "" {
		return nil, apierr.BadRequestError("Attribute key is required").WithCode(apierr.CodeInvalidInput)
	}
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, apierr.BadRequestError("Attribute label is required").WithCode(apierr.CodeInvalidInput)
	}
	if valueType == "" {
		valueType = "enum"
	}
	if _, ok := sellerAttributeAllowedTypes[valueType]; !ok {
		return nil, apierr.BadRequestError("Invalid attribute value type").WithCode(apierr.CodeInvalidInput)
	}
	if valueType == "enum" && len(allowedValues) == 0 {
		return nil, apierr.BadRequestError("Enum attributes require at least one allowed value").
			WithCode(apierr.CodeInvalidInput)
	}

	var created *attribute.AttributeItem
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		// Platform attributes are resolved by key when a variant is validated, so
		// a merchant definition with the same key would silently replace the
		// platform's allowed values for every product in the category.
		platformDefs, err := s.attrs.ListByCategory(ctx, categoryID)
		if err != nil {
			return err
		}
		for _, def := range platformDefs {
			if def.Key == code {
				return apierr.ConflictError("An attribute with this key already exists for this category").
					WithCode(apierr.CodeInvalidInput)
			}
		}
		raw, err := json.Marshal(allowedValues)
		if err != nil {
			return err
		}
		ai := &attribute.AttributeItem{
			ID:            uuid.New(),
			CategoryID:    &categoryID,
			Key:           code,
			Label:         label,
			ValueType:     valueType,
			AllowedValues: raw,
			IsFilterable:  false,
		}
		if err := s.attrs.CreateSellerAttribute(ctx, seller.ID, ai); err != nil {
			return err
		}
		created = ai
		return nil
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		if errors.Is(database.MapError(err), database.ErrConflict) {
			return nil, apierr.ConflictError("An attribute with this key already exists for this category").
				WithCode(apierr.CodeInvalidInput)
		}
		if errors.Is(database.MapError(err), database.ErrForeignKeyViolation) {
			return nil, apierr.BadRequestError("Category does not exist").WithCode(apierr.CodeInvalidInput)
		}
		return nil, s.handleError(err, "Failed to create attribute")
	}
	return created, nil
}

// UpdateAttribute edits a merchant-owned attribute definition.
func (s *Service) UpdateAttribute(
	ctx context.Context,
	userID, attributeID uuid.UUID,
	label *string,
	allowedValues []string,
) error {
	if label != nil {
		trimmed := strings.TrimSpace(*label)
		if trimmed == "" {
			return apierr.BadRequestError("Attribute label is required").WithCode(apierr.CodeInvalidInput)
		}
		label = &trimmed
	}
	var raw json.RawMessage
	if allowedValues != nil {
		// An empty option list would leave an enum attribute with no valid
		// values, which blocks every later variant using that axis.
		if len(allowedValues) == 0 {
			return apierr.BadRequestError("Enum attributes require at least one allowed value").
				WithCode(apierr.CodeInvalidInput)
		}
		var err error
		raw, err = json.Marshal(allowedValues)
		if err != nil {
			return err
		}
	}
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		return s.attrs.UpdateSellerAttribute(ctx, seller.ID, attributeID, attribute.UpdateAttributeFields{
			Label:         label,
			AllowedValues: raw,
		})
	})
	if err != nil {
		return s.handleError(err, "Failed to update attribute")
	}
	return nil
}

// DeleteAttribute removes a merchant-owned attribute definition.
func (s *Service) DeleteAttribute(
	ctx context.Context,
	userID, attributeID uuid.UUID,
) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		return s.attrs.DeleteSellerAttribute(ctx, seller.ID, attributeID)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete attribute")
	}
	return nil
}
