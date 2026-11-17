package sellers_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/sellers"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/log"
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
	return s.seedDocumentObjectWith("application/pdf", 2048)
}

func (s *SellerServiceSuite) seedDocumentObjectWith(contentType string, size int64) uuid.UUID {
	objectID := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO storage_objects (id, bucket, object_key, content_type, file_size, status)
		VALUES ($1, 'seller-documents', $2, $3, $4, 'uploaded')
	`, objectID, uuid.New().String(), contentType, size)
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

func (s *SellerServiceSuite) TestCreateDocumentReplacesPendingAndRejected() {
	s.clearDocuments()
	first := s.createDocument(model.SellerDocumentTypeTaxCard, "first.pdf")
	firstObjectID := first.StorageObjectID

	replacement, err := s.service.CreateDocument(s.ctx, s.sellerUID, sellers.CreateDocumentInput{
		DocumentType: model.SellerDocumentTypeTaxCard,
		ObjectID:     s.seedDocumentObject(),
		FileName:     "second.pdf",
	})
	s.Require().NoError(err)
	s.Equal(first.ID, replacement.ID, "replacement must reuse the document row")
	s.NotEqual(firstObjectID, replacement.StorageObjectID, "replacement must bind the new object")
	s.Equal("second.pdf", replacement.FileName)
	s.Equal(model.SellerDocumentStatusPending, replacement.Status)

	_, err = s.service.AdminReviewDocument(s.ctx, s.sellerID, replacement.ID, s.sellerUID, model.SellerDocumentStatusRejected, "illegible")
	s.Require().NoError(err)

	secondReplacement, err := s.service.CreateDocument(s.ctx, s.sellerUID, sellers.CreateDocumentInput{
		DocumentType: model.SellerDocumentTypeTaxCard,
		ObjectID:     s.seedDocumentObject(),
		FileName:     "third.pdf",
	})
	s.Require().NoError(err)
	s.Equal(replacement.ID, secondReplacement.ID)
	s.Equal(model.SellerDocumentStatusPending, secondReplacement.Status)
	s.Nil(secondReplacement.RejectionReason, "rejection reason must be cleared on replacement")
	s.Nil(secondReplacement.ReviewedBy, "review metadata must be cleared on replacement")
	s.Nil(secondReplacement.ReviewedAt, "review metadata must be cleared on replacement")

	listed, err := s.service.ListMyDocuments(s.ctx, s.sellerUID)
	s.Require().NoError(err)
	s.Require().Len(listed, 1, "replacement must not create a second document")
}

func (s *SellerServiceSuite) TestCreateDocumentRefusesToReplaceApproved() {
	s.clearDocuments()
	approved := s.createDocument(model.SellerDocumentTypeCommercialRegister, "register.pdf")
	_, err := s.service.AdminReviewDocument(s.ctx, s.sellerID, approved.ID, s.sellerUID, model.SellerDocumentStatusApproved, "")
	s.Require().NoError(err)

	_, err = s.service.CreateDocument(s.ctx, s.sellerUID, sellers.CreateDocumentInput{
		DocumentType: model.SellerDocumentTypeCommercialRegister,
		ObjectID:     s.seedDocumentObject(),
		FileName:     "register-v2.pdf",
	})

	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusConflict, apiErr.Status)
	s.Equal(errcode.CodeSellerDocumentConflict, apiErr.Code)
	s.Contains(apiErr.Message, "Approved documents are locked for compliance")
}

func (s *SellerServiceSuite) TestCreateDocumentRejectsDisallowedBoundObject() {
	s.clearDocuments()
	_, err := s.service.CreateDocument(s.ctx, s.sellerUID, sellers.CreateDocumentInput{
		DocumentType: model.SellerDocumentTypeTaxCard,
		ObjectID:     s.seedDocumentObjectWith("image/gif", 2048),
		FileName:     "photo.gif",
	})
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("mime", apiErr.Fields[0].Constraint)

	_, err = s.service.CreateDocument(s.ctx, s.sellerUID, sellers.CreateDocumentInput{
		DocumentType: model.SellerDocumentTypeNationalID,
		ObjectID:     s.seedDocumentObjectWith("application/pdf", 0),
		FileName:     "empty.pdf",
	})
	apiErr = s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("size", apiErr.Fields[0].Constraint)
}

func (s *SellerServiceSuite) TestAdminReviewRequiresRejectionReason() {
	s.clearDocuments()
	doc := s.createDocument(model.SellerDocumentTypeNationalID, "national-id.pdf")

	_, err := s.service.AdminReviewDocument(s.ctx, s.sellerID, doc.ID, s.sellerUID, model.SellerDocumentStatusRejected, "   ")

	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("rejectionReason", apiErr.Fields[0].Field)
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

func (s *SellerServiceSuite) documentEngine() (*gin.Engine, *http.Cookie) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.ErrorHandler(log.NewNoopLogger()))
	sellers.NewRouter(sellers.NewHandler(s.service, stubMediaUploader{}), staticSessionStore{
		session: &model.UserSession{
			ID:       "doc-session",
			UserID:   s.sellerUID,
			UserRole: model.CustomerRole,
			Type:     model.SessionTypeAuthenticated,
		},
	}).MapRoutes(engine.Group("/api/v1"))
	return engine, &http.Cookie{Name: middleware.SessionCookieName, Value: "doc-session"}
}

func (s *SellerServiceSuite) multipartDocumentRequest(docType, filename string, content []byte) (*bytes.Buffer, string) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	s.Require().NoError(writer.WriteField("documentType", docType))
	part, err := writer.CreateFormFile("file", filename)
	s.Require().NoError(err)
	_, err = part.Write(content)
	s.Require().NoError(err)
	s.Require().NoError(writer.Close())
	return &body, writer.FormDataContentType()
}

func (s *SellerServiceSuite) postDocument(engine *gin.Engine, cookie *http.Cookie, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sellers/me/documents", body)
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func (s *SellerServiceSuite) TestMultipartUploadAcceptsPDFAndSanitizesFileName() {
	s.clearDocuments()
	engine, cookie := s.documentEngine()
	body, contentType := s.multipartDocumentRequest("tax_card", "../../commercial.pdf", []byte("%PDF-1.4\n%prim\n"))

	rec := s.postDocument(engine, cookie, body, contentType)

	s.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())
	docs, err := s.service.ListMyDocuments(s.ctx, s.sellerUID)
	s.Require().NoError(err)
	s.Require().Len(docs, 1)
	s.Equal("commercial.pdf", docs[0].FileName)
	s.Equal("application/pdf", docs[0].ContentType)
	s.Equal(model.SellerDocumentStatusPending, docs[0].Status)
}

func (s *SellerServiceSuite) TestMultipartUploadRejectsUnsupportedMIME() {
	s.clearDocuments()
	engine, cookie := s.documentEngine()
	body, contentType := s.multipartDocumentRequest("tax_card", "photo.gif", []byte("GIF89a\x01\x00\x01\x00"))

	rec := s.postDocument(engine, cookie, body, contentType)

	s.Require().Equal(http.StatusBadRequest, rec.Code, rec.Body.String())
	docs, err := s.service.ListMyDocuments(s.ctx, s.sellerUID)
	s.Require().NoError(err)
	s.Empty(docs)
}

func (s *SellerServiceSuite) TestMultipartUploadRejectsEmptyAndOversizedFiles() {
	s.clearDocuments()
	engine, cookie := s.documentEngine()

	emptyBody, emptyType := s.multipartDocumentRequest("tax_card", "empty.pdf", nil)
	emptyRec := s.postDocument(engine, cookie, emptyBody, emptyType)
	s.Require().Equal(http.StatusBadRequest, emptyRec.Code, emptyRec.Body.String())

	big := bytes.Repeat([]byte("%PDF-1.4\n"), int(sellers.MaxKYCDocumentBytes)/9+2)
	bigBody, bigType := s.multipartDocumentRequest("national_id", "large.pdf", big)
	bigRec := s.postDocument(engine, cookie, bigBody, bigType)
	s.Require().Equal(http.StatusBadRequest, bigRec.Code, bigRec.Body.String())

	docs, err := s.service.ListMyDocuments(s.ctx, s.sellerUID)
	s.Require().NoError(err)
	s.Empty(docs)
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
