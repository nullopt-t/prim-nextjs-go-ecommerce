package object_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/nullopt-t/prim-backend/internal/object"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// nopLogger silences the error middleware during handler tests.
type nopLogger struct{}

func (nopLogger) Info(string, ...log.Meta)  {}
func (nopLogger) Error(string, ...log.Meta) {}
func (nopLogger) Warn(string, ...log.Meta)  {}
func (nopLogger) Debug(string, ...log.Meta) {}

func uploadRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/media", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// TestUploadMedia_RejectsNonRasterPayloads ensures the handler sniffs content
// rather than trusting the client's filename/Content-Type. An SVG (which can
// execute script when served from the public bucket) disguised as a PNG must be
// rejected before it reaches object storage.
func TestUploadMedia_RejectsNonRasterPayloads(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.ErrorHandler(nopLogger{}))
	router.POST("/media", object.NewHandler(nil).UploadMedia)

	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, uploadRequest(t, "logo.png", svg))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

// TestUploadMedia_RejectsEmptyFile verifies an empty payload is rejected as an
// invalid image rather than being forwarded to the object store.
func TestUploadMedia_RejectsEmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.ErrorHandler(nopLogger{}))
	router.POST("/media", object.NewHandler(nil).UploadMedia)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, uploadRequest(t, "empty.png", nil))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", w.Code, http.StatusBadRequest, w.Body.String())
	}
}
