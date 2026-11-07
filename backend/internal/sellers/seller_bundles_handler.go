package sellers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// SellerBundleItemRequest is one bundle component.
type SellerBundleItemRequest struct {
	// Component product UUID.
	ProductID string `json:"productId" binding:"required,uuid"`
	// Optional component variant UUID; defaults to the product's default line.
	VariantID *string `json:"variantId,omitempty" binding:"omitempty,uuid"`
	// Units in the kit.
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// SellerBundleRequest is the bundle creation payload.
type SellerBundleRequest struct {
	// Bundle display title.
	Title string `json:"title" binding:"required,max=200"`
	// Optional description.
	Description *string `json:"description,omitempty"`
	// Bundle price in minor units, charged per kit.
	BundlePrice int64 `json:"bundlePrice" binding:"required,gt=0"`
	// ISO currency code; defaults to EGP.
	Currency *string `json:"currency,omitempty"`
	// Optional public cover image URL.
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	// Kit composition; at least one item.
	Items []SellerBundleItemRequest `json:"items" binding:"required,min=1"`
}

// ListSellerBundles godoc
// @Summary      List my bundles
// @Description  Returns product bundles owned by the authenticated seller with their composition.
// @Tags         Seller
// @Produce      json
// @Success      200 {object} api.DataResponse{data=[]SellerBundle}
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Router       /seller/bundles [get]
// @Security     SessionAuth
func (h *Handler) ListSellerBundles(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	bundles, err := h.service.ListBundles(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: bundles})
}

// CreateSellerBundle godoc
// @Summary      Create a bundle
// @Description  Bundles existing catalog items into a kit product priced at the bundle price.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        request body SellerBundleRequest true "Bundle payload"
// @Success      201 {object} api.DataResponse{data=SellerBundle}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Router       /seller/bundles [post]
// @Security     SessionAuth
func (h *Handler) CreateSellerBundle(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req SellerBundleRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	items := make([]BundleItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			_ = c.Error(apierr.InvalidUUIDError().WithFields(api.FieldError{
				Field: "items", Constraint: "uuid", Message: "productId must be a UUID",
			}))
			return
		}
		var variantID *uuid.UUID
		if item.VariantID != nil {
			parsed, err := uuid.Parse(*item.VariantID)
			if err != nil {
				_ = c.Error(apierr.InvalidUUIDError().WithFields(api.FieldError{
					Field: "items", Constraint: "uuid", Message: "variantId must be a UUID",
				}))
				return
			}
			variantID = &parsed
		}
		items = append(items, BundleItemInput{ProductID: productID, VariantID: variantID, Quantity: item.Quantity})
	}

	created, err := h.service.CreateBundle(c.Request.Context(), userID, CreateBundleInput{
		Title:        req.Title,
		Description:  req.Description,
		BundlePrice:  req.BundlePrice,
		Currency:     derefString(req.Currency),
		ThumbnailURL: req.ThumbnailURL,
		Items:        items,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: created})
}

// DeleteSellerBundle godoc
// @Summary      Delete a bundle
// @Description  Removes the bundle composition and its backing product.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Bundle UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Bundle not found"
// @Router       /seller/bundles/{id} [delete]
// @Security     SessionAuth
func (h *Handler) DeleteSellerBundle(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	bundleID, ok := bindSellerID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteBundle(c.Request.Context(), userID, bundleID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Bundle deleted"})
}
