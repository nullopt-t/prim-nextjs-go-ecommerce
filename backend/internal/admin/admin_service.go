package admin

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type ObjectService interface {
	UploadObject(ctx context.Context, contentType string, size int64, bucket string, file io.Reader) (*model.Object, error)
}

type AdminService struct {
	txRunner      database.Runner
	repo          *AdminRepository
	objectService ObjectService
}

func NewService(txRunner database.Runner, repo *AdminRepository, objectService ObjectService) *AdminService {
	return &AdminService{
		txRunner:      txRunner,
		repo:          repo,
		objectService: objectService,
	}
}

func (s *AdminService) UploadMedia(ctx context.Context, contentType string, size int64, file io.Reader) (*MediaItem, error) {
	bucket := "catalog"
	obj, err := s.objectService.UploadObject(ctx, contentType, size, bucket, file)
	if err != nil {
		return nil, err
	}
	return &MediaItem{
		ID:          obj.ID,
		Bucket:      obj.Bucket,
		ObjectKey:   obj.Key,
		ContentType: obj.ContentType,
		FileSize:    obj.FileSize,
		Status:      string(obj.Status),
		URL:         "http://localhost:9000/" + obj.Bucket + "/" + obj.Key,
		CreatedAt:   obj.CreatedAt,
	}, nil
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats *DashboardStats
	err := s.txRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		stats, err = s.repo.GetDashboardStats(ctx, db)
		return err
	})
	return stats, err
}

func (s *AdminService) ListCoupons(ctx context.Context) ([]PromotionItem, error) {
	var list []PromotionItem
	err := s.txRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		list, err = s.repo.ListCoupons(ctx, db)
		return err
	})
	return list, err
}

func (s *AdminService) CreateCoupon(ctx context.Context, p *PromotionItem) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		if p.ID == uuid.Nil {
			p.ID = uuid.New()
		}
		if p.Status == "" {
			p.Status = "active"
		}
		return s.repo.CreateCoupon(ctx, tx, p)
	})
}

func (s *AdminService) UpdateCouponStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.UpdateCouponStatus(ctx, tx, id, status)
	})
}

func (s *AdminService) DeleteCoupon(ctx context.Context, id uuid.UUID) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.DeleteCoupon(ctx, tx, id)
	})
}

func (s *AdminService) ListPayments(ctx context.Context, limit, offset int) ([]PaymentItem, int, error) {
	var payments []PaymentItem
	var total int
	err := s.txRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		payments, total, err = s.repo.ListPayments(ctx, db, limit, offset)
		return err
	})
	return payments, total, err
}

func (s *AdminService) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.UpdatePaymentStatus(ctx, tx, id, status)
	})
}

func (s *AdminService) RefundPayment(ctx context.Context, id uuid.UUID) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.UpdatePaymentStatus(ctx, tx, id, "refunded")
	})
}

func (s *AdminService) ListMedia(ctx context.Context, limit, offset int) ([]MediaItem, int, error) {
	var list []MediaItem
	var total int
	err := s.txRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		list, total, err = s.repo.ListMedia(ctx, db, limit, offset)
		return err
	})
	return list, total, err
}

func (s *AdminService) ListAttributes(ctx context.Context) ([]AttributeItem, error) {
	var list []AttributeItem
	err := s.txRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		list, err = s.repo.ListAttributes(ctx, db)
		return err
	})
	return list, err
}

func (s *AdminService) CreateAttribute(ctx context.Context, ai *AttributeItem) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		if ai.ID == uuid.Nil {
			ai.ID = uuid.New()
		}
		if ai.ValueType == "" {
			ai.ValueType = "text"
		}
		return s.repo.CreateAttribute(ctx, tx, ai)
	})
}

func (s *AdminService) DeleteAttribute(ctx context.Context, id uuid.UUID) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.DeleteAttribute(ctx, tx, id)
	})
}

func (s *AdminService) ListInventoryOverview(ctx context.Context, limit, offset int) ([]InventoryStockItem, int, error) {
	var items []InventoryStockItem
	var total int
	err := s.txRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		items, total, err = s.repo.ListInventoryOverview(ctx, db, limit, offset)
		return err
	})
	return items, total, err
}

func (s *AdminService) ListRecentLedgers(ctx context.Context, limit int, variantID *uuid.UUID, productID *uuid.UUID) ([]InventoryLedgerItem, error) {
	var list []InventoryLedgerItem
	err := s.txRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		list, err = s.repo.ListRecentLedgers(ctx, db, limit, variantID, productID)
		return err
	})
	return list, err
}

func (s *AdminService) AdjustInventory(ctx context.Context, variantID uuid.UUID, delta int, reason string) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		ledger := &model.InventoryLedger{
			ID:        uuid.New(),
			VariantID: variantID,
			Quantity:  delta,
			Reason:    model.InventoryReason(reason),
			CreatedAt: time.Now(),
		}
		query := `
			INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at)
			VALUES ($1, $2, $3, $4::inventory_reason, NOW())
		`
		_, err := tx.Exec(ctx, query, ledger.ID, ledger.VariantID, ledger.Quantity, ledger.Reason)
		return err
	})
}

func (s *AdminService) AdjustProductInventory(ctx context.Context, productID uuid.UUID, delta int, reason string) error {
	return s.txRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		var variantID uuid.UUID
		err := tx.QueryRow(ctx, `
			SELECT id FROM product_variants
			WHERE product_id = $1 AND deleted_at IS NULL
			ORDER BY is_default DESC, created_at ASC
			LIMIT 1
		`, productID).Scan(&variantID)
		if err != nil {
			return err
		}

		ledger := &model.InventoryLedger{
			ID:        uuid.New(),
			VariantID: variantID,
			Quantity:  delta,
			Reason:    model.InventoryReason(reason),
			CreatedAt: time.Now(),
		}
		query := `
			INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at)
			VALUES ($1, $2, $3, $4::inventory_reason, NOW())
		`
		_, err = tx.Exec(ctx, query, ledger.ID, ledger.VariantID, ledger.Quantity, ledger.Reason)
		return err
	})
}
