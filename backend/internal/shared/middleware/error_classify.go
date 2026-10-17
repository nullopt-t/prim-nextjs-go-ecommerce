package middleware

import (
	"errors"
	"net/http"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// classifyInfraError maps transient infrastructure failures onto accurate HTTP
// statuses so they are not misreported as a blanket 500.
//
// It intentionally covers only failures that are never the client's fault and
// never a simple validation problem: request cancellation, dependency
// timeouts, pool/connection exhaustion, resource limits, and retryable
// concurrency conflicts. Domain errors (not found, uniqueness conflict, invalid
// input) remain owned by the services, which know their own context and codes.
//
// The error is first normalised through database.MapError so that raw driver
// errors (pgx/pgconn, context) are recognised even on code paths that do not
// map them themselves.
//
// It returns the mapped APIError and true when a classification applies.
func classifyInfraError(err error) (*apierr.APIError, bool) {
	if err == nil {
		return nil, false
	}

	mapped := database.MapError(err)

	switch {
	case errors.Is(mapped, database.ErrCanceled):
		return apierr.New(statusClientClosedRequest, "Client closed request").
			WithCode(apierr.CodeClientClosedRequest), true

	case errors.Is(mapped, database.ErrTimeout):
		return apierr.New(http.StatusGatewayTimeout, "The request timed out").
			WithCode(apierr.CodeGatewayTimeout), true

	case errors.Is(mapped, database.ErrConnectionFailure), errors.Is(mapped, database.ErrLimitExceeded):
		return apierr.New(http.StatusServiceUnavailable, "Service temporarily unavailable").
			WithCode(apierr.CodeServiceUnavailable), true

	case errors.Is(mapped, database.ErrTransactionFailure), errors.Is(mapped, database.ErrLockNotAvailable):
		return apierr.New(http.StatusConflict, "The resource was modified concurrently, please retry").
			WithCode(apierr.CodeResourceConflict), true
	}

	return nil, false
}

// rootCause returns the deepest wrapped cause message when available, falling
// back to the error's own message. Used to keep infrastructure logs informative
// without re-rendering the original APIError's (now-superseded) status.
func rootCause(err error) string {
	if err == nil {
		return ""
	}
	if u := errors.Unwrap(err); u != nil {
		return u.Error()
	}
	return err.Error()
}
