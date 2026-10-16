package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// AddressHandler processes HTTP requests for customer shipping and billing addresses.
// It enforces authenticated session context, translates JSON payloads to domain inputs,
// coordinates address business logic via AddressService, and formats JSON responses.
type AddressHandler struct {
	service *AddressService
}

// NewAddressHandler initializes an AddressHandler instance with the address domain service dependency.
func NewAddressHandler(service *AddressService) *AddressHandler {
	return &AddressHandler{service: service}
}

// AddressResponse represents the public API representation of a user's physical postal address.
// Sensitive internal fields (such as soft-deletion timestamps) are excluded from this serialization.
type AddressResponse struct {
	ID         string  `json:"id"`
	UserID     string  `json:"userId"`
	Label      *string `json:"label,omitempty"`
	FullName   string  `json:"fullName"`
	Phone      *string `json:"phone,omitempty"`
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2,omitempty"`
	City       string  `json:"city"`
	State      *string `json:"state,omitempty"`
	PostalCode *string `json:"postalCode,omitempty"`
	Country    string  `json:"country"`
	IsDefault  bool    `json:"isDefault"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

// mapAddressResponse transforms an internal UserAddress domain model entity into the public AddressResponse DTO,
// formatting timestamps as ISO 8601 / RFC3339 strings and stringifying UUID identifiers.
func mapAddressResponse(a *model.UserAddress) AddressResponse {
	return AddressResponse{
		ID:         a.ID.String(),
		UserID:     a.UserID.String(),
		Label:      a.Label,
		FullName:   a.FullName,
		Phone:      a.Phone,
		Line1:      a.Line1,
		Line2:      a.Line2,
		City:       a.City,
		State:      a.State,
		PostalCode: a.PostalCode,
		Country:    a.Country,
		IsDefault:  a.IsDefault,
		CreatedAt:  a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  a.UpdatedAt.Format(time.RFC3339),
	}
}

// ListAddresses godoc
//
//	@Summary		List saved user addresses
//	@Description	Returns all active saved shipping and billing addresses for the authenticated customer.
//	@Tags			Addresses
//	@Produce		json
//	@Success		200	{object}	api.DataResponse{data=[]AddressResponse}
//	@Failure		401	{object}	api.UnauthorizedErrorResponse
//	@Router			/user/addresses [get]
func (h *AddressHandler) ListAddresses(c *gin.Context) {
	// Extract authenticated user ID from session context to ensure users can only list their own addresses
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Fetch active non-deleted addresses ordered with default first
	addresses, err := h.service.ListAddresses(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Transform domain entities to API response format
	resp := make([]AddressResponse, len(addresses))
	for i := range addresses {
		resp[i] = mapAddressResponse(&addresses[i])
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: resp})
}

// GetAddress godoc
//
//	@Summary		Get address details
//	@Tags			Addresses
//	@Produce		json
//	@Param			id	path		string	true	"Address UUID"
//	@Success		200	{object}	api.DataResponse{data=AddressResponse}
//	@Failure		401	{object}	api.UnauthorizedErrorResponse
//	@Failure		404	{object}	api.NotFoundErrorResponse
//	@Router			/user/addresses/{id} [get]
func (h *AddressHandler) GetAddress(c *gin.Context) {
	// Extract caller's user UUID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Parse target address UUID from path parameter
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid address id").WithCode(apierr.CodeInvalidInput))
		return
	}

	// Retrieve address strictly scoped to the authenticated user ID (preventing IDOR)
	address, err := h.service.GetAddress(c.Request.Context(), id, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapAddressResponse(address)})
}

// CreateAddress godoc
//
//	@Summary		Create new address
//	@Tags			Addresses
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateAddressInput	true	"Address details"
//	@Success		201		{object}	api.DataResponse{data=AddressResponse}
//	@Failure		400		{object}	api.BadRequestErrorResponse
//	@Failure		401		{object}	api.UnauthorizedErrorResponse
//	@Router			/user/addresses [post]
func (h *AddressHandler) CreateAddress(c *gin.Context) {
	// Extract authenticated user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Parse JSON input payload
	var req CreateAddressInput
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("invalid request payload").WithCode(apierr.CodeInvalidInput))
		return
	}

	// Delegate address validation, creation, and default address promotion to service
	address, err := h.service.CreateAddress(c.Request.Context(), userID, &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: mapAddressResponse(address)})
}

// UpdateAddress godoc
//
//	@Summary		Update address
//	@Tags			Addresses
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Address UUID"
//	@Param			request	body		UpdateAddressInput	true	"Address details"
//	@Success		200		{object}	api.DataResponse{data=AddressResponse}
//	@Failure		400		{object}	api.BadRequestErrorResponse
//	@Failure		401		{object}	api.UnauthorizedErrorResponse
//	@Failure		404		{object}	api.NotFoundErrorResponse
//	@Router			/user/addresses/{id} [put]
func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	// Extract authenticated user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Parse address UUID path parameter
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid address id").WithCode(apierr.CodeInvalidInput))
		return
	}

	// Parse JSON update input payload
	var req UpdateAddressInput
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("invalid request payload").WithCode(apierr.CodeInvalidInput))
		return
	}

	// Update address details with ownership validation in database transaction
	address, err := h.service.UpdateAddress(c.Request.Context(), id, userID, &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapAddressResponse(address)})
}

// SetDefaultAddress godoc
//
//	@Summary		Set address as default
//	@Tags			Addresses
//	@Produce		json
//	@Param			id	path		string	true	"Address UUID"
//	@Success		200	{object}	api.DataResponse{data=AddressResponse}
//	@Failure		401	{object}	api.UnauthorizedErrorResponse
//	@Failure		404	{object}	api.NotFoundErrorResponse
//	@Router			/user/addresses/{id}/default [patch]
func (h *AddressHandler) SetDefaultAddress(c *gin.Context) {
	// Extract authenticated user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Parse address UUID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid address id").WithCode(apierr.CodeInvalidInput))
		return
	}

	// Atomically unset previous default and mark this address as default
	address, err := h.service.SetDefaultAddress(c.Request.Context(), id, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapAddressResponse(address)})
}

// DeleteAddress godoc
//
//	@Summary		Delete address
//	@Tags			Addresses
//	@Produce		json
//	@Param			id	path		string	true	"Address UUID"
//	@Success		200	{object}	api.MessageResponse
//	@Failure		401	{object}	api.UnauthorizedErrorResponse
//	@Failure		404	{object}	api.NotFoundErrorResponse
//	@Router			/user/addresses/{id} [delete]
func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	// Extract caller's user UUID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Parse address UUID from URL path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid address id").WithCode(apierr.CodeInvalidInput))
		return
	}

	// Soft-delete address and promote the next most recent active address if default was deleted
	if err := h.service.DeleteAddress(c.Request.Context(), id, userID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Address deleted successfully"})
}
