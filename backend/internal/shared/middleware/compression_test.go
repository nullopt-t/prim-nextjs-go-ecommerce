package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopLogger struct{}

func (nopLogger) Info(string, ...log.Meta)  {}
func (nopLogger) Error(string, ...log.Meta) {}
func (nopLogger) Warn(string, ...log.Meta)  {}
func (nopLogger) Debug(string, ...log.Meta) {}

// serve runs a single request through a gin engine with the compression
// middleware and the given handler.
func serve(t *testing.T, acceptEncoding string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.Compress())
	r.GET("/test", handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if acceptEncoding != "" {
		req.Header.Set("Accept-Encoding", acceptEncoding)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeGzip(t *testing.T, body []byte) string {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer zr.Close()

	raw, err := io.ReadAll(zr)
	require.NoError(t, err)
	return string(raw)
}

func TestCompress_LargeJSONIsGzipped(t *testing.T) {
	payload := strings.Repeat("product-name ", 400)

	w := serve(t, "gzip, deflate, br", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": payload})
	})

	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Contains(t, w.Header().Values("Vary"), "Accept-Encoding")
	assert.Less(t, w.Body.Len(), len(payload), "compressed body should be smaller")

	decoded := decodeGzip(t, w.Body.Bytes())
	assert.Contains(t, decoded, payload)
}

func TestCompress_NoAcceptEncodingLeavesBodyPlain(t *testing.T) {
	payload := strings.Repeat("product-name ", 400)

	w := serve(t, "", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": payload})
	})

	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Contains(t, w.Body.String(), payload)
}

func TestCompress_SmallBodyIsNotCompressed(t *testing.T) {
	w := serve(t, "gzip", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestCompress_RespectsQualityZero(t *testing.T) {
	payload := strings.Repeat("product-name ", 400)

	w := serve(t, "gzip;q=0, deflate", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": payload})
	})

	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Contains(t, w.Body.String(), payload)
}

func TestCompress_WildcardEnablesGzip(t *testing.T) {
	payload := strings.Repeat("product-name ", 400)

	w := serve(t, "*", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": payload})
	})

	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Contains(t, decodeGzip(t, w.Body.Bytes()), payload)
}

func TestCompress_SkipsPreCompressedTypes(t *testing.T) {
	blob := bytes.Repeat([]byte{0x89, 0x50, 0x4e, 0x47}, 2000)

	w := serve(t, "gzip", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/png", blob)
	})

	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Equal(t, blob, w.Body.Bytes())
}

func TestCompress_SkipsAlreadyEncodedResponse(t *testing.T) {
	payload := strings.Repeat("data ", 400)

	w := serve(t, "gzip", func(c *gin.Context) {
		c.Header("Content-Encoding", "br")
		c.Data(http.StatusOK, "application/json", []byte(payload))
	})

	assert.Equal(t, "br", w.Header().Get("Content-Encoding"))
}

func TestCompress_NoBodyStatusIsPreserved(t *testing.T) {
	w := serve(t, "gzip", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"))
	assert.Empty(t, w.Body.String())
}

func TestCompress_CompressesCSVExport(t *testing.T) {
	csv := strings.Repeat("order-123,customer@example.com,paid,120.00\n", 200)

	w := serve(t, "gzip", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(csv))
	})

	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Equal(t, csv, decodeGzip(t, w.Body.Bytes()))
}

func TestCompress_CompressesErrorResponses(t *testing.T) {
	payload := strings.Repeat("x", 4096)

	w := serve(t, "gzip", func(c *gin.Context) {
		c.JSON(http.StatusBadGateway, gin.H{"error": payload})
	})

	assert.Equal(t, http.StatusBadGateway, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Contains(t, decodeGzip(t, w.Body.Bytes()), payload)
}

// The real chain registers Compress before ErrorHandler, which renders the error
// body only after c.Next() returns. This verifies that late write is still
// captured and compressed.
func TestCompress_CompressesErrorHandlerOutput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.Compress())
	r.Use(middleware.ErrorHandler(nopLogger{}))
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(apierr.InternalServerError(strings.Repeat("boom ", 1024)).WithStack())
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	assert.Contains(t, decodeGzip(t, w.Body.Bytes()), "boom")
}
