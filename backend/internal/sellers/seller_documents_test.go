package sellers_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/sellers"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
)

func TestSellerDocumentRoutesAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := sellers.NewRouter(sellers.NewHandler(new(sellers.Service), stubMediaUploader{}), staticSessionStore{})
	engine := gin.New()
	router.MapRoutes(engine.Group("/api/v1"))

	want := map[string]bool{
		"GET /api/v1/sellers/me/documents":                         true,
		"POST /api/v1/sellers/me/documents":                        true,
		"DELETE /api/v1/sellers/me/documents/:id":                  true,
		"GET /api/v1/admin/sellers/:id/documents":                  true,
		"PATCH /api/v1/admin/sellers/:id/documents/:doc_id/review": true,
	}
	for _, route := range engine.Routes() {
		delete(want, route.Method+" "+route.Path)
	}
	for route := range want {
		t.Errorf("missing route: %s", route)
	}
}

// seedDocumentObject inserts the storage object a KYC document binds to and
// returns its id. The Noop provider gives every object a deterministic URL.
func (s *SellerServiceSuite) seedDocumentObject() uuid.UUID {
	objectID := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO storage_objects (id, bucket, object_key, content_type, file_size, status)
		VALUES ($1, 'seller-documents', $2, 'application/pdf', 2048, 'uploaded')
	`, objectID, uuid.New().String())
	s.Require().NoError(err)
	return objectID
}

func (s *SellerServiceSuite) clearDocuments() {
	_, err := s.db.ExecContext(s.ctx, `DELETE FROM seller_documents WHERE seller_id = $1`, s.sellerID)
	s.Require().NoError(err)
}

func (s *SellerServiceSuite) createDocument(docType model.SellerDocumentType, name string) *model.SellerDocument {
	doc, err := s.service.CreateDocument(s.ctx, s.sellerUID, sellers.CreateDocumentInput{
		DocumentType: docType,
		ObjectID:     s.seedDocumentObject(),
		FileName:     name,
	})
	s.Require().NoError(err)
	return doc
}

func (s *SellerServiceSuite) TestCreateDocumentBindsObjectAndListsWithPublicURL() {
	s.clearDocuments()

	doc := s.createDocument(model.SellerDocumentTypeTaxCard, "tax-card.pdf")

	s.NotEqual(uuid.Nil, doc.ID)
	s.Equal(model.SellerDocumentTypeTaxCard, doc.DocumentType)
	s.Equal(model.SellerDocumentStatusPending, doc.Status)
	s.Equal("tax-card.pdf", doc.FileName)
	s.Equal(int64(2048), doc.FileSize)
	s.Equal("application/pdf", doc.ContentType)
	s.NotEmpty(doc.PublicURL)

	listed, err := s.service.ListMyDocuments(s.ctx, s.sellerUID)
	s.Require().NoError(err)
	s.Require().Len(listed, 1)
	s.Equal(doc.ID, listed[0].ID)
	s.Equal(doc.PublicURL, listed[0].PublicURL)
}

func (s *SellerServiceSuite) TestCreateDocumentRejectsDuplicateTypeForSameSeller() {
	s.clearDocuments()
	s.createDocument(model.SellerDocumentTypeTaxCard, "first.pdf")

	_, err := s.service.CreateDocument(s.ctx, s.sellerUID, sellers.CreateDocumentInput{
		DocumentType: model.SellerDocumentTypeTaxCard,
		ObjectID:     s.seedDocumentObject(),
		FileName:     "second.pdf",
	})

	apiErr := s.assertAPIError(err)
	s.Equal(409, apiErr.Status)
	s.Equal(errcode.CodeSellerDocumentAlreadyExists, apiErr.Code)
}

func (s *SellerServiceSuite) TestAdminReviewApprovesAndRejectsDocuments() {
	s.clearDocuments()
	approved := s.createDocument(model.SellerDocumentTypeCommercialRegister, "register.pdf")
	rejected := s.createDocument(model.SellerDocumentTypeNationalID, "national-id.pdf")

	reviewedApproval, err := s.service.AdminReviewDocument(s.ctx, s.sellerID, approved.ID, s.sellerUID, model.SellerDocumentStatusApproved, "")
	s.Require().NoError(err)
	s.Equal(model.SellerDocumentStatusApproved, reviewedApproval.Status)
	s.NotNil(reviewedApproval.ReviewedAt)
	s.Require().NotNil(reviewedApproval.ReviewedBy)
	s.Equal(s.sellerUID, *reviewedApproval.ReviewedBy)
	s.Empty(reviewedApproval.RejectionReason)

	reviewedRejection, err := s.service.AdminReviewDocument(s.ctx, s.sellerID, rejected.ID, s.sellerUID, model.SellerDocumentStatusRejected, "Blurry scan")
	s.Require().NoError(err)
	s.Equal(model.SellerDocumentStatusRejected, reviewedRejection.Status)
	s.Require().NotNil(reviewedRejection.RejectionReason)
	s.Equal("Blurry scan", *reviewedRejection.RejectionReason)

	listed, err := s.service.AdminListSellerDocuments(s.ctx, s.sellerID)
	s.Require().NoError(err)
	s.Require().Len(listed, 2)
	statuses := map[uuid.UUID]model.SellerDocumentStatus{}
	for _, doc := range listed {
		statuses[doc.ID] = doc.Status
	}
	s.Equal(model.SellerDocumentStatusApproved, statuses[approved.ID])
	s.Equal(model.SellerDocumentStatusRejected, statuses[rejected.ID])
}

func (s *SellerServiceSuite) TestDeleteMyDocumentOnlyAllowsPendingOrRejected() {
	s.clearDocuments()
	pending := s.createDocument(model.SellerDocumentTypeOther, "pending.pdf")
	s.Require().NoError(s.service.DeleteMyDocument(s.ctx, s.sellerUID, pending.ID))

	listed, err := s.service.ListMyDocuments(s.ctx, s.sellerUID)
	s.Require().NoError(err)
	s.Empty(listed)

	approved := s.createDocument(model.SellerDocumentTypeOther, "approved.pdf")
	_, err = s.service.AdminReviewDocument(s.ctx, s.sellerID, approved.ID, s.sellerUID, model.SellerDocumentStatusApproved, "")
	s.Require().NoError(err)

	err = s.service.DeleteMyDocument(s.ctx, s.sellerUID, approved.ID)
	apiErr := s.assertAPIError(err)
	s.Equal(409, apiErr.Status)
	s.Equal(errcode.CodeSellerDocumentConflict, apiErr.Code)
}

func (s *SellerServiceSuite) TestUpdateMyProfilePersistsKYCLegalFields() {
	legalName, taxNumber, commercialRegister := "PRIM Trading LLC", "123456789", "CR-1234"
	updated, err := s.service.UpdateMyProfile(s.ctx, s.sellerUID, sellers.UpdateSellerInput{
		LegalName:          &legalName,
		TaxNumber:          &taxNumber,
		CommercialRegister: &commercialRegister,
	})
	s.Require().NoError(err)
	s.Require().NotNil(updated.LegalName)
	s.Require().NotNil(updated.TaxNumber)
	s.Require().NotNil(updated.CommercialRegister)
	s.Equal(legalName, *updated.LegalName)
	s.Equal(taxNumber, *updated.TaxNumber)
	s.Equal(commercialRegister, *updated.CommercialRegister)

	reloaded, err := s.service.GetSellerByUser(s.ctx, s.sellerUID)
	s.Require().NoError(err)
	s.Require().NotNil(reloaded.LegalName)
	s.Require().NotNil(reloaded.TaxNumber)
	s.Require().NotNil(reloaded.CommercialRegister)
	s.Equal(legalName, *reloaded.LegalName)
	s.Equal(taxNumber, *reloaded.TaxNumber)
	s.Equal(commercialRegister, *reloaded.CommercialRegister)
}
