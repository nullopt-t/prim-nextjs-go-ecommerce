package setting

import (
	"context"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// Service coordinates administrative settings reads and writes.
type Service struct {
	dr   database.Runner
	repo *Repository
}

// NewService constructs a settings service bound to a transaction runner and repository.
func NewService(dr database.Runner, repo *Repository) *Service {
	return &Service{
		dr:   dr,
		repo: repo,
	}
}

// GetSettings returns all persisted settings as a key/value map.
func (s *Service) GetSettings(ctx context.Context) (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		list, err := s.repo.GetAll(ctx, db)
		if err != nil {
			return err
		}
		for _, it := range list {
			settings[it.Key] = it.Value
		}
		return nil
	})
	if err != nil {
		return nil, apierr.InternalServerError("Failed to fetch settings").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return settings, nil
}

// UpdateSettings upserts the provided key/value pairs within a single transaction.
func (s *Service) UpdateSettings(ctx context.Context, items map[string]interface{}) error {
	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		for key, value := range items {
			if err := s.repo.Upsert(ctx, tx, key, value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return apierr.InternalServerError("Failed to update settings").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}
