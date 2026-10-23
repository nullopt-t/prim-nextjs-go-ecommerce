package audit

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

// Handler exposes the audit trail over HTTP.
type Handler struct {
	service *Service
}

// NewHandler constructs an audit HTTP handler.
func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

type auditEntryResponse struct {
	ID         string          `json:"id"`
	ActorID    *string         `json:"actorId"`
	ActorRole  string          `json:"actorRole"`
	Action     string          `json:"action"`
	EntityType string          `json:"entityType"`
	EntityID   string          `json:"entityId"`
	Method     string          `json:"method"`
	Path       string          `json:"path"`
	StatusCode int             `json:"statusCode"`
	IPAddress  string          `json:"ipAddress"`
	UserAgent  string          `json:"userAgent"`
	Before     json.RawMessage `json:"before,omitempty" swaggertype:"object"`
	After      json.RawMessage `json:"after,omitempty" swaggertype:"object"`
	Metadata   json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt  string          `json:"createdAt"`
}

// ListAuditLogs godoc
//
//	@Summary		List audit trail
//	@Description	Returns a filterable, keyset-paginated view of privileged admin mutations (superadmin only).
//	@Tags			Admin Audit
//	@Accept			json
//	@Produce		json
//	@Param			actorId		query		string	false	"Filter by actor user UUID"
//	@Param			action		query		string	false	"Filter by action name (e.g. user.role.update)"
//	@Param			entityType	query		string	false	"Filter by entity type (e.g. user)"
//	@Param			entityId	query		string	false	"Filter by entity id"
//	@Param			from		query		string	false	"Only entries at/after this RFC3339 timestamp"
//	@Param			to			query		string	false	"Only entries at/before this RFC3339 timestamp"
//	@Param			cursor		query		string	false	"Opaque pagination cursor"
//	@Param			pageSize	query		int		false	"Page size (default 50, max 100)"
//	@Success		200	{object}	api.DataResponse{data=[]auditEntryResponse}	"Audit trail"
//	@Failure		400	{object}	api.BadRequestErrorResponse	"Invalid filter or cursor"
//	@Failure		403	{object}	api.ForbiddenErrorResponse	"Superadmin privileges required"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/admin/audit [get]
func (h *Handler) ListAuditLogs(c *gin.Context) {
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

	filter, err := parseFilter(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	entries, page, err := h.service.List(c.Request.Context(), filter, q.PageSize, cursor)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": toResponses(entries),
		"meta": page,
	})
}

func parseFilter(c *gin.Context) (Filter, error) {
	var f Filter

	if v := c.Query("actorId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return f, apierr.BadRequestError("Invalid actorId").WithCode(apierr.CodeInvalidInput)
		}
		f.ActorID = &id
	}

	f.Action = c.Query("action")
	f.EntityType = c.Query("entityType")
	f.EntityID = c.Query("entityId")

	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, apierr.BadRequestError("Invalid 'from' timestamp (RFC3339 expected)").WithCode(apierr.CodeInvalidInput)
		}
		f.From = &t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, apierr.BadRequestError("Invalid 'to' timestamp (RFC3339 expected)").WithCode(apierr.CodeInvalidInput)
		}
		f.To = &t
	}

	return f, nil
}

func toResponses(entries []*model.AuditLog) []auditEntryResponse {
	out := make([]auditEntryResponse, 0, len(entries))
	for _, e := range entries {
		var actorID *string
		if e.ActorID != nil {
			s := e.ActorID.String()
			actorID = &s
		}
		out = append(out, auditEntryResponse{
			ID:         strconv.FormatInt(e.ID, 10),
			ActorID:    actorID,
			ActorRole:  e.ActorRole,
			Action:     e.Action,
			EntityType: e.EntityType,
			EntityID:   e.EntityID,
			Method:     e.Method,
			Path:       e.Path,
			StatusCode: e.StatusCode,
			IPAddress:  e.IPAddress,
			UserAgent:  e.UserAgent,
			Before:     json.RawMessage(e.Before),
			After:      json.RawMessage(e.After),
			Metadata:   json.RawMessage(e.Metadata),
			CreatedAt:  e.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}
