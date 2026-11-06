package history

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// Service owns browsing-history rules, transaction boundaries, and error mapping.
type Service struct {
	runner  database.Runner
	repo    *Repository
	storage storage.StorageProvider
}

// NewService constructs a browsing-history service.
func NewService(
	runner database.Runner,
	repo *Repository,
	storageProvider storage.StorageProvider,
) *Service {
	if runner == nil {
		panic("history: database runner is required")
	}
	if repo == nil {
		panic("history: repository is required")
	}
	if storageProvider == nil {
		panic("history: storage provider is required")
	}
	return &Service{
		runner:  runner,
		repo:    repo,
		storage: storageProvider,
	}
}

func (s *Service) handleError(err error, fallback string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mapped := database.MapError(err)
	if errors.Is(mapped, database.ErrNotFound) {
		return apierr.NotFoundError("History item not found").
			WithCode(errcode.CodeHistoryItemNotFound).
			Wrap(err)
	}
	return apierr.InternalServerError(fallback).
		WithCode(apierr.CodeInternalError).
		Wrap(err).
		WithStack()
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return DefaultListLimit
	}
	if limit > MaxItems {
		return MaxItems
	}
	return limit
}

// RecordView records or refreshes one product view and trims the rolling window
// in the same transaction.
func (s *Service) RecordView(ctx context.Context, userID, productID uuid.UUID) error {
	var found bool

	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		var touchErr error
		found, touchErr = s.repo.Touch(ctx, userID, productID)
		if touchErr != nil || !found {
			return touchErr
		}
		return s.repo.Prune(ctx, userID, MaxItems)
	})
	if err != nil {
		return s.handleError(err, "Failed to record product view")
	}
	if !found {
		return apierr.NotFoundError("Product not found").
			WithCode(errcode.CodeProductNotFound)
	}
	return nil
}

// List returns the newest visible views for the authenticated customer.
func (s *Service) List(ctx context.Context, userID uuid.UUID, limit int) ([]ViewedProductItem, error) {
	limit = clampLimit(limit)

	var items []ViewedProductItem
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		items, err = s.repo.List(ctx, userID, limit)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to load browsing history")
	}

	for i := range items {
		s.resolveThumbnail(&items[i])
	}
	return items, nil
}

// SyncGuestHistory merges guest product identifiers into the authenticated
// account, ignores products that are no longer publicly visible, and caps the
// result at the rolling-window size.
func (s *Service) SyncGuestHistory(
	ctx context.Context,
	userID uuid.UUID,
	productIDs []uuid.UUID,
	slugs []string,
) (int, error) {
	if len(productIDs)+len(slugs) == 0 {
		return 0, apierr.ValidationFailedError("at least one product id or slug is required").
			WithFields(api.FieldError{
				Field:      "slugs",
				Constraint: "required",
				Message:    "provide at least one product id or slug",
			})
	}
	if len(productIDs)+len(slugs) > MaxSyncItems {
		return 0, apierr.ValidationFailedError("too many history items").
			WithFields(api.FieldError{
				Field:      "slugs",
				Constraint: "max",
				Message:    "at most 100 product ids and slugs may be merged at once",
			})
	}

	uniqueIDs := make([]uuid.UUID, 0, len(productIDs))
	seenIDs := make(map[uuid.UUID]struct{}, len(productIDs))
	for _, id := range productIDs {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seenIDs[id]; ok {
			continue
		}
		seenIDs[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	uniqueSlugs := make([]string, 0, len(slugs))
	seenSlugs := make(map[string]struct{}, len(slugs))
	for _, slug := range slugs {
		slug = strings.ToLower(strings.TrimSpace(slug))
		if slug == "" {
			continue
		}
		if _, ok := seenSlugs[slug]; ok {
			continue
		}
		seenSlugs[slug] = struct{}{}
		uniqueSlugs = append(uniqueSlugs, slug)
	}
	if len(uniqueIDs) == 0 && len(uniqueSlugs) == 0 {
		return 0, apierr.ValidationFailedError("at least one valid product id or slug is required")
	}

	var synced int
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		var err error
		synced, err = s.repo.Merge(ctx, userID, uniqueIDs, uniqueSlugs)
		if err != nil {
			return err
		}
		return s.repo.Prune(ctx, userID, MaxItems)
	})
	if err != nil {
		return 0, s.handleError(err, "Failed to merge browsing history")
	}
	return synced, nil
}

// DeleteView removes one product from the caller's history.
func (s *Service) DeleteView(ctx context.Context, userID, productID uuid.UUID) error {
	var found bool
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		var deleteErr error
		found, deleteErr = s.repo.Delete(ctx, userID, productID)
		return deleteErr
	})
	if err != nil {
		return s.handleError(err, "Failed to delete browsing-history item")
	}
	if !found {
		return apierr.NotFoundError("History item not found").
			WithCode(errcode.CodeHistoryItemNotFound)
	}
	return nil
}

// Clear removes all history for the caller.
func (s *Service) Clear(ctx context.Context, userID uuid.UUID) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		return s.repo.Clear(ctx, userID)
	})
	if err != nil {
		return s.handleError(err, "Failed to clear browsing history")
	}
	return nil
}

func (s *Service) resolveThumbnail(item *ViewedProductItem) {
	if item.thumbnailBucket == nil || item.thumbnailKey == nil {
		return
	}
	if *item.thumbnailBucket == "" || *item.thumbnailKey == "" {
		return
	}
	url := s.storage.GetPublicURL(*item.thumbnailBucket, *item.thumbnailKey)
	item.ThumbnailURL = &url
}
