package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// Package middleware also provides transparent HTTP response compression.
//
// The API is JSON-first: list, dashboard, audit, and report responses routinely
// span tens to hundreds of kilobytes of highly repetitive text. Negotiating
// gzip on those responses cuts transfer size (and client-observed latency)
// without changing a single handler or the response envelope.

const (
	// minCompressBytes is the smallest response body worth compressing. Below
	// this size the gzip header/trailer overhead can exceed the savings, so
	// small payloads are passed through untouched.
	minCompressBytes = 1024

	// maxBufferedBytes caps how much of a response is buffered to decide whether
	// to compress it. Larger bodies stream through uncompressed instead of being
	// held in memory twice.
	maxBufferedBytes = 16 << 20 // 16 MiB
)

// gzipWriterPool recycles gzip.Writer instances across requests, avoiding a
// fresh deflate window and hash-table allocation per compressed response.
var gzipWriterPool = sync.Pool{
	New: func() any {
		w, err := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		if err != nil {
			// Only an invalid level can fail, and the level is a constant here.
			panic(err)
		}
		return w
	},
}

// Compress transparently gzip-encodes compressible responses when the client
// advertises support via Accept-Encoding. Responses that are already encoded,
// too small, streaming, or of a pre-compressed media type are passed through
// unchanged. It should be registered near the top of the chain so it observes
// the fully rendered body (including error responses written by ErrorHandler).
func Compress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !acceptsGzip(c.Request.Header.Get("Accept-Encoding")) {
			c.Next()
			return
		}

		original := c.Writer
		cw := &compressWriter{ResponseWriter: original}
		c.Writer = cw

		// The body is buffered so the Content-Encoding decision can be made once
		// the status, content type, and final size are known. finish() must run
		// even if a downstream handler panics.
		defer func() {
			cw.finish()
			c.Writer = original
		}()

		c.Next()
	}
}

// compressWriter buffers writes so the middleware can decide, after the handler
// completes, whether the rendered response is worth compressing. Until finish()
// runs no body bytes (and no Content-Encoding header) reach the client.
type compressWriter struct {
	gin.ResponseWriter

	body     bytes.Buffer
	wrote    bool // the handler produced body bytes
	disable  bool // negotiation opted out (flush, oversized body, ...)
	finished bool
}

var _ gin.ResponseWriter = (*compressWriter)(nil)

// Write buffers the payload for later compression negotiation.
func (w *compressWriter) Write(p []byte) (int, error) {
	w.wrote = true
	if w.disable {
		return w.ResponseWriter.Write(p)
	}
	if w.body.Len()+len(p) > maxBufferedBytes {
		return w.disableAndWrite(p)
	}
	return w.body.Write(p)
}

// WriteString buffers the payload, mirroring Write without the []byte copy.
func (w *compressWriter) WriteString(s string) (int, error) {
	w.wrote = true
	if w.disable {
		return w.ResponseWriter.WriteString(s)
	}
	if w.body.Len()+len(s) > maxBufferedBytes {
		return w.disableAndWriteString(s)
	}
	return w.body.WriteString(s)
}

// disableAndWrite gives up on negotiation, emits whatever was buffered, and
// forwards the current write straight through.
func (w *compressWriter) disableAndWrite(p []byte) (int, error) {
	w.disable = true
	w.writePlain(w.body.Bytes())
	w.body = bytes.Buffer{}
	return w.ResponseWriter.Write(p)
}

func (w *compressWriter) disableAndWriteString(s string) (int, error) {
	w.disable = true
	w.writePlain(w.body.Bytes())
	w.body = bytes.Buffer{}
	return w.ResponseWriter.WriteString(s)
}

// Flush signals a streaming response. Compression cannot be negotiated after a
// partial body is already on the wire, so the buffer is emitted as-is.
func (w *compressWriter) Flush() {
	w.disable = true
	w.writePlain(w.body.Bytes())
	w.body = bytes.Buffer{}
	w.ResponseWriter.Flush()
}

// WriteHeaderNow is intentionally deferred: the status line and headers stay
// mutable until finish() so Content-Encoding can still be added. Gin writes the
// header on its own at the end of the request if nothing was flushed here.
func (w *compressWriter) WriteHeaderNow() {}

