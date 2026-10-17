package system

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/pkg/api"
)

// reportTimeout bounds the whole health report so the handler always responds
// promptly, even if several probes are slow.
const reportTimeout = 5 * time.Second

// Handler exposes administrative infrastructure health endpoints over HTTP.
type Handler struct {
	service *Service
}

// NewHandler constructs a system health HTTP handler.
func NewHandler(s *Service) *Handler {
	return &Handler{
		service: s,
	}
}

// GetHealth godoc
//
//	@Summary		Get core services health
//	@Description	Probes PostgreSQL, Redis, Kafka, and MinIO and returns live connection status, latency, and queue lag.
//	@Tags			Admin System
//	@Accept			json
//	@Produce		json
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Authentication required"
//	@Failure		403	{object}	api.ForbiddenErrorResponse		"Admin privileges required"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=HealthReport}	"Core services health report"
//	@Router			/admin/system/health [get]
func (h *Handler) GetHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), reportTimeout)
	defer cancel()

	report := h.service.Check(ctx)

	// The report is always 200: individual dependency failures are represented in
	// the per-service statuses so the client can render a full picture.
	c.JSON(http.StatusOK, api.DataResponse{Data: report})
}
