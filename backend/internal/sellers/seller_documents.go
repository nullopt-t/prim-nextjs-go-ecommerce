package sellers

import (
	"context"
	"errors"
	"io"
	"mime"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// MaxKYCDocumentBytes is the hard per-file upload limit (10 MiB). The minimum
// accepted size is one byte; empty files are rejected at the boundary.
const MaxKYCDocumentBytes int64 = 10 * 1024 * 1024

// sellerDocumentBucket holds verification evidence. Each multipart upload gets
// a fresh UUID key; JSON binding reuses an existing object id.
const sellerDocumentBucket = "seller-documents"

var allowedKYCDocumentMIME = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
}

// isAllowedKYCDocumentMIME enforces the strict KYC file-type allowlist. The
// media uploader's raster-image list is deliberately not reused here: KYC
// accepts PDFs and rejects GIF/AVIF/SVG.
func isAllowedKYCDocumentMIME(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return allowedKYCDocumentMIME[strings.ToLower(mediaType)]
}

// validateKYCUploadSize rejects empty and oversized uploads with a typed 400.
func validateKYCUploadSize(size int64) error {
	if size < 1 || size > MaxKYCDocumentBytes {
		return validationError("file", "size", "File must be between 1 byte and 10 MB")
	}
	return nil
}

// sanitizeDocumentFileName strips directory components from either separator,
// trims whitespace, rejects dot-only names, and clamps to 255 bytes.
func sanitizeDocumentFileName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "/")
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return "", validationError("fileName", "format", "A valid file name is required")
	}
	if len(name) > 255 {
		name = name[:255]
	}
	return name, nil
}

// CreateDocumentInput binds an already-uploaded storage object to a KYC type.
type CreateDocumentInput struct {
	DocumentType model.SellerDocumentType
	ObjectID     uuid.UUID
	FileName     string
}

// UploadDocumentInput carries a multipart file streamed straight from the
// request body. ContentType is sniffed at the boundary, never trusted from the
// client header.
type UploadDocumentInput struct {
	DocumentType model.SellerDocumentType
	FileName     string
	ContentType  string
	Size         int64
	File         io.Reader
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

// approvedDocumentLockedError is the single source of the compliance lock
// message shared by replacement and deletion guards.
func approvedDocumentLockedError() error {
	return apierr.ConflictError("Approved documents are locked for compliance and cannot be replaced").
		WithCode(errcode.CodeSellerDocumentConflict)
}

// cleanupReplacedObject removes the previous blob only when no other document
// still references it. Failures are logged, never surfaced: the DB row is
// already committed and a leftover blob is sweepable.
func (s *Service) cleanupReplacedObject(ctx context.Context, objectID *uuid.UUID) {
	if objectID == nil {
		return
	}
	var orphaned bool
	_ = s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		count, err := s.repo.DocumentCountForObject(ctx, *objectID)
		if err != nil {
			return err
		}
		orphaned = count == 0
		return nil
	})
	if orphaned {
		if err := s.objects.DeleteObjectByID(ctx, *objectID); err != nil {
			s.logger.Warn("seller document replacement blob cleanup failed", log.Meta{"object": *objectID, "err": err})
		}
	}
}