// Written reports true once the handler has produced body bytes, matching the
// semantics of the wrapped writer (used by ErrorHandler's overwrite guard).
func (w *compressWriter) Written() bool {
	return w.wrote || w.ResponseWriter.Written()
}

// Size reports the number of buffered bytes until the response is finished, at
// which point the wrapped writer holds the true (possibly compressed) size.
func (w *compressWriter) Size() int {
	if w.finished || w.disable {
		return w.ResponseWriter.Size()
	}
	return w.body.Len()
}

// finish commits the buffered response, compressing it when appropriate.
func (w *compressWriter) finish() {
	if w.finished {
		return
	}
	w.finished = true

	body := w.body.Bytes()
	w.body = bytes.Buffer{}

	if w.disable || len(body) < minCompressBytes || !w.shouldCompress() {
		w.writePlain(body)
		return
	}

	header := w.Header()
	header.Set("Content-Encoding", "gzip")
	header.Del("Content-Length")
	addVary(header, "Accept-Encoding")

	w.ResponseWriter.WriteHeaderNow()

	gz := gzipWriterPool.Get().(*gzip.Writer)
	defer gzipWriterPool.Put(gz)
	gz.Reset(w.ResponseWriter)
	_, _ = gz.Write(body)
	_ = gz.Close()
}

// writePlain emits the body unchanged. An empty body is left for gin to finalize
// so a later middleware (e.g. Recovery) can still set the status code.
func (w *compressWriter) writePlain(body []byte) {
	if len(body) == 0 {
		return
	}
	_, _ = w.ResponseWriter.Write(body)
}

// shouldCompress applies the response-level guards: no-body statuses, an
// existing Content-Encoding, a no-transform cache directive, and media types
// that are already compressed.
func (w *compressWriter) shouldCompress() bool {
	switch status := w.Status(); {
	case status < http.StatusOK,
		status == http.StatusNoContent,
		status == http.StatusResetContent,
		status == http.StatusNotModified:
		return false
	}

	header := w.Header()
	if header.Get("Content-Encoding") != "" {
		return false
	}
	if strings.Contains(strings.ToLower(header.Get("Cache-Control")), "no-transform") {
		return false
	}
	return isCompressibleContentType(header.Get("Content-Type"))
}

// acceptsGzip reports whether the Accept-Encoding value permits gzip, honouring
// explicit exclusions (gzip;q=0) and the wildcard token.
func acceptsGzip(header string) bool {
	explicit, wildcard := false, false
	for _, token := range strings.Split(header, ",") {
		name, params, _ := strings.Cut(token, ";")
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" {
			continue
		}

		q := 1.0
		for _, p := range strings.Split(params, ";") {
			p = strings.TrimSpace(p)
			if !strings.HasPrefix(p, "q=") {
				continue
			}
			if v, err := strconv.ParseFloat(strings.TrimSpace(p[2:]), 64); err == nil {
				q = v
			}
		}

		switch name {
		case "gzip", "x-gzip":
			if q > 0 {
				return true
			}
			explicit = true
		case "*":
			if q > 0 {
				wildcard = true
			}
		}
	}
	return wildcard && !explicit
}

// isCompressibleContentType reports whether a media type benefits from gzip.
// Text-like and structured types compress well; images, video, archives, and
// other pre-compressed payloads do not.
func isCompressibleContentType(contentType string) bool {
	mediaType, _, _ := strings.Cut(contentType, ";")
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if mediaType == "" || mediaType == "text/event-stream" {
		return false
	}
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}
	switch mediaType {
	case "application/json",
		"application/javascript",
		"application/x-javascript",
		"application/xml",
		"application/xhtml+xml",
		"application/rss+xml",
		"application/atom+xml",
		"application/graphql",
		"application/x-www-form-urlencoded",
		"image/svg+xml":
		return true
	}
	return strings.HasSuffix(mediaType, "+json") || strings.HasSuffix(mediaType, "+xml")
}

// addVary appends a value to the Vary header without duplicating an existing
// token, so caches key on Accept-Encoding once a response is compressed.
func addVary(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for _, part := range strings.Split(existing, ",") {
			part = strings.TrimSpace(part)
			if part == "*" || strings.EqualFold(part, value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}
