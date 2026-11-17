package sellers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// CreateSellerDocumentRequest binds an uploaded storage object to a KYC type.
type CreateSellerDocumentRequest struct {
	// One of commercial_register, tax_card, national_id, other.
	DocumentType model.SellerDocumentType `json:"documentType" binding:"required,oneof=commercial_register tax_card national_id other" enums:"commercial_register,tax_card,national_id,other"`
	// ID returned by the existing storage-object upload flow.
	ObjectID string `json:"objectId" binding:"required,uuid" example:"10000000-0000-0000-0000-00000000000f"`
	// Original file name shown to reviewers.
	FileName string `json:"fileName" binding:"required,min=1,max=255" example:"commercial-register.pdf"`
}

// SellerDocumentResponse is the merchant/admin view of a verification document.
type SellerDocumentResponse struct {
	ID              uuid.UUID                  `json:"id"`
	DocumentType    model.SellerDocumentType   `json:"documentType" enums:"commercial_register,tax_card,national_id,other"`
	FileName        string                     `json:"fileName"`
	FileSize        int64                      `json:"fileSize"`
	ContentType     string                     `json:"contentType"`
	URL             string                     `json:"url,omitempty"`
	Status          model.SellerDocumentStatus `json:"status" enums:"pending,approved,rejected"`
	RejectionReason string                     `json:"rejectionReason,omitempty"`
	ReviewedBy      *uuid.UUID                 `json:"reviewedBy,omitempty"`
	ReviewedAt      string                     `json:"reviewedAt,omitempty"`
	CreatedAt       string                     `json:"createdAt"`
	UpdatedAt       string                     `json:"updatedAt"`
}

func toSellerDocumentResponse(doc *model.SellerDocument) SellerDocumentResponse {
	res := SellerDocumentResponse{
		ID:              doc.ID,
		DocumentType:    doc.DocumentType,
		FileName:        doc.FileName,
		FileSize:        doc.FileSize,
		ContentType:     doc.ContentType,
		URL:             doc.PublicURL,
		Status:          doc.Status,
		ReviewedBy:      doc.ReviewedBy,
		RejectionReason: derefString(doc.RejectionReason),
		CreatedAt:       doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       doc.UpdatedAt.Format(time.RFC3339),
	}
	if doc.ReviewedAt != nil {
		res.ReviewedAt = doc.ReviewedAt.Format(time.RFC3339)
	}
	return res
}

func toSellerDocumentResponses(docs []*model.SellerDocument) []SellerDocumentResponse {
	out := make([]SellerDocumentResponse, 0, len(docs))
	for _, doc := range docs {
		out = append(out, toSellerDocumentResponse(doc))
	}
	return out
}

func bindDocumentID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("doc_id"))
	if err != nil {
		_ = c.Error(apierr.InvalidUUIDError().WithFields(
			api.FieldError{Field: "doc_id", Constraint: "uuid", Message: "invalid document id"},
		))
		return uuid.Nil, false
	}
	return id, true
}

// ListMyDocuments godoc
// @Summary      List my verification documents
// @Description  Returns the authenticated seller's KYC documents with public storage URLs.
// @Tags         Seller
// @Produce      json
// @Success      200 {object} api.DataResponse{data=[]sellers.SellerDocumentResponse}
// @Failure      401 {object} api.ErrorResponse "Authentication required"
// @Router       /sellers/me/documents [get]
// @Security     SessionAuth
func (h *Handler) ListMyDocuments(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}
	docs, err := h.service.ListMyDocuments(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerDocumentResponses(docs)})
}

