package brand

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
)

type BrandHandler struct {
	bservice *BrandService
}

func NewHandler(
	sb *BrandService,
) *BrandHandler {
	return &BrandHandler{
		bservice: sb,
	}
}

type CreateBrandRequest struct {
	Name    string `json:"name" binding:"required" example:"apple"`
	LogoURL string `json:"logo_url" binding:"required" example:"https://pictures.com/apple.png"`
	LogoAlt string `json:"logo_alt" binding:"required" example:"apple logo"`
}

type BrandResponse struct {
	ID        uuid.UUID `json:"id" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	Name      string    `json:"name" example:"nvidia"`
	LogoURL   string    `json:"logo_url" example:"https://pictures.com/nvidia.png"`
	LogoAlt   string    `json:"logo_alt" example:"nvidia logo"`
	CreatedAt string    `json:"created_at" example:"2026-07-01T05:04:38Z"`
	UpdatedAt string    `json:"udpated_at" example:"2026-07-01T05:04:38Z"`
}

// CreateBrand godoc
// @Summary create a new products brand
// @Description create a new products brand
// @Tags Brand
// @Accept json
// @Produce json
// @Param brand body CreateBrandRequest true "brand data"
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=BrandResponse}
// @Router /brands [post]
func (bh *BrandHandler) CreateBrand(c *gin.Context) {
	var req CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &CreateBrandInput{
		Name:    req.Name,
		LogoURL: req.LogoURL,
		LogoAlt: req.LogoAlt,
	}

	ctx := c.Request.Context()
	brand, err := bh.bservice.CreateBrand(
		ctx,
		in,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusCreated,
		api.DataResponse{
			Data: BrandResponse{
				ID:        brand.ID,
				Name:      brand.Name,
				LogoURL:   brand.LogoURL,
				LogoAlt:   brand.LogoAlt,
				CreatedAt: brand.CreatedAt.Format(time.RFC3339),
				UpdatedAt: brand.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

type BrandIDParam struct {
	ID string `uri:"id" binding:"uuid"`
}

// GetBrandByID godoc
// @Summary get brand by id
// @Description get brand by id
// @Tags Brand
// @Accept json
// @Produce json
// @Param id path BrandIDParam true "Brand id(uuid)"
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=BrandResponse}
// @Router /brands/{id} [get]
func (bh *BrandHandler) GetBrandByID(c *gin.Context) {
	var param BrandIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	brandID, err := uuid.Parse(param.ID)
	if err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	brand, err := bh.bservice.GetBrandByID(
		ctx,
		brandID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: BrandResponse{
				ID:        brand.ID,
				Name:      brand.Name,
				LogoURL:   brand.LogoURL,
				LogoAlt:   brand.LogoAlt,
				CreatedAt: brand.CreatedAt.Format(time.RFC3339),
				UpdatedAt: brand.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

// ListBrands godoc
// @Summary list all the brands in pages
// @Description list all the brands in pages
// @Tags Brand
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "page query"
// @Failure 500 {object} api.ErrorResponse
// @Failure 200 {object} api.PaginatedResponse{data=[]BrandResponse,meta=api.Page}
// @Router /brands [get]
func (bh *BrandHandler) ListBrands(c *gin.Context) {
	var q api.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if q.Page == 0 {
		q.Page = 1
	}

	if q.PageSize == 0 {
		q.PageSize = 10
	}

	ctx := c.Request.Context()
	brands, page, err := bh.bservice.List(
		ctx,
		&q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*BrandResponse, 0, len(brands))
	for _, brand := range brands {
		res = append(res, &BrandResponse{
			ID:        brand.ID,
			Name:      brand.Name,
			LogoURL:   brand.LogoURL,
			LogoAlt:   brand.LogoAlt,
			CreatedAt: brand.CreatedAt.Format(time.RFC3339),
			UpdatedAt: brand.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: page,
		},
	)
}
