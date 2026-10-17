package attribute

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
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
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		list, err = s.repo.ListAttributes(ctx, db)
		return err
	})
	return list, err
}

func (s *AttributeService) CreateAttribute(ctx context.Context, ai *AttributeItem) error {
	return s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		if ai.ID == uuid.Nil {
			ai.ID = uuid.New()
		}
		if ai.ValueType == "" {
			ai.ValueType = "text"
		}
		return s.repo.CreateAttribute(ctx, tx, ai)
	})
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
	return s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.UpdateAttribute(ctx, tx, id, UpdateAttributeFields{
			Label:         in.Label,
			ValueType:     in.ValueType,
			AllowedValues: in.AllowedValues,
			IsFilterable:  in.IsFilterable,
			SortOrder:     in.SortOrder,
		})
	})
}

func (s *AttributeService) DeleteAttribute(ctx context.Context, id uuid.UUID) error {
	return s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.DeleteAttribute(ctx, tx, id)
	})
}
