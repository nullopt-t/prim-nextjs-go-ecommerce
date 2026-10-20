package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
)

const (
	// adminPrefix scopes automatic recording to back-office API routes.
	adminPrefix = "/api/v1/admin"
	// maxBodyBytes caps a captured request payload so oversized bodies are never
	// buffered into memory.
	maxBodyBytes = 32 * 1024
	// recordTimeout bounds the detached audit write.
	recordTimeout = 3 * time.Second
)

// sensitiveKeys are request-body fields whose values are replaced before an
// audit payload is stored.
var sensitiveKeys = map[string]struct{}{
	"password":            {},
	"currentpassword":     {},
	"newpassword":         {},
	"oldpassword":         {},
	"passwordconfirm":     {},
	"token":               {},
	"accesstoken":         {},
	"refreshtoken":        {},
	"secret":              {},
	"clientsecret":        {},
	"webhooksecret":       {},
	"stripewebhooksecret": {},
	"apikey":              {},
	"api_key":             {},
	"authorization":       {},
	"otp":                 {},
	"otpcode":             {},
	"recoverycode":        {},
}

// Middleware records privileged admin mutations (any non-safe method under
// /api/v1/admin) into the audit trail. It wraps the request so it can read the
// authenticated actor from the gin context after the downstream auth middleware
// and handler have run.
func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuditable(c) {
			c.Next()
			return
		}

		payload := capturePayload(c)

		c.Next()

		actorID, role := actor(c)
		if actorID == nil {
			// Unauthenticated attempt (e.g. rejected at the auth gate): there is
			// no attributable actor, so nothing to record.
			return
		}

		action, entityType, entityID := resolveAction(c)
		entry := &model.AuditLog{
			ActorID:    actorID,
			ActorRole:  role,
			Action:     action,
			EntityType: entityType,
			EntityID:   entityID,
			Method:     c.Request.Method,
			Path:       c.FullPath(),
			StatusCode: c.Writer.Status(),
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			After:      payload,
			Metadata:   routeMetadata(c),
			CreatedAt:  time.Now().UTC(),
		}

		// Detach from the request lifecycle: the client may have gone away by the
		// time we persist, and an audit failure must not surface as a request error.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), recordTimeout)
		defer cancel()
		_ = s.Record(ctx, entry)
	}
}

// isAuditable reports whether the request mutates admin state.
func isAuditable(c *gin.Context) bool {
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	return path == adminPrefix || strings.HasPrefix(path, adminPrefix+"/")
}

// capturePayload reads and restores a JSON request body, returning a redacted copy.
func capturePayload(c *gin.Context) []byte {
	if !strings.Contains(c.ContentType(), "application/json") {
		return nil
	}
	raw, err := readAndRestoreBody(c)
	if err != nil || len(raw) == 0 {
		return nil
	}
	return redactJSON(raw)
}

// actor extracts the authenticated user id and role from the gin context. The
// auth middleware stores the id as a uuid.UUID (or, defensively, a string).
func actor(c *gin.Context) (*uuid.UUID, string) {
	idVal, ok := c.Get("userID")
	if !ok {
		return nil, ""
	}

	var id uuid.UUID
	switch v := idVal.(type) {
	case uuid.UUID:
		id = v
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, ""
		}
		id = parsed
	default:
		return nil, ""
	}
	if id == uuid.Nil {
		return nil, ""
	}

	role, _ := c.Get("userRole")
	roleStr, _ := role.(string)
	return &id, roleStr
}

// resolveAction derives a normalized action name, entity type, and entity id
// from the matched route.
func resolveAction(c *gin.Context) (action, entityType, entityID string) {
	full := c.FullPath()
	if full == "" {
		full = c.Request.URL.Path
	}
	segs := strings.Split(strings.Trim(full, "/"), "/")

	idx := -1
	for i, s := range segs {
		if s == "admin" {
			idx = i
			break
		}
	}
	entityID = entityParam(c)
	if idx == -1 || idx+1 >= len(segs) {
		return strings.ToLower(c.Request.Method), "", entityID
	}

	resource := segs[idx+1]
	sub := ""
	rest := segs[idx+2:]
	if len(rest) > 0 {
		if last := rest[len(rest)-1]; !strings.HasPrefix(last, ":") {
			sub = last
		}
	}

	entityType = singularize(resource)
	verb := methodVerb(c.Request.Method)

	switch {
	case sub == "":
		action = entityType + "." + verb
	case verb == "create":
		// e.g. POST /admin/payments/:id/refund -> payment.refund
		action = entityType + "." + sub
	default:
		action = entityType + "." + sub + "." + verb
	}
	return action, entityType, entityID
}

func entityParam(c *gin.Context) string {
	for _, key := range []string{"id", "productId", "variantId", "variant_id", "media_id"} {
		if v := c.Param(key); v != "" {
			return v
		}
	}
	return ""
}

func routeMetadata(c *gin.Context) []byte {
	if len(c.Params) == 0 {
		return nil
	}
	m := make(map[string]string, len(c.Params))
	for _, p := range c.Params {
		m[p.Key] = p.Value
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

func methodVerb(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

// singularize converts a route resource segment to a stable entity type.
func singularize(resource string) string {
	switch resource {
	case "media":
		return "media"
	case "inventory":
		return "inventory"
	case "settings":
		return "settings"
	}
	return strings.TrimSuffix(resource, "s")
}

func readAndRestoreBody(c *gin.Context) ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}
	// Admin JSON payloads are small; read the whole body, then restore it so the
	// downstream handler can still bind it. Only the stored copy is capped.
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) > maxBodyBytes {
		raw = raw[:maxBodyBytes]
	}
	return raw, nil
}

// redactJSON replaces sensitive field values with a placeholder.
func redactJSON(raw []byte) []byte {
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil
	}
	redacted := redactValue(decoded)
	out, err := json.Marshal(redacted)
	if err != nil {
		return nil
	}
	return out
}

func redactValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			if _, sensitive := sensitiveKeys[strings.ToLower(k)]; sensitive {
				val[k] = "[REDACTED]"
				continue
			}
			val[k] = redactValue(child)
		}
		return val
	case []any:
		for i, child := range val {
			val[i] = redactValue(child)
		}
		return val
	default:
		return v
	}
}
