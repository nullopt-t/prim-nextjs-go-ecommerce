package attribute

import (
	"context"

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

func (s *AttributeService) DeleteAttribute(ctx context.Context, id uuid.UUID) error {
	return s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.DeleteAttribute(ctx, tx, id)
	})
}
