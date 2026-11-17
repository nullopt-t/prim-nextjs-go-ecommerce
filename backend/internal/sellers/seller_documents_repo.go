package sellers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// CreateSellerDocumentInput carries the KYC document row to insert.
type CreateSellerDocumentInput struct {
	SellerID        uuid.UUID
	DocumentType    model.SellerDocumentType
	StorageObjectID uuid.UUID
	FileName        string
	FileSize        int64
	ContentType     string
}

const sellerDocumentColumns = `
	id, seller_id, document_type, storage_object_id, file_name, file_size,
	content_type, status, rejection_reason, created_at, updated_at,
	reviewed_at, reviewed_by
`

func scanSellerDocument(row pgx.Row) (*model.SellerDocument, error) {
	doc := new(model.SellerDocument)
	err := row.Scan(
		&doc.ID, &doc.SellerID, &doc.DocumentType, &doc.StorageObjectID,
		&doc.FileName, &doc.FileSize, &doc.ContentType, &doc.Status,
		&doc.RejectionReason, &doc.CreatedAt, &doc.UpdatedAt,
		&doc.ReviewedAt, &doc.ReviewedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("scan seller document: %w", err)
	}
	return doc, nil
}

// ListDocuments returns a seller's KYC documents, newest first.
func (r *SellerRepository) ListDocuments(ctx context.Context, sellerID uuid.UUID) ([]*model.SellerDocument, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := executor.QueryContext(ctx, `
		SELECT `+sellerDocumentColumns+` FROM seller_documents
		WHERE seller_id = $1
		ORDER BY created_at DESC, id DESC
	`, sellerID)
	if err != nil {
		return nil, fmt.Errorf("list seller documents: %w", err)
	}
	defer rows.Close()

	docs := make([]*model.SellerDocument, 0)
	for rows.Next() {
		doc, err := scanSellerDocument(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list seller documents: %w", err)
	}
	return docs, nil
}

// GetDocumentByID loads one KYC document.
func (r *SellerRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*model.SellerDocument, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	return scanSellerDocument(executor.QueryRowContext(ctx,
		`SELECT `+sellerDocumentColumns+` FROM seller_documents WHERE id = $1`, id))
}

// GetDocumentBySellerAndType loads a seller's document of one type. The
// UNIQUE(seller_id, document_type) constraint guarantees at most one row.
func (r *SellerRepository) GetDocumentBySellerAndType(
	ctx context.Context,
	sellerID uuid.UUID,
	docType model.SellerDocumentType,
) (*model.SellerDocument, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	return scanSellerDocument(executor.QueryRowContext(ctx,
		`SELECT `+sellerDocumentColumns+` FROM seller_documents WHERE seller_id = $1 AND document_type = $2`,
		sellerID, docType))
}

// ReplaceDocument swaps the storage object and metadata of an existing
// non-approved document and resets it to a fresh pending submission.
func (r *SellerRepository) ReplaceDocument(ctx context.Context, id uuid.UUID, in CreateSellerDocumentInput) (*model.SellerDocument, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	return scanSellerDocument(executor.QueryRowContext(ctx, `
		UPDATE seller_documents SET
			storage_object_id = $2,
			file_name = $3,
			file_size = $4,
			content_type = $5,
			status = 'pending',
			rejection_reason = NULL,
			reviewed_at = NULL,
			reviewed_by = NULL,
			updated_at = now()
		WHERE id = $1
		RETURNING `+sellerDocumentColumns,
		id, in.StorageObjectID, in.FileName, in.FileSize, in.ContentType,
	))
}

// DocumentCountForObject counts documents still bound to a storage object.
// It guards replacement cleanup so a shared object is never deleted.
func (r *SellerRepository) DocumentCountForObject(ctx context.Context, objectID uuid.UUID) (int, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return 0, err
	}
	var count int
	err = executor.QueryRowContext(ctx,
		`SELECT count(*) FROM seller_documents WHERE storage_object_id = $1`, objectID).Scan(&count)
	return count, err
}

// CreateDocument inserts a KYC document. The UNIQUE(seller_id, document_type)
// constraint is the final guard against duplicate submissions.
func (r *SellerRepository) CreateDocument(ctx context.Context, in CreateSellerDocumentInput) (*model.SellerDocument, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	return scanSellerDocument(executor.QueryRowContext(ctx, `
		INSERT INTO seller_documents (
			id, seller_id, document_type, storage_object_id, file_name,
			file_size, content_type, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', now(), now())
		RETURNING `+sellerDocumentColumns,
		uuid.New(), in.SellerID, in.DocumentType, in.StorageObjectID,
		in.FileName, in.FileSize, in.ContentType,
	))
}

// DeleteDocument removes a KYC document row. Callers must have checked the
// ownership and review-status preconditions inside the same transaction.
func (r *SellerRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}
	tag, err := executor.ExecContext(ctx, `DELETE FROM seller_documents WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete seller document: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ReviewDocument records an admin decision on a KYC document.
func (r *SellerRepository) ReviewDocument(
	ctx context.Context,
	id uuid.UUID,
	status model.SellerDocumentStatus,
	reason *string,
	reviewedBy uuid.UUID,
) (*model.SellerDocument, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	return scanSellerDocument(executor.QueryRowContext(ctx, `
		UPDATE seller_documents SET
			status = $2,
			rejection_reason = $3,
			reviewed_at = now(),
			reviewed_by = $4,
			updated_at = now()
		WHERE id = $1
		RETURNING `+sellerDocumentColumns,
		id, status, reason, reviewedBy,
	))
}
