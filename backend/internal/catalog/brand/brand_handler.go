package brand

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
)

type BrandHandler struct {
	brandService *BrandService
}

func NewHandler(brandService *BrandService) *BrandHandler {
	return &BrandHandler{
		brandService: brandService,
	}
}

type CreateBrandRequest struct {
	Name string  `json:"name" binding:"required" example:"apple"`
	Link *string `json:"link,omitempty" example:"https://apple.com"`
}

type UpdateBrandRequest struct {
	Name         *string    `json:"name,omitempty" example:"apple"`
	Link         *string    `json:"link,omitempty" example:"https://apple.com"`
	LogoObjectID *uuid.UUID `json:"logo_object_id,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
}

type BrandResponse struct {
	ID      string  `json:"id" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"` // Public ID of the brand
	Name    string  `json:"name" example:"nvidia"`
	Link    *string `json:"link,omitempty" example:"https://nvidia.com"`
	LogoURL *string `json:"logo_url,omitempty" example:"https://example.com/logo.png"`
}

type AdminBrandResponse struct {
	ID           string     `json:"id" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`        // Internal ID of the brand
	PublicID     string     `json:"public_id" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"` // Public ID of the brand
	Name         string     `json:"name" example:"nvidia"`
	Link         *string    `json:"link,omitempty" example:"https://nvidia.com"`
	LogoObjectID *uuid.UUID `json:"logo_object_id,omitempty" example:"358b2e03-0b3f-40a4-8163-ebed0cb252ee"`
	CreatedAt    string     `json:"created_at" example:"2026-07-01T05:04:38Z"`
	UpdatedAt    string     `json:"updated_at" example:"2026-07-01T05:04:38Z"`
	DeletedAt    *string    `json:"deleted_at,omitempty" example:"2026-07-01T05:04:38Z"`
}

// CreateBrand godoc
//
//	@Summary		Create a product brand
//	@Description	Adds a new product manufacturer/brand to the store catalog. Brand names must be unique.
//	@Tags			Brands
//	@Accept			json
//	@Produce		json
//	@Param			brand	body		CreateBrandRequest						true	"Brand details (name and optional external website link)"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Validation error or missing brand name"
//	@Failure		409		{object}	api.ConflictErrorResponse				"A brand with this name already exists"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		201		{object}	api.DataResponse{data=AdminBrandResponse}	"Created brand details"
//	@Router			/admin/brands [post]
func (bh *BrandHandler) CreateBrand(c *gin.Context) {
	req := &CreateBrandRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		_ = c.Error(apierr.ErrBadRequest("Brand name is required").WithCode(apierr.CodeValidationFailed))
		return
	}

	in := &CreateBrandInput{
		Name: name,
		Link: req.Link,
	}

	brand, err := bh.brandService.CreateBrand(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusCreated,
		api.DataResponse{
			Data: mapAdminBrandResponse(brand),
		},
	)
}

// GetBrandByID godoc
//
//	@Summary		Get brand details by ID
//	@Description	Retrieves active product brand details by its UUID.
//	@Tags			Brands
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string									true	"Brand UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse				"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse				"Brand not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=AdminBrandResponse}	"Brand details"
//	@Router			/admin/brands/{id} [get]
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

	brand, err := bh.brandService.GetBrandByID(c.Request.Context(), brandID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: mapAdminBrandResponse(brand),
		},
	)
}