// CreateDocument godoc
// @Summary      Submit a verification document
// @Description  Binds an existing storage object to one KYC document type for the authenticated seller. One document per type is allowed.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        request body CreateSellerDocumentRequest true "Document payload"
// @Success      201 {object} api.DataResponse{data=sellers.SellerDocumentResponse}
// @Failure      400 {object} api.ErrorResponse "Validation failed"
// @Failure      409 {object} api.ErrorResponse "Document type already submitted"
// @Router       /sellers/me/documents [post]
// @Security     SessionAuth
func (h *Handler) CreateDocument(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}
	var req CreateSellerDocumentRequest
	if !bindSellerJSON(c, &req) {
		return
	}
	objectID, err := uuid.Parse(req.ObjectID)
	if err != nil {
		_ = c.Error(apierr.InvalidUUIDError().WithFields(
			api.FieldError{Field: "objectId", Constraint: "uuid", Message: "invalid storage object id"},
		))
		return
	}
	doc, err := h.service.CreateDocument(c.Request.Context(), userID, CreateDocumentInput{
		DocumentType: req.DocumentType,
		ObjectID:     objectID,
		FileName:     req.FileName,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, api.DataResponse{Data: toSellerDocumentResponse(doc)})
}

// DeleteMyDocument godoc
// @Summary      Delete my verification document
// @Description  Removes a pending or rejected document. Approved evidence is retained for compliance.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Document UUID" format(uuid)
// @Success      204 "Deleted"
// @Failure      403 {object} api.ErrorResponse "Seller not authorized"
// @Failure      404 {object} api.ErrorResponse "Document not found"
// @Failure      409 {object} api.ErrorResponse "Document already approved"
// @Router       /sellers/me/documents/{id} [delete]
// @Security     SessionAuth
func (h *Handler) DeleteMyDocument(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}
	documentID, ok := bindSellerID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteMyDocument(c.Request.Context(), userID, documentID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AdminListSellerDocuments godoc
// @Summary      List a seller's verification documents (admin)
// @Tags         Admin Sellers
// @Produce      json
// @Param        id path string true "Seller UUID" format(uuid)
// @Success      200 {object} api.DataResponse{data=[]sellers.SellerDocumentResponse}
// @Failure      404 {object} api.ErrorResponse "Seller not found"
// @Router       /admin/sellers/{id}/documents [get]
// @Security     SessionAuth
func (h *Handler) AdminListSellerDocuments(c *gin.Context) {
	sellerID, ok := bindSellerID(c)
	if !ok {
		return
	}
	docs, err := h.service.AdminListSellerDocuments(c.Request.Context(), sellerID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerDocumentResponses(docs)})
}

// AdminReviewDocumentRequest is the admin KYC decision payload.
type AdminReviewDocumentRequest struct {
	// New review status.
	Status model.SellerDocumentStatus `json:"status" binding:"required,oneof=approved rejected" enums:"approved,rejected"`
	// Optional explanation for a rejection.
	RejectionReason string `json:"rejectionReason" binding:"max=1000" example:"Document is illegible"`
}

// AdminReviewDocument godoc
// @Summary      Review a seller verification document (admin)
// @Tags         Admin Sellers
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Seller UUID" format(uuid)
// @Param        doc_id  path string                     true "Document UUID" format(uuid)
// @Param        request body AdminReviewDocumentRequest true "Review decision"
// @Success      200 {object} api.DataResponse{data=sellers.SellerDocumentResponse}
// @Failure      400 {object} api.ErrorResponse "Validation failed"
// @Failure      404 {object} api.ErrorResponse "Document not found"
// @Router       /admin/sellers/{id}/documents/{doc_id}/review [patch]
// @Security     SessionAuth
func (h *Handler) AdminReviewDocument(c *gin.Context) {
	sellerID, ok := bindSellerID(c)
	if !ok {
		return
	}
	documentID, ok := bindDocumentID(c)
	if !ok {
		return
	}
	var req AdminReviewDocumentRequest
	if !bindSellerJSON(c, &req) {
		return
	}
	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}
	doc, err := h.service.AdminReviewDocument(c.Request.Context(), sellerID, documentID, reviewerID, req.Status, req.RejectionReason)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerDocumentResponse(doc)})
}
