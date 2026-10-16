package api_test

import (
	"encoding/json"
	"testing"

	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessResponseJSON(t *testing.T) {
	resp := api.NewSuccessResponse([]string{}, "success")
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"data":[]`)
	assert.Contains(t, string(data), `"message":"success"`)
}

func TestPaginatedResponseJSON(t *testing.T) {
	resp := api.NewPaginatedResponse(nil, nil)
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"data":[]`)
}

func TestBadRequestErrorResponseAlias(t *testing.T) {
	var errResp api.BadRequestErrorResponse
	errResp.Code = "TEST"
	var badReq api.BadReqResponse = errResp
	assert.Equal(t, "TEST", badReq.Code)
}
