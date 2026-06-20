package role

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
)

type RoleHandler struct {
	roleService *RoleService
}

func NewRoleHandler(
	roleService *RoleService,
) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

type RoleResponse struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	CreatedAt string `json:"created_at"`
}

// GetAll godoc
// @Summary fetch all the roles using pagination
// @Description fetch all the roles using pagination
// @Tags Roles
// @Accept json
// @Produce json
// @Param query query api.PageQuery true "query"
// @Success 200 {object} api.PaginatedResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /roles [get]
func (h *RoleHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	var q api.PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if q.PageSize == 0 {
		q.PageSize = 5
	}

	if q.Page == 0 {
		q.Page = 1
	}

	roles, page, err := h.roleService.GetAll(ctx, q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*RoleResponse, 0, len(roles))
	for _, role := range roles {
		res = append(res, &RoleResponse{
			ID:        role.ID,
			Code:      role.Code,
			CreatedAt: role.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: res,
		Meta: page,
	})
}
