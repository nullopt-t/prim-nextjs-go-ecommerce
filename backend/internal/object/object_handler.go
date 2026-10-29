package object

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	fileutil "github.com/nullopt-t/prim-backend/pkg/file"
)

// MaxMediaUploadBytes caps a single media upload at 10 MiB. Product/brand media
// are raster images, which are far below this; the cap protects memory and the
// object store from oversized or abusive uploads.
const MaxMediaUploadBytes int64 = 10 << 20

type ObjectHandler struct {
	service *ObjectService
}

func NewHandler(service *ObjectService) *ObjectHandler {
	if service == nil {
		panic("object: service is required")
	}
	return &ObjectHandler{service: service}
}

func (h *ObjectHandler) ListMedia(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 50, MaxPageSize: 100})

	cursor, err := q.Keyset()
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	media, page, err := h.service.ListMedia(c.Request.Context(), q.PageSize, cursor)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": media,
		"meta": page,
	})
}

func (h *ObjectHandler) UploadMedia(c *gin.Context) {
	// Cap the total request body before parsing the multipart form so an
	// oversized upload is rejected without buffering it in memory/temp files.
	// Leave headroom for multipart boundaries/headers so a file right at the
	// 10 MiB limit is not rejected by transport overhead; the exact file size
	// is enforced separately below.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxMediaUploadBytes+(1<<20))

	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apierr.BadRequestError("file is required").WithCode(apierr.CodeInvalidInput))
		return
	}

	if file.Size > MaxMediaUploadBytes {
		_ = c.Error(apierr.BadRequestError("File exceeds the 10 MB upload limit").
			WithCode(apierr.CodeInvalidInput))
		return
	}

	src, err := file.Open()
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to read uploaded file").Wrap(err))
		return
	}
	defer func() { _ = src.Close() }()

	// Trust the bytes, not the client: sniff the real content type and enforce a
	// raster-image allowlist. This rejects SVG (which can carry scripts and cause
	// stored XSS when served from the public bucket) and any other disguised
	// payload regardless of the Content-Type header the client claims.
	detected, err := fileutil.DetectContentType(src)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Failed to inspect uploaded file").
			WithCode(apierr.CodeInvalidInput).Wrap(err))
		return
	}
	if !fileutil.IsAllowedImageMime(detected) {
		_ = c.Error(apierr.BadRequestError("Unsupported file type; upload a JPEG, PNG, GIF, WebP or AVIF image").
			WithCode(apierr.CodeInvalidInput))
		return
	}
	contentType := detected

	bucket := "catalog"
	obj, err := h.service.UploadObject(c.Request.Context(), contentType, file.Size, bucket, src)
	if err != nil {
		_ = c.Error(err)
		return
	}

	item := MediaItem{
		ID:          obj.ID,
		Bucket:      obj.Bucket,
		ObjectKey:   obj.Key,
		ContentType: obj.ContentType,
		FileSize:    obj.FileSize,
		Status:      string(obj.Status),
		URL:         h.service.GetObjectURL(c.Request.Context(), obj.Bucket, obj.Key),
		CreatedAt:   obj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: item})
}

// DeleteMedia soft-deletes a media object by ID and removes it from object
// storage. Mirrors the lifecycle used by DeleteObject: mark 'deleting', remove
// the binary from the bucket, then mark 'deleted'.
func (h *ObjectHandler) DeleteMedia(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid media ID").WithCode(apierr.CodeInvalidInput))
		return
	}

	if err := h.service.DeleteObjectByID(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Media deleted"})
}