// CreateDocument binds an existing storage object to a KYC type. A pending or
// rejected document of the same type is replaced in place; an approved one is
// locked for compliance and rejected with 409.
func (s *Service) CreateDocument(ctx context.Context, userID uuid.UUID, in CreateDocumentInput) (*model.SellerDocument, error) {
	if !validDocumentType(in.DocumentType) {
		return nil, validationError("documentType", "oneof", "Unsupported document type")
	}
	fileName, err := sanitizeDocumentFileName(in.FileName)
	if err != nil {
		return nil, err
	}

	var doc *model.SellerDocument
	var oldObjectID *uuid.UUID
	txErr := s.runner.WithTx(ctx, func(ctx context.Context) error {
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
		// The JSON binding path must obey the same strict allowlist as a direct
		// upload; otherwise any media object could bypass KYC file-type rules.
		if !isAllowedKYCDocumentMIME(obj.ContentType) {
			return validationError("file", "mime", "Unsupported file type; upload a PDF, JPEG, PNG or WebP file")
		}
		if err := validateKYCUploadSize(obj.FileSize); err != nil {
			return err
		}

		existing, err := s.repo.GetDocumentBySellerAndType(ctx, seller.ID, in.DocumentType)
		if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
			return s.handleError(err, "Failed to load seller document")
		}
		row := CreateSellerDocumentInput{
			SellerID:        seller.ID,
			DocumentType:    in.DocumentType,
			StorageObjectID: obj.ID,
			FileName:        fileName,
			FileSize:        obj.FileSize,
			ContentType:     obj.ContentType,
		}
		if existing != nil {
			if existing.Status == model.SellerDocumentStatusApproved {
				return approvedDocumentLockedError()
			}
			oldObjectID = &existing.StorageObjectID
			doc, err = s.repo.ReplaceDocument(ctx, existing.ID, row)
		} else {
			doc, err = s.repo.CreateDocument(ctx, row)
		}
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
	if txErr != nil {
		return nil, s.handleError(txErr, "Failed to submit seller document")
	}
	s.cleanupReplacedObject(ctx, oldObjectID)
	return doc, nil
}

// UploadDocument streams a multipart KYC file into storage, then reuses the
// same strict upsert as CreateDocument. The freshly uploaded object is removed
// if the document write fails, so a rejected upload leaves no orphan blob.
func (s *Service) UploadDocument(ctx context.Context, userID uuid.UUID, in UploadDocumentInput) (*model.SellerDocument, error) {
	if !validDocumentType(in.DocumentType) {
		return nil, validationError("documentType", "oneof", "Unsupported document type")
	}
	if _, err := sanitizeDocumentFileName(in.FileName); err != nil {
		return nil, err
	}
	if !isAllowedKYCDocumentMIME(in.ContentType) {
		return nil, validationError("file", "mime", "Unsupported file type; upload a PDF, JPEG, PNG or WebP file")
	}
	if err := validateKYCUploadSize(in.Size); err != nil {
		return nil, err
	}

	// Fail before touching storage when the existing document is locked.
	lockErr := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		seller, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			return s.sellerForDocumentError(err)
		}
		existing, err := s.repo.GetDocumentBySellerAndType(ctx, seller.ID, in.DocumentType)
		if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
			return s.handleError(err, "Failed to load seller document")
		}
		if existing != nil && existing.Status == model.SellerDocumentStatusApproved {
			return approvedDocumentLockedError()
		}
		return nil
	})
	if lockErr != nil {
		return nil, s.handleError(lockErr, "Failed to submit seller document")
	}

	obj, err := s.objects.UploadObject(ctx, in.ContentType, in.Size, sellerDocumentBucket, in.File)
	if err != nil {
		return nil, s.handleError(err, "Failed to upload seller document")
	}

	doc, err := s.CreateDocument(ctx, userID, CreateDocumentInput{
		DocumentType: in.DocumentType,
		ObjectID:     obj.ID,
		FileName:     in.FileName,
	})
	if err != nil {
		if cleanupErr := s.objects.DeleteObjectByID(ctx, obj.ID); cleanupErr != nil {
			s.logger.Warn("seller document upload rollback cleanup failed", log.Meta{"object": obj.ID, "err": cleanupErr})
		}
		return nil, err
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
			return apierr.ConflictError("Approved documents are locked for compliance and cannot be deleted").
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
// decided and when. A rejection reason is mandatory for the rejected status.
func (s *Service) AdminReviewDocument(
	ctx context.Context,
	sellerID, documentID, reviewerID uuid.UUID,
	status model.SellerDocumentStatus,
	reason string,
) (*model.SellerDocument, error) {
	if status != model.SellerDocumentStatusApproved && status != model.SellerDocumentStatusRejected {
		return nil, validationError("status", "oneof", "Status must be approved or rejected")
	}
	trimmed := strings.TrimSpace(reason)
	if status == model.SellerDocumentStatusRejected && trimmed == "" {
		return nil, validationError("rejectionReason", "required", "Rejection reason is required when rejecting a document")
	}
	var rejectionReason *string
	if trimmed != "" {
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
