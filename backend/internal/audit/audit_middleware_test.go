package audit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext(method, path string, params gin.Params) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(method, path, nil)
	c.Params = params
	return c
}

func TestResolveAction(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		path       string
		params     gin.Params
		wantAction string
		wantType   string
		wantID     string
	}{
		{
			name:   "role update",
			method: http.MethodPatch, path: "/api/v1/admin/users/:id/role",
			params:     gin.Params{{Key: "id", Value: "u1"}},
			wantAction: "user.role.update", wantType: "user", wantID: "u1",
		},
		{
			name:   "status update",
			method: http.MethodPatch, path: "/api/v1/admin/users/:id/status",
			params:     gin.Params{{Key: "id", Value: "u1"}},
			wantAction: "user.status.update", wantType: "user", wantID: "u1",
		},
		{
			name:   "user delete",
			method: http.MethodDelete, path: "/api/v1/admin/users/:id",
			params:     gin.Params{{Key: "id", Value: "u1"}},
			wantAction: "user.delete", wantType: "user", wantID: "u1",
		},
		{
			name:   "media delete",
			method: http.MethodDelete, path: "/api/v1/admin/media/:id",
			params:     gin.Params{{Key: "id", Value: "m1"}},
			wantAction: "media.delete", wantType: "media", wantID: "m1",
		},
		{
			name:   "media upload",
			method: http.MethodPost, path: "/api/v1/admin/media",
			wantAction: "media.create", wantType: "media", wantID: "",
		},
		{
			name:   "payment refund",
			method: http.MethodPost, path: "/api/v1/admin/payments/:id/refund",
			params:     gin.Params{{Key: "id", Value: "p1"}},
			wantAction: "payment.refund", wantType: "payment", wantID: "p1",
		},
		{
			name:   "revoke sessions",
			method: http.MethodPost, path: "/api/v1/admin/users/:id/revoke-sessions",
			params:     gin.Params{{Key: "id", Value: "u1"}},
			wantAction: "user.revoke-sessions", wantType: "user", wantID: "u1",
		},
		{
			name:   "variant adjust uses productId param",
			method: http.MethodPost, path: "/api/v1/admin/products/:productId/adjust",
			params:     gin.Params{{Key: "productId", Value: "prod1"}},
			wantAction: "product.adjust", wantType: "product", wantID: "prod1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestContext(tc.method, tc.path, tc.params)
			action, entityType, entityID := resolveAction(c)
			assert.Equal(t, tc.wantAction, action)
			assert.Equal(t, tc.wantType, entityType)
			assert.Equal(t, tc.wantID, entityID)
		})
	}
}

func TestIsAuditable(t *testing.T) {
	cases := []struct {
		method string
		path   string
		want   bool
	}{
		{http.MethodPatch, "/api/v1/admin/users/u1/status", true},
		{http.MethodDelete, "/api/v1/admin/media/m1", true},
		{http.MethodPost, "/api/v1/admin/promotions", true},
		{http.MethodPut, "/api/v1/admin/settings", true},
		{http.MethodGet, "/api/v1/admin/users", false},
		{http.MethodHead, "/api/v1/admin/users", false},
		{http.MethodOptions, "/api/v1/admin/users", false},
		{http.MethodPost, "/api/v1/checkout", false},
		{http.MethodPost, "/api/v1/adminFake", false},
	}
	for _, tc := range cases {
		c := newTestContext(tc.method, tc.path, nil)
		assert.Equalf(t, tc.want, isAuditable(c), "%s %s", tc.method, tc.path)
	}
}

func TestRedactJSON(t *testing.T) {
	raw := []byte(`{
		"password": "hunter2",
		"keep": "visible",
		"nested": {"token": "abc", "apiKey": "k", "email": "a@b.c"},
		"list": [{"secret": "s1"}, {"ok": true}]
	}`)

	out := redactJSON(raw)
	require.NotNil(t, out)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(out, &decoded))

	assert.Equal(t, "[REDACTED]", decoded["password"])
	assert.Equal(t, "visible", decoded["keep"])

	nested := decoded["nested"].(map[string]any)
	assert.Equal(t, "[REDACTED]", nested["token"])
	assert.Equal(t, "[REDACTED]", nested["apiKey"])
	assert.Equal(t, "a@b.c", nested["email"])

	list := decoded["list"].([]any)
	assert.Equal(t, "[REDACTED]", list[0].(map[string]any)["secret"])
	assert.Equal(t, true, list[1].(map[string]any)["ok"])
}

func TestRedactJSONInvalidReturnsNil(t *testing.T) {
	assert.Nil(t, redactJSON([]byte("not json")))
}

func TestSingularize(t *testing.T) {
	assert.Equal(t, "user", singularize("users"))
	assert.Equal(t, "payment", singularize("payments"))
	assert.Equal(t, "media", singularize("media"))
	assert.Equal(t, "inventory", singularize("inventory"))
}
