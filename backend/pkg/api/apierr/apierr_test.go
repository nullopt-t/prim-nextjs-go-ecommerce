package apierr_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilReceiverSafety(t *testing.T) {
	var err *apierr.APIError

	// Calling builder methods on a nil *APIError is an invalid internal programmer state
	// and should panic / fail visibly rather than silently returning nil.
	assert.Panics(t, func() {
		_ = err.WithCode("TEST")
	})
	assert.Panics(t, func() {
		_ = err.WithMessage("msg")
	})
	assert.Panics(t, func() {
		_ = err.Wrap(errors.New("cause"))
	})
}


func TestWithMessagePreservesCauseAndStack(t *testing.T) {
	cause := errors.New("underlying cause")
	err := apierr.New(http.StatusBadRequest, "initial").
		Wrap(cause).
		WithStack()

	require.Equal(t, cause, err.Unwrap())
	require.NotEmpty(t, err.Stack())

	err.WithMessage("updated message")
	assert.Equal(t, "updated message", err.Error())
	assert.Equal(t, cause, err.Unwrap())
	assert.NotEmpty(t, err.Stack())
}

func TestConstructors(t *testing.T) {
	refErr := apierr.InvalidReferenceError()
	assert.Equal(t, http.StatusBadRequest, refErr.Status)
	assert.Equal(t, apierr.CodeInvalidReference, refErr.Code)

	queryErr := apierr.InvalidQueryParameterError()
	assert.Equal(t, http.StatusBadRequest, queryErr.Status)
	assert.Equal(t, apierr.CodeInvalidQueryParameter, queryErr.Code)

	fileReqErr := apierr.FileRequiredError()
	assert.Equal(t, http.StatusBadRequest, fileReqErr.Status)
	assert.Equal(t, apierr.CodeFileRequired, fileReqErr.Code)

	fileTooLargeErr := apierr.FileTooLargeError()
	assert.Equal(t, http.StatusRequestEntityTooLarge, fileTooLargeErr.Status)
	assert.Equal(t, apierr.CodeFileTooLarge, fileTooLargeErr.Code)

	unsupportedMediaErr := apierr.UnsupportedMediaTypeError()
	assert.Equal(t, http.StatusUnsupportedMediaType, unsupportedMediaErr.Status)
	assert.Equal(t, apierr.CodeUnsupportedMediaType, unsupportedMediaErr.Code)

	storageErr := apierr.StorageError()
	assert.Equal(t, http.StatusInternalServerError, storageErr.Status)
	assert.Equal(t, apierr.CodeStorageError, storageErr.Code)
}
