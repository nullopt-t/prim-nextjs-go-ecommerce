package brand

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
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
	Name    string  `json:"name" binding:"required" example:"apple"`
	Slug    *string `json:"slug" example:"apple"`
	LogoURL string  `json:"logo_url" binding:"required" example:"https://pictures.com/apple.png"`
	LogoAlt string  `json:"logo_alt" binding:"required" example:"apple logo"`
}

type BrandResponse struct {
	ID        string `json:"id,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	Name      string `json:"name,omitempty" example:"nvidia"`
	Slug      string `json:"slug,omitempty" example:"nvidia"`
	Status    string `json:"status,omitempty" example:"active"`
	LogoURL   string `json:"logo_url,omitempty" example:"https://pictures.com/nvidia.png"`
	LogoAlt   string `json:"logo_alt,omitempty" example:"nvidia logo"`
	CreatedAt string `json:"created_at,omitempty" example:"2026-07-01T05:04:38Z"`
	UpdatedAt string `json:"updated_at,omitempty" example:"2026-07-01T05:04:38Z"`
	// UpdatedBy string `json:"updated_by,omitzero" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	// CreatedBy string `json:"created_by,omitzero" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
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
		LogoURL: req.LogoURL,
		LogoAlt: req.LogoAlt,
	}

	if req.Slug != nil {
		in.Slug = *req.Slug
	} else {
		in.Slug = fmt.Sprintf("%s-%d", utils.Slugify(in.Name), time.Now().Unix())
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
				ID:        brand.ID.String(),
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
		_ = c.Error(security.SecureErrInvalidUUID(err))
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
				ID:        brand.ID.String(),
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
				ID:      brand.ID.String(),
				Name:    brand.Name,
				Slug:    brand.Slug,
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
		_ = c.Error(security.SecureErrInvalidUUID(err))
		return
	}

	err = bh.bservice.UpdateBrand(
		c.Request.Context(),
		brandID,
		UpdateBrandInput{
			Name:    req.Name,
			LogoURL: req.LogoURL,
			LogoAlt: req.LogoAlt,
		},
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK,
		api.MessageResponse{
			Message: "updated successfully",
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
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.ApplyDefaults(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	ctx := c.Request.Context()
	brandList, err := bh.bservice.List(
		ctx,
		q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*BrandResponse, 0, len(brandList.Brands))
	for _, brand := range brandList.Brands {
		res = append(res, &BrandResponse{
			ID:        brand.ID.String(),
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
			Meta: brandList.Page,
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
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.ApplyDefaults(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	}).Parse()

	ctx := c.Request.Context()
	brandList, err := bh.bservice.List(
		ctx,
		q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*BrandResponse, 0, len(brandList.Brands))
	for _, brand := range brandList.Brands {
		res = append(res, &BrandResponse{
			ID:        brand.ID.String(),
			Name:      brand.Name,
			LogoURL:   brand.LogoURL,
			LogoAlt:   brand.LogoAlt,
			Status:    brand.Status.String(),
			CreatedAt: brand.CreatedAt.Format(time.RFC3339),
			UpdatedAt: brand.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: brandList.Page,
		},
	)
}

// UnarchiveBrand godoc
// @Tags Brand
// @Accept json
// @Produce json
// @Param id path string true "brand id"
// @Failure 500 {object} api.ErrorResponse
// @Failure 200 {object} api.SuccessResponse
// @Router /admin/brands/{id}/unarchive [post]
func (bh *BrandHandler) UnarchiveBrand(c *gin.Context) {
	param := &BrandIDParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	brandID, err := uuid.Parse(param.ID)
	if err != nil {
		_ = c.Error(security.SecureErrInvalidUUID(err))
		return
	}

	ctx := c.Request.Context()
	if err := bh.bservice.Unarchive(ctx, brandID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "brand unarchived",
		},
	)
}

// ArchiveBrand godoc
// @Tags Brand
// @Accept json
// @Produce json
// @Param id path string true "brand id"
// @Failure 500 {object} api.ErrorResponse
// @Failure 200 {object} api.SuccessResponse
// @Router /admin/brands/{id}/archive [post]
func (bh *BrandHandler) ArchiveBrand(c *gin.Context) {
	param := &BrandIDParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	brandID, err := uuid.Parse(param.ID)
	if err != nil {
		_ = c.Error(security.SecureErrInvalidUUID(err))
		return
	}

	ctx := c.Request.Context()
	if err := bh.bservice.Archive(ctx, brandID); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "brand archived",
		},
	)
}
