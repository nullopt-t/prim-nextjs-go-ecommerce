package sellers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// CreateDocumentInput binds an already-uploaded storage object to a KYC type.
type CreateDocumentInput struct {
	DocumentType model.SellerDocumentType
	ObjectID     uuid.UUID
	FileName     string
}

func validDocumentType(t model.SellerDocumentType) bool {
	switch t {
	case model.SellerDocumentTypeCommercialRegister, model.SellerDocumentTypeTaxCard,
		model.SellerDocumentTypeNationalID, model.SellerDocumentTypeOther:
		return true
	}
	return false
}

func (s *Service) sellerForDocumentError(err error) error {
	if errors.Is(database.MapError(err), database.ErrNotFound) {
		return apierr.NotFoundError("Seller not found").WithCode(errcode.CodeSellerNotFound)
	}
	return s.handleError(err, "Failed to load seller")
}

func (s *Service) documentNotFoundError() error {
	return apierr.NotFoundError("Seller document not found").WithCode(errcode.CodeSellerDocumentNotFound)
}

// attachDocumentURLs resolves each document's public URL from its storage
// object. A missing object leaves the URL empty rather than failing the list.
func (s *Service) attachDocumentURLs(ctx context.Context, docs []*model.SellerDocument) {
	for _, doc := range docs {
		obj, err := s.objects.GetByID(ctx, doc.StorageObjectID)
		if err == nil && obj != nil {
			doc.PublicURL = s.storage.GetPublicURL(obj.Bucket, obj.Key)
		}
	}
}

// ListMyDocuments returns the caller's KYC documents with public URLs.
func (s *Service) ListMyDocuments(ctx context.Context, userID uuid.UUID) ([]*model.SellerDocument, error) {
	var docs []*model.SellerDocument
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		seller, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			return s.sellerForDocumentError(err)
		}
		docs, err = s.repo.ListDocuments(ctx, seller.ID)
		if err != nil {
			return s.handleError(err, "Failed to list seller documents")
		}
		s.attachDocumentURLs(ctx, docs)
		return nil
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to list seller documents")
	}
	return docs, nil
}

// CreateDocument records one KYC document for the caller's seller profile.
// The storage object must already exist; its metadata becomes the document's
// file name, size, and content type.
func (s *Service) CreateDocument(ctx context.Context, userID uuid.UUID, in CreateDocumentInput) (*model.SellerDocument, error) {
	if !validDocumentType(in.DocumentType) {
		return nil, validationError("documentType", "oneof", "Unsupported document type")
	}
	in.FileName = strings.TrimSpace(in.FileName)
	if in.FileName == "" || len(in.FileName) > 255 {
		return nil, validationError("fileName", "length", "File name must be between 1 and 255 characters")
	}

	var doc *model.SellerDocument
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			return s.sellerForDocumentError(err)
		}
		obj, err := s.objects.GetByID(ctx, in.ObjectID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.BadRequestError("Storage object not found").WithCode(apierr.CodeInvalidReference)
			}
			return s.handleError(err, "Failed to load storage object")
		}
		if obj.DeletedAt != nil || obj.Status == model.ObjectStatusDeleted {
			return apierr.BadRequestError("Storage object is no longer available").WithCode(apierr.CodeInvalidReference)
		}

		doc, err = s.repo.CreateDocument(ctx, CreateSellerDocumentInput{
			SellerID:        seller.ID,
			DocumentType:    in.DocumentType,
			StorageObjectID: obj.ID,
			FileName:        in.FileName,
			FileSize:        obj.FileSize,
			ContentType:     obj.ContentType,
		})
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrConflict) {
				return apierr.ConflictError("A document of this type was already submitted").
					WithCode(errcode.CodeSellerDocumentAlreadyExists)
			}
			return s.handleError(err, "Failed to submit seller document")
		}
		doc.PublicURL = s.storage.GetPublicURL(obj.Bucket, obj.Key)
		return nil
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to submit seller document")
	}
	return doc, nil
}

// DeleteMyDocument removes a seller's pending or rejected document. Approved
// evidence is kept for compliance. The blob is deleted only after the DB row
// is gone, so a storage failure leaves at most a sweepable orphan.
func (s *Service) DeleteMyDocument(ctx context.Context, userID, documentID uuid.UUID) error {
	var obj *model.Object
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			return s.sellerForDocumentError(err)
		}
		doc, err := s.repo.GetDocumentByID(ctx, documentID)
		if err != nil || doc.SellerID != seller.ID {
			return s.documentNotFoundError()
		}
		if doc.Status != model.SellerDocumentStatusPending && doc.Status != model.SellerDocumentStatusRejected {
			return apierr.ConflictError("Only pending or rejected documents can be deleted").
				WithCode(errcode.CodeSellerDocumentConflict)
		}
		obj, _ = s.objects.GetByID(ctx, doc.StorageObjectID)
		return s.repo.DeleteDocument(ctx, documentID)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete seller document")
	}

	if obj != nil && obj.DeletedAt == nil {
		if err := s.storage.Delete(ctx, obj.Bucket, obj.Key); err != nil {
			s.logger.Warn("seller document blob cleanup failed", log.Meta{"document": documentID, "err": err})
		}
	}
	return nil
}

// AdminListSellerDocuments returns all KYC documents of one seller.
func (s *Service) AdminListSellerDocuments(ctx context.Context, sellerID uuid.UUID) ([]*model.SellerDocument, error) {
	var docs []*model.SellerDocument
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		if _, err := s.repo.GetByID(ctx, sellerID); err != nil {
			return s.sellerForDocumentError(err)
		}
		var err error
		docs, err = s.repo.ListDocuments(ctx, sellerID)
		if err != nil {
			return s.handleError(err, "Failed to list seller documents")
		}
		s.attachDocumentURLs(ctx, docs)
		return nil
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to list seller documents")
	}
	return docs, nil
}

// AdminReviewDocument approves or rejects one KYC document and records who
// decided and when. A rejection reason is optional and stored when supplied.
func (s *Service) AdminReviewDocument(
	ctx context.Context,
	sellerID, documentID, reviewerID uuid.UUID,
	status model.SellerDocumentStatus,
	reason string,
) (*model.SellerDocument, error) {
	if status != model.SellerDocumentStatusApproved && status != model.SellerDocumentStatusRejected {
		return nil, validationError("status", "oneof", "Status must be approved or rejected")
	}
	var rejectionReason *string
	if trimmed := strings.TrimSpace(reason); trimmed != "" {
		if len(trimmed) > 1000 {
			return nil, validationError("rejectionReason", "max", "Rejection reason must be at most 1000 characters")
		}
		rejectionReason = &trimmed
	}

	var doc *model.SellerDocument
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetDocumentByID(ctx, documentID)
		if err != nil || existing.SellerID != sellerID {
			return s.documentNotFoundError()
		}
		doc, err = s.repo.ReviewDocument(ctx, documentID, status, rejectionReason, reviewerID)
		if err != nil {
			return s.handleError(err, "Failed to review seller document")
		}
		s.attachDocumentURLs(ctx, []*model.SellerDocument{doc})
		return nil
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to review seller document")
	}
	return doc, nil
}
