package setting

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// cacheTTL bounds how long a resolved config is served from memory before the
// database is consulted again. It keeps per-request consumers (e.g. checkout and
// the maintenance middleware) cheap while still reflecting admin changes quickly.
const cacheTTL = 5 * time.Second

// Service coordinates administrative settings reads and writes.
type Service struct {
	dr   database.Runner
	repo *Repository

	mu       sync.RWMutex
	cached   *Config
	cachedAt time.Time
}

// NewService constructs a settings service bound to a transaction runner and repository.
func NewService(dr database.Runner, repo *Repository) *Service {
	return &Service{
		dr:   dr,
		repo: repo,
	}
}

// handleError maps persistence, driver, and context errors to domain APIErrors.
func (s *Service) handleError(err error, fallbackMsg string) error {
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
		return apierr.NotFoundError("Setting not found").
			WithCode(apierr.CodeNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Setting already exists").
			WithCode(apierr.CodeAlreadyExists).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced record does not exist").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrCheckViolation),
		errors.Is(mappedErr, database.ErrNotNullViolation),
		errors.Is(mappedErr, database.ErrInvalidInput):
		return apierr.BadRequestError("Invalid input").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrTimeout), errors.Is(mappedErr, database.ErrCanceled):
		return apierr.New(http.StatusGatewayTimeout, "database operation timed out or canceled").
			WithCode(apierr.CodeGatewayTimeout).
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

// GetSettings returns all persisted settings as a key/value map.
func (s *Service) GetSettings(ctx context.Context) (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	err := s.dr.WithExecutor(ctx, func(db database.QueryExecutor) error {
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
		return nil, s.handleError(err, "Failed to fetch settings")
	}

	return settings, nil
}

// GetConfig returns the resolved, typed configuration overlaid on defaults. The
// result is memoized for cacheTTL and invalidated whenever settings are updated.
func (s *Service) GetConfig(ctx context.Context) (Config, error) {
	if cfg, ok := s.cachedConfig(); ok {
		return cfg, nil
	}

	settings, err := s.GetSettings(ctx)
	if err != nil {
		return DefaultConfig(), err
	}

	cfg := configFromMap(settings)
	s.storeConfig(cfg)
	return cfg, nil
}

// GetPublicConfig returns the storefront-safe projection of the resolved config.
func (s *Service) GetPublicConfig(ctx context.Context) (PublicConfig, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return PublicConfig{}, err
	}
	return cfg.Public(), nil
}

// UpdateSettings upserts the provided key/value pairs within a single transaction.
func (s *Service) UpdateSettings(ctx context.Context, items map[string]interface{}) error {
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		for key, value := range items {
			if err := s.repo.Upsert(ctx, tx, key, value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return s.handleError(err, "Failed to update settings")
	}

	// Force the next read to observe the freshly written values immediately.
	s.invalidate()
	return nil
}

func (s *Service) cachedConfig() (Config, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cached == nil || time.Since(s.cachedAt) > cacheTTL {
		return Config{}, false
	}
	return *s.cached, true
}

func (s *Service) storeConfig(cfg Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached = &cfg
	s.cachedAt = time.Now()
}

func (s *Service) invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached = nil
}
