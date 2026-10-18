package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type recordingLogger struct {
	errors []string
	warns  []string
	debugs []string
}

func (r *recordingLogger) Info(string, ...log.Meta)      {}
func (r *recordingLogger) Error(m string, _ ...log.Meta) { r.errors = append(r.errors, m) }
func (r *recordingLogger) Warn(m string, _ ...log.Meta)  { r.warns = append(r.warns, m) }
func (r *recordingLogger) Debug(m string, _ ...log.Meta) { r.debugs = append(r.debugs, m) }

func canceledRequest() *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	return req.WithContext(ctx)
}

func serveError(t *testing.T, logger *recordingLogger, err error, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorHandler(logger))
	r.GET("/api/v1/products", func(c *gin.Context) {
		c.Error(err)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func codeOf(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding response body %q: %v", w.Body.String(), err)
	}
	return body.Code
}

// A request whose client disconnected must not be reported as a 500 or logged
// as an internal error; it is recorded as 499 instead.
func TestErrorHandler_ClientCanceledIsNotServerError(t *testing.T) {
	logger := &recordingLogger{}
	w := serveError(t, logger, apierr.InternalServerError("Failed to list products").WithStack(), canceledRequest())

	if w.Code != statusClientClosedRequest {
		t.Fatalf("status = %d, want %d", w.Code, statusClientClosedRequest)
	}
	if len(logger.errors) != 0 {
		t.Fatalf("logged errors = %v, want none", logger.errors)
	}
	if len(logger.debugs) == 0 {
		t.Fatalf("expected a debug log for the canceled request")
	}
}

// A genuine internal error on a live request must still be a logged 500.
func TestErrorHandler_InternalErrorStillLogged(t *testing.T) {
	logger := &recordingLogger{}
	w := serveError(t, logger, apierr.InternalServerError("Failed to list products").WithStack(),
		httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if len(logger.errors) == 0 {
		t.Fatalf("expected an error log for a real 500")
	}
}

// A statement_timeout / deadline surfaced from the database must be reported as
// 504 Gateway Timeout rather than a blanket 500.
func TestErrorHandler_MapsDatabaseTimeoutToGatewayTimeout(t *testing.T) {
	logger := &recordingLogger{}
	w := serveError(t, logger, apierr.InternalServerError("Failed to list products").Wrap(database.ErrTimeout).WithStack(),
		httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504", w.Code)
	}
	if got := codeOf(t, w); got != apierr.CodeGatewayTimeout {
		t.Fatalf("code = %q, want %q", got, apierr.CodeGatewayTimeout)
	}
}

// Pool/connection exhaustion must be reported as 503 Service Unavailable.
func TestErrorHandler_MapsConnectionFailureToServiceUnavailable(t *testing.T) {
	logger := &recordingLogger{}
	w := serveError(t, logger, apierr.InternalServerError("Failed to list products").Wrap(database.ErrConnectionFailure).WithStack(),
		httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
	if got := codeOf(t, w); got != apierr.CodeServiceUnavailable {
		t.Fatalf("code = %q, want %q", got, apierr.CodeServiceUnavailable)
	}
}

// A mapped database cancellation (context.Canceled translated by the db layer)
// must be classified as 499 even when the live request context is not cancelled.
func TestErrorHandler_MapsDatabaseCanceledTo499(t *testing.T) {
	logger := &recordingLogger{}
	w := serveError(t, logger, apierr.InternalServerError("Failed to list products").Wrap(database.ErrCanceled).WithStack(),
		httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))

	if w.Code != statusClientClosedRequest {
		t.Fatalf("status = %d, want %d", w.Code, statusClientClosedRequest)
	}
}

// A raw statement_timeout error (SQLSTATE 57014) that a service never mapped
// must still surface as 504 Gateway Timeout, not a 500.
func TestErrorHandler_MapsRawStatementTimeoutToGatewayTimeout(t *testing.T) {
	logger := &recordingLogger{}
	raw := &pgconn.PgError{Code: "57014", Message: "canceling statement due to statement timeout"}
	w := serveError(t, logger, apierr.InternalServerError("Failed to list products").Wrap(raw).WithStack(),
		httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504", w.Code)
	}
	if got := codeOf(t, w); got != apierr.CodeGatewayTimeout {
		t.Fatalf("code = %q, want %q", got, apierr.CodeGatewayTimeout)
	}
}

// A raw pool-exhaustion error (SQLSTATE 53300) must surface as 503.
func TestErrorHandler_MapsRawPoolExhaustionToServiceUnavailable(t *testing.T) {
	logger := &recordingLogger{}
	raw := &pgconn.PgError{Code: "53300", Message: "sorry, too many clients already"}
	w := serveError(t, logger, apierr.InternalServerError("Failed to list products").Wrap(raw).WithStack(),
		httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
}

// A raw context cancellation on a still-live request must be 499, not 500.
func TestErrorHandler_MapsRawContextCanceledTo499(t *testing.T) {
	logger := &recordingLogger{}
	w := serveError(t, logger, apierr.InternalServerError("Failed to list products").Wrap(context.Canceled).WithStack(),
		httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))

	if w.Code != statusClientClosedRequest {
		t.Fatalf("status = %d, want %d", w.Code, statusClientClosedRequest)
	}
}

// Transient 503/504 responses must advertise Retry-After so clients back off,
// while terminal 5xx (and 4xx) responses must not.
func TestErrorHandler_RetryAfterHint(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantCode  int
		wantRetry bool
	}{
		{
			name:      "database timeout is 504 with retry hint",
			err:       apierr.InternalServerError("Failed to list products").Wrap(database.ErrTimeout).WithStack(),
			wantCode:  http.StatusGatewayTimeout,
			wantRetry: true,
		},
		{
			name:      "connection failure is 503 with retry hint",
			err:       apierr.InternalServerError("Failed to list products").Wrap(database.ErrConnectionFailure).WithStack(),
			wantCode:  http.StatusServiceUnavailable,
			wantRetry: true,
		},
		{
			name:      "direct service unavailable carries retry hint",
			err:       apierr.New(http.StatusServiceUnavailable, "Service temporarily unavailable").WithCode(apierr.CodeServiceUnavailable),
			wantCode:  http.StatusServiceUnavailable,
			wantRetry: true,
		},
		{
			name:      "internal error has no retry hint",
			err:       apierr.InternalServerError("Failed to list products").WithStack(),
			wantCode:  http.StatusInternalServerError,
			wantRetry: false,
		},
		{
			name:      "conflict has no retry hint",
			err:       apierr.ConflictError("Concurrent modification"),
			wantCode:  http.StatusConflict,
			wantRetry: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &recordingLogger{}
			w := serveError(t, logger, tc.err, httptest.NewRequest(http.MethodGet, "/api/v1/products", nil))

			if w.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantCode)
			}
			got := w.Header().Get("Retry-After")
			if tc.wantRetry && got != retryAfterSeconds {
				t.Fatalf("Retry-After = %q, want %q", got, retryAfterSeconds)
			}
			if !tc.wantRetry && got != "" {
				t.Fatalf("Retry-After = %q, want none", got)
			}
		})
	}
}
