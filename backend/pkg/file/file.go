// Package file provides helper utilities for file analysis, MIME type detection, and file extensions.
package file

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
)

// DetectContentType reads up to 512 bytes from the multipart file header to determine
// its MIME content type using http.DetectContentType, and restores the file offset.
func DetectContentType(file multipart.File) (string, error) {
	if file == nil {
		return "", fmt.Errorf("file is nil")
	}

	// Capture initial file offset if seekable
	initialOffset, seekErr := file.Seek(0, io.SeekCurrent)

	// Ensure the file offset is restored upon function exit
	defer func() {
		if seekErr == nil {
			_, _ = file.Seek(initialOffset, io.SeekStart)
		} else {
			_, _ = file.Seek(0, io.SeekStart)
		}
	}()

	// Allocate a 512-byte buffer as required by http.DetectContentType sniff algorithm.
	buf := make([]byte, 512)

	// Read up to 512 bytes from file header.
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read file header: %w", err)
	}

	if n == 0 {
		return "", fmt.Errorf("file is empty (0 bytes)")
	}

	detected := http.DetectContentType(buf[:n])

	// Fallback detection for SVG files which may be sniffed as text/xml, text/plain, or application/octet-stream
	if strings.HasPrefix(detected, "text/") || detected == "application/octet-stream" || detected == "text/xml; charset=utf-8" {
		lowerContent := bytes.ToLower(buf[:n])
		if bytes.Contains(lowerContent, []byte("<svg")) {
			detected = "image/svg+xml"
		}
	}

	return detected, nil
}

// AllowedImageMIMETypes are the raster image formats permitted for user uploads.
// SVG is deliberately excluded: SVG is XML-based and can carry JavaScript, which
// leads to stored XSS when such files are served from a public bucket.
var AllowedImageMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/avif": true,
}

// IsAllowedImageMime reports whether the content type is a raster image format
// permitted for upload (SVG and other non-raster types are rejected).
func IsAllowedImageMime(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return AllowedImageMIMETypes[strings.ToLower(mediaType)]
}

// MimeExtension maps MIME types to their conventional file extension including the leading dot.
func MimeExtension(ct string) string {
	cleanCT := strings.TrimSpace(ct)
	if cleanCT == "" {
		return ""
	}

	// Parse media type to strip parameters such as charset (e.g. "image/svg+xml; charset=utf-8")
	mediaType, _, err := mime.ParseMediaType(cleanCT)
	if err != nil {
		mediaType = strings.TrimSpace(strings.Split(cleanCT, ";")[0])
	}
	mediaType = strings.ToLower(mediaType)

	// Direct mappings for preferred extensions
	switch mediaType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/avif":
		return ".avif"
	case "image/svg+xml":
		return ".svg"
	case "image/bmp":
		return ".bmp"
	case "image/tiff":
		return ".tiff"
	case "video/mp4":
		return ".mp4"
	case "video/mkv":
		return ".mkv"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "application/pdf":
		return ".pdf"
	}

	// Fallback to standard library mime.ExtensionsByType
	if exts, err := mime.ExtensionsByType(mediaType); err == nil && len(exts) > 0 {
		return exts[0]
	}

	return ""
}
