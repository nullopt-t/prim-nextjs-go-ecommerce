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
	Slug    string `json:"slug" binding:"required" example:"apple"`
	LogoURL string `json:"logo_url" binding:"required" example:"https://pictures.com/apple.png"`
	LogoAlt string `json:"logo_alt" binding:"required" example:"apple logo"`
}

type BrandResponse struct {
	ID        uuid.UUID `json:"id,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	Name      string    `json:"name,omitempty" example:"nvidia"`
	Slug      string    `json:"slug,omitempty" example:"nvidia"`
	LogoURL   string    `json:"logo_url,omitempty" example:"https://pictures.com/nvidia.png"`
	LogoAlt   string    `json:"logo_alt,omitempty" example:"nvidia logo"`
	CreatedAt string    `json:"created_at,omitempty" example:"2026-07-01T05:04:38Z"`
	UpdatedAt string    `json:"udpated_at,omitempty" example:"2026-07-01T05:04:38Z"`
	UpdatedBy uuid.UUID `json:"updated_by,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	CreatedBy uuid.UUID `json:"created_by,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
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
	req := &CreateBrandRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &CreateBrandInput{
		Name:    req.Name,
		Slug:    req.Slug,
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
				Slug:      brand.Slug,
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
	param := &BrandIDParam{}
	if err := c.ShouldBindUri(param); err != nil {
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
				Slug:      brand.Slug,
				LogoURL:   brand.LogoURL,
				LogoAlt:   brand.LogoAlt,
				UpdatedAt: brand.UpdatedAt.Format(time.RFC3339),
				CreatedAt: brand.CreatedAt.Format(time.RFC3339),
			},
		},
	)
}

type BrandSlugParam struct {
	Slug string `uri:"slug" binding:"required" example:"apple"`
}

// GetBrandBySlug godoc
// @Summary get a brand by slug
// @Description get a brand by slug
// @Tags Brand
// @Accept json
// @Produce json
// @Param slug path string true "brand slug"
// @Failure 500 {object} api.ErrorResponse
// @Failure 200 {object} api.DataResponse{data=BrandResponse}
// @Router /brands/{slug} [get]
func (bh *BrandHandler) GetBrandBySlug(c *gin.Context) {
	param := &BrandSlugParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	brand, err := bh.bservice.GetBrandBySlug(
		ctx,
		param.Slug,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: BrandResponse{
				ID:      brand.ID,
				Name:    brand.Name,
				LogoURL: brand.LogoURL,
				LogoAlt: brand.LogoAlt,
			},
		},
	)
}

type UpdateBrandRequest struct {
	Name    *string `json:"name"  example:"apple"`
	LogoURL *string `json:"logo_url"  example:"https://example.com/logo.png"`
	LogoAlt *string `json:"logo_alt"  example:"apple logo"`
}

// UpdateBrand godoc
// @Summary update a brand
// @Description update a brand
// @Tags Brand
// @Accept json
// @Produce json
// @Param id path string true "brand id"
// @Param input body UpdateBrandRequest true "brand input"
// @Failure 500 {object} api.ErrorResponse
// @Failure 200 {object} api.DataResponse{data=BrandResponse}
// @Router /brands/{slug} [put]
func (bh *BrandHandler) UpdateBrand(c *gin.Context) {
	req := &UpdateBrandRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	param := &BrandIDParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	brandID, err := uuid.Parse(param.ID)
	if err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &UpdateBrandInput{
		Name:    req.Name,
		LogoURL: req.LogoURL,
		LogoAlt: req.LogoAlt,
	}

	err = bh.bservice.UpdateBrand(
		c.Request.Context(),
		brandID,
		in,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: in,
	})
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

// ListAdminBrands godoc
// @Summary list all the brands in pages
// @Tags Brand
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "page query"
// @Failure 500 {object} api.ErrorResponse
// @Failure 200 {object} api.PaginatedResponse{data=[]BrandResponse,meta=api.Page}
// @Router /admin/brands [get]
func (bh *BrandHandler) ListAdminBrands(c *gin.Context) {
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
