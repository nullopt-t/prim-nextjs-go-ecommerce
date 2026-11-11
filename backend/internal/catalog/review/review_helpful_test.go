package review_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/review"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeReviewRunner struct{}

func (fakeReviewRunner) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (fakeReviewRunner) WithExecutor(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (fakeReviewRunner) WithRetry(ctx context.Context, fn func(ctx context.Context) error, n int) error {
	return fn(ctx)
}

func TestReviewService_ConstructorsFailFast(t *testing.T) {
	assert.Panics(t, func() {
		review.NewService(nil, review.NewReviewRepository(), user.NewPostgresRepository(), storage.NoopProvider{})
	})
	assert.Panics(t, func() {
		review.NewService(fakeReviewRunner{}, nil, user.NewPostgresRepository(), storage.NoopProvider{})
	})
	assert.Panics(t, func() {
		review.NewService(fakeReviewRunner{}, review.NewReviewRepository(), nil, storage.NoopProvider{})
	})
}

func TestReviewService_ToggleHelpful_Validation(t *testing.T) {
	svc := review.NewService(fakeReviewRunner{}, review.NewReviewRepository(), user.NewPostgresRepository(), storage.NoopProvider{})

	// Nil Review ID returns 400
	_, _, err := svc.ToggleHelpful(context.Background(), uuid.Nil, uuid.New())
	require.Error(t, err)
	var apiErr *apierr.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.Status)

	// Nil User ID returns 401
	_, _, err = svc.ToggleHelpful(context.Background(), uuid.New(), uuid.Nil)
	require.Error(t, err)
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 401, apiErr.Status)
}
