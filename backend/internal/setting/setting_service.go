package setting

import (
	"context"
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
