package audit

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdversarial_Audit_HandleError(t *testing.T) {
	svc := &Service{}
	alreadyWrapped := apierr.BadRequestError("already wrapped").WithCode("CUSTOM_CODE")

	tests := []struct {
		name         string
		err          error
		fallbackMsg  string
		wantStatus   int
		wantCode     string
		wantExactErr error
	}{
		{
			name:        "1. ErrNoRows -> 404",
			err:         pgx.ErrNoRows,
			fallbackMsg: "audit log",
			wantStatus:  http.StatusNotFound,
			wantCode:    apierr.CodeNotFound,
		},
		{
			name:        "2. Unique constraint violation -> 409",
			err:         &pgconn.PgError{Code: "23505"},
			fallbackMsg: "audit log",
			wantStatus:  http.StatusConflict,
			wantCode:    apierr.CodeResourceConflict,
		},
		{
			name:        "3. Foreign key violation -> 400",
			err:         &pgconn.PgError{Code: "23503"},
			fallbackMsg: "audit log",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidReference,
		},
		{
			name:        "4a. Check constraint violation -> 400",
			err:         &pgconn.PgError{Code: "23514"},
			fallbackMsg: "audit log",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "4b. Not null violation -> 400",
			err:         &pgconn.PgError{Code: "23502"},
			fallbackMsg: "audit log",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "5a. Timeout (context.DeadlineExceeded) -> 504 GatewayTimeout",
			err:         context.DeadlineExceeded,
			fallbackMsg: "audit log",
			wantStatus:  http.StatusGatewayTimeout,
			wantCode:    apierr.CodeGatewayTimeout,
		},
		{
			name:        "5b. Timeout (database.ErrTimeout) -> 504 GatewayTimeout",
			err:         database.ErrTimeout,
			fallbackMsg: "audit log",
			wantStatus:  http.StatusGatewayTimeout,
			wantCode:    apierr.CodeGatewayTimeout,
		},
		{
			name:        "5c. Cancellation (context.Canceled) -> 504 GatewayTimeout",
			err:         context.Canceled,
			fallbackMsg: "audit log",
			wantStatus:  http.StatusGatewayTimeout,
			wantCode:    apierr.CodeGatewayTimeout,
		},
		{
			name:        "5d. Cancellation (database.ErrCanceled) -> 504 GatewayTimeout",
			err:         database.ErrCanceled,
			fallbackMsg: "audit log",
			wantStatus:  http.StatusGatewayTimeout,
			wantCode:    apierr.CodeGatewayTimeout,
		},
		{
			name:         "6. Already-wrapped apierr.APIError -> returned unchanged",
			err:          alreadyWrapped,
			fallbackMsg:  "audit log",
			wantExactErr: alreadyWrapped,
			wantStatus:   http.StatusBadRequest,
			wantCode:     "CUSTOM_CODE",
		},
		{
			name:        "7. Unclassified error -> 500 with fallback message",
			err:         errors.New("unclassified disk failure"),
			fallbackMsg: "Failed to persist audit log",
			wantStatus:  http.StatusInternalServerError,
			wantCode:    apierr.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := svc.handleError(tt.err, tt.fallbackMsg)
			require.Error(t, res)
			if tt.wantExactErr != nil {
				assert.Equal(t, tt.wantExactErr, res)
			}
			var apiErr *apierr.APIError
			require.True(t, errors.As(res, &apiErr))
			assert.Equal(t, tt.wantStatus, apiErr.Status)
			assert.Equal(t, tt.wantCode, apiErr.Code)
		})
	}
}
