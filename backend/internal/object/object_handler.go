package object

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

type ObjectHandler struct {
	service *ObjectService
}

func NewHandler(service *ObjectService) *ObjectHandler {
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
		_ = c.Error(apierr.InternalServerError("Failed to fetch media").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": media,
		"meta": page,
	})
}

func (h *ObjectHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apierr.BadRequestError("file is required").WithCode(apierr.CodeInvalidInput))
		return
	}

	src, err := file.Open()
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to read uploaded file").Wrap(err))
		return
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	bucket := "catalog"
	obj, err := h.service.UploadObject(c.Request.Context(), contentType, file.Size, bucket, src)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to upload media").Wrap(err))
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