// UpdateBrand godoc
//
//	@Summary		Update brand details
//	@Description	Updates specific fields of an existing brand such as name, website link, or logo object reference.
//	@Tags			Brands
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Brand UUID"	format(uuid)
//	@Param			input	body		UpdateBrandRequest				true	"Fields to update (name, link, logo_id)"
//	@Failure		400		{object}	api.BadRequestErrorResponse		"Validation error or referenced logo object not found"
//	@Failure		404		{object}	api.NotFoundErrorResponse		"Brand not found"
//	@Failure		409		{object}	api.ConflictErrorResponse		"A brand with updated name already exists"
//	@Failure		500		{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200		{object}	api.MessageResponse				"Update confirmation message"
//	@Router			/admin/brands/{id} [patch]
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

	err = bh.brandService.UpdateBrand(
		c.Request.Context(),
		brandID,
		UpdateBrandInput{
			Name: body.Name,
			Link: body.Link,
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
//
//	@Summary		Soft-delete a brand
//	@Description	Marks an active brand as soft-deleted (`deleted_at = NOW()`), removing it from active listings.
//	@Tags			Brands
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string							true	"Brand UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Brand not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.MessageResponse				"Deletion confirmation message"
//	@Router			/admin/brands/{id} [delete]
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

	if err := bh.brandService.DeleteBrandByID(c.Request.Context(), brandID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListBrands godoc
//
//	@Summary		List active product brands
//	@Description	Returns a paginated list of active product brands for customer storefront browsing.
//	@Tags			Brands
//	@Accept			json
//	@Produce		json
//	@Param			q	query		pagination.ListQuery												true	"Pagination, search query, and sorting parameters"
//	@Failure		400	{object}	api.BadRequestErrorResponse											"Invalid query parameters"
//	@Failure		500	{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Success		200	{object}	api.PaginatedResponse{data=[]BrandResponse,meta=pagination.Page}	"Paginated list of active brands"
//	@Router			/brands [get]
func (bh *BrandHandler) ListBrands(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(pagination.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := bh.brandService.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]*BrandResponse, 0, len(result.Items))
	for _, brand := range result.Items {
		res = append(res, mapBrandResponse(brand))
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
//
//	@Summary		List all brands including soft-deleted ones (Admin)
//	@Description	Returns a paginated list of all product brands including soft-deleted records for administrator management.
//	@Tags			Brands
//	@Accept			json
//	@Produce		json
//	@Param			q	query		pagination.ListQuery													true	"Pagination, search query, and sorting parameters"
//	@Failure		400	{object}	api.BadRequestErrorResponse												"Invalid query parameters"
//	@Failure		500	{object}	api.InternalServerErrorResponse											"Internal server error"
//	@Success		200	{object}	api.PaginatedResponse{data=[]AdminBrandResponse,meta=pagination.Page}	"Paginated list of all brands including deleted"
//	@Router			/admin/brands [get]
func (bh *BrandHandler) ListAdminBrands(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(pagination.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := bh.brandService.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]*AdminBrandResponse, 0, len(result.Items))
	for _, brand := range result.Items {
		res = append(res, mapAdminBrandResponse(brand))
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: result.Page,
		},
	)
}

// UploadBrandLogo godoc
//
//	@Summary		Upload brand logo
//	@Description	Uploads an image file to be used as the brand's logo
//	@Tags			Brands
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id		path		string						true	"Brand UUID"	format(uuid)
//	@Param			file	formData	file						true	"Logo image file"
//	@Failure		400		{object}	api.BadRequestErrorResponse	"Invalid UUID format or missing file"
//	@Failure		500		{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200		{object}	api.MessageResponse			"Upload confirmation message"
//	@Router			/admin/brands/{id}/logo [post]
func (bh *BrandHandler) UploadBrandLogo(c *gin.Context) {
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

	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("file is required").WithCode(apierr.CodeInvalidInput))
		return
	}

	_, err = bh.brandService.UploadBrandLogo(c.Request.Context(), brandID, file)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "logo uploaded successfully",
	})
}

func mapBrandResponse(brand *model.ProductBrand) *BrandResponse {
	return &BrandResponse{
		ID:      brand.PublicID,
		Name:    brand.Name,
		Link:    brand.Link,
		LogoURL: brand.LogoURL,
	}
}

func mapAdminBrandResponse(brand *model.ProductBrand) *AdminBrandResponse {
	res := &AdminBrandResponse{
		ID:           brand.ID.String(),
		PublicID:     brand.PublicID,
		Name:         brand.Name,
		Link:         brand.Link,
		LogoObjectID: brand.LogoObjectID,
		CreatedAt:    brand.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    brand.UpdatedAt.Format(time.RFC3339),
	}
	if brand.DeletedAt != nil {
		deletedStr := brand.DeletedAt.Format(time.RFC3339)
		res.DeletedAt = &deletedStr
	}
	return res
}
