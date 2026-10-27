package attribute

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type AttributeService struct {
	dr   database.Runner
	repo *AttributeRepository
}

func NewService(dr database.Runner, repo *AttributeRepository) *AttributeService {
	return &AttributeService{
		dr:   dr,
		repo: repo,
	}
}

func (s *AttributeService) ListAttributes(ctx context.Context) ([]AttributeItem, error) {
	var list []AttributeItem
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		list, err = s.repo.ListAttributes(ctx, exec)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to fetch attributes")
	}
	return list, nil
}

func (s *AttributeService) GetAttributeByID(ctx context.Context, id uuid.UUID) (*AttributeItem, error) {
	var item *AttributeItem
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		item, err = s.repo.GetByID(ctx, exec, id)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to fetch attribute")
	}
	return item, nil
}

func (s *AttributeService) CreateAttribute(ctx context.Context, ai *AttributeItem) error {
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		if ai.ID == uuid.Nil {
			ai.ID = uuid.New()
		}
		if ai.ValueType == "" {
			ai.ValueType = "text"
		}
		return s.repo.CreateAttribute(ctx, tx, ai)
	})
	if err != nil {
		return s.handleError(err, "Failed to create attribute")
	}
	return nil
}

// UpdateAttributeInput describes the optional attribute fields to patch.
type UpdateAttributeInput struct {
	Label         *string
	ValueType     *string
	AllowedValues json.RawMessage
	IsFilterable  *bool
	SortOrder     *int
}

func (s *AttributeService) UpdateAttribute(ctx context.Context, id uuid.UUID, in UpdateAttributeInput) error {
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		return s.repo.UpdateAttribute(ctx, tx, id, UpdateAttributeFields(in))
	})
	if err != nil {
		return s.handleError(err, "Failed to update attribute")
	}
	return nil
}

func (s *AttributeService) DeleteAttribute(ctx context.Context, id uuid.UUID) error {
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		return s.repo.DeleteAttribute(ctx, tx, id)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete attribute")
	}
	return nil
}

func (s *AttributeService) handleError(err error, defaultMsg string) error {
	if err == nil {
		return nil
	}
	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError("Attribute not found").
			WithCode(errcode.CodeAttributeNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Attribute with this key already exists").
			WithCode(errcode.CodeAttributeAlreadyExists).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced category does not exist").
			WithCode(errcode.CodeCategoryNotFound).
			Wrap(err)
	default:
		return apierr.InternalServerError(defaultMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}
