package user_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunner struct{}

func (fakeRunner) WithTx(ctx context.Context, fn func(ctx context.Context, tx *database.Tx) error) error {
	return fn(ctx, nil)
}

func (fakeRunner) WithExecutor(ctx context.Context, fn func(qe database.QueryExecutor) error) error {
	return fn(nil)
}

func TestUserService_ConstructorsFailFast(t *testing.T) {
	assert.Panics(t, func() {
		user.NewService(nil, user.NewPostgresRepository(), log.NewNoopLogger())
	})
	assert.Panics(t, func() {
		user.NewService(fakeRunner{}, nil, log.NewNoopLogger())
	})
	assert.Panics(t, func() {
		user.NewService(fakeRunner{}, user.NewPostgresRepository(), nil)
	})
}

func TestUserService_UpdateProfile_Validation(t *testing.T) {
	svc := user.NewService(fakeRunner{}, user.NewPostgresRepository(), log.NewNoopLogger())

	name := "Alice"
	phone := "+201234567890"

	// Nil UUID fails fast with BadRequestError
	_, err := svc.UpdateProfile(context.Background(), uuid.Nil, &name, &phone)
	require.Error(t, err)
	var apiErr *apierr.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.Status)
}
