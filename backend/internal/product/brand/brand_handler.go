package brand

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
)

type BrandHandler struct {
	bservice *BrandService
}

func NewHandler(sb *BrandService) *BrandHandler {
	return &BrandHandler{
		bservice: sb,
	}
}

type CreateBrandRequest struct {
	Name string  `json:"name" binding:"required" example:"apple"`
	Link *string `json:"link,omitempty" example:"https://apple.com"`
}

type UpdateBrandRequest struct {
	Name                *string    `json:"name,omitempty" example:"apple"`
	Link                *string    `json:"link,omitempty" example:"https://apple.com"`
	LogoStorageObjectID *uuid.UUID `json:"logo_storage_object_id,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
}

type BrandResponse struct {
	ID                  string     `json:"id" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	Name                string     `json:"name" example:"nvidia"`
	Link                *string    `json:"link,omitempty" example:"https://nvidia.com"`
	LogoStorageObjectID *uuid.UUID `json:"logo_storage_object_id,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	CreatedAt           string     `json:"created_at" example:"2026-07-01T05:04:38Z"`
	UpdatedAt           string     `json:"updated_at" example:"2026-07-01T05:04:38Z"`
}

type AdminBrandResponse struct {
	ID                  string     `json:"id" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	Name                string     `json:"name" example:"nvidia"`
	Link                *string    `json:"link,omitempty" example:"https://nvidia.com"`
	LogoStorageObjectID *uuid.UUID `json:"logo_storage_object_id,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	CreatedAt           string     `json:"created_at" example:"2026-07-01T05:04:38Z"`
	UpdatedAt           string     `json:"updated_at" example:"2026-07-01T05:04:38Z"`
	DeletedAt           *string    `json:"deleted_at,omitempty" example:"2026-07-01T05:04:38Z"`
}

// CreateBrand godoc
// @Summary create a new product brand
// @Description create a new product brand
// @Tags Brand
// @Accept json
// @Produce json
// @Param brand body CreateBrandRequest true "brand data"
// @Failure 400 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=BrandResponse}
// @Router /admin/brands [post]
func (bh *BrandHandler) CreateBrand(c *gin.Context) {
	req := &CreateBrandRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &CreateBrandInput{
		Name: req.Name,
		Link: req.Link,
	}

	brand, err := bh.bservice.CreateBrand(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusCreated,
		api.DataResponse{
			Data: BrandResponse{
				ID:                  brand.ID.String(),
				Name:                brand.Name,
				Link:                brand.Link,
				LogoStorageObjectID: brand.LogoStorageObjectID,
				CreatedAt:           brand.CreatedAt.Format(time.RFC3339),
				UpdatedAt:           brand.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

// GetBrandByID godoc
// @Summary get brand by id
// @Description get brand by id
// @Tags Brand
// @Accept json
// @Produce json
// @Param id path string true "Brand ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=BrandResponse}
// @Router /admin/brands/{id} [get]
func (bh *BrandHandler) GetBrandByID(c *gin.Context) {
	brandID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			apierr.ErrInvalidUUID().WithFields(
				api.FieldError{
					Field:   "brand_id",
					Message: "invalid brand UUID format",
				},
			),
		)
		return
	}

	brand, err := bh.bservice.GetBrandByID(c.Request.Context(), brandID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: BrandResponse{
				ID:                  brand.ID.String(),
				Name:                brand.Name,
				Link:                brand.Link,
				LogoStorageObjectID: brand.LogoStorageObjectID,
				CreatedAt:           brand.CreatedAt.Format(time.RFC3339),
				UpdatedAt:           brand.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

// UpdateBrand godoc
// @Summary update a brand
// @Description update specific brand fields by id
// @Tags Brand
// @Accept json
// @Produce json
// @Param id path string true "Brand ID (UUID)" format(uuid)
// @Param input body UpdateBrandRequest true "brand update payload"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/brands/{id} [patch]
func (bh *BrandHandler) UpdateBrand(c *gin.Context) {
	brandID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			apierr.ErrInvalidUUID().WithFields(
				api.FieldError{
					Field:   "brand_id",
					Message: "invalid brand UUID format",
				},
			),
		)
		return
	}

	body := &UpdateBrandRequest{}
	if err = c.ShouldBindJSON(body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	err = bh.bservice.UpdateBrand(
		c.Request.Context(),
		brandID,
		UpdateBrandInput{
			Name:                body.Name,
			Link:                body.Link,
			LogoStorageObjectID: body.LogoStorageObjectID,
		},
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "updated successfully",
	})
}

// DeleteBrandByID godoc
// @Summary soft-delete a brand
// @Description soft-delete an active brand by id
// @Tags Brand
// @Accept json
// @Produce json
// @Param id path string true "Brand ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/brands/{id} [delete]
func (bh *BrandHandler) DeleteBrandByID(c *gin.Context) {
	brandID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			apierr.ErrInvalidUUID().WithFields(
				api.FieldError{
					Field:   "brand_id",
					Message: "invalid brand UUID format",
				},
			),
		)
		return
	}

	if err := bh.bservice.DeleteBrandByID(c.Request.Context(), brandID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "deleted successfully",
	})
}

// ListBrands godoc
// @Summary list public active brands
// @Description list public active brands in pages
// @Tags Brand
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "page query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]BrandResponse,meta=api.Page}
// @Router /brands [get]
func (bh *BrandHandler) ListBrands(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := bh.bservice.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]*BrandResponse, 0, len(result.Items))
	for _, brand := range result.Items {
		res = append(res, &BrandResponse{
			ID:                  brand.ID.String(),
			Name:                brand.Name,
			Link:                brand.Link,
			LogoStorageObjectID: brand.LogoStorageObjectID,
			CreatedAt:           brand.CreatedAt.Format(time.RFC3339),
			UpdatedAt:           brand.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: result.Page,
		},
	)
}

// ListAdminBrands godoc
// @Summary list all brands including soft-deleted ones (admin)
// @Description list all brands in pages for administration
// @Tags Brand
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "page query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]AdminBrandResponse,meta=api.Page}
// @Router /admin/brands [get]
func (bh *BrandHandler) ListAdminBrands(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := bh.bservice.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]*AdminBrandResponse, 0, len(result.Items))
	for _, brand := range result.Items {
		item := &AdminBrandResponse{
			ID:                  brand.ID.String(),
			Name:                brand.Name,
			Link:                brand.Link,
			LogoStorageObjectID: brand.LogoStorageObjectID,
			CreatedAt:           brand.CreatedAt.Format(time.RFC3339),
			UpdatedAt:           brand.UpdatedAt.Format(time.RFC3339),
		}
		if brand.DeletedAt != nil {
			deletedStr := brand.DeletedAt.Format(time.RFC3339)
			item.DeletedAt = &deletedStr
		}
		res = append(res, item)
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: result.Page,
		},
	)
}
