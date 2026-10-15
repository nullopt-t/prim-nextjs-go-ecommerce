package admin

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type AdminService struct {
	txRunner database.Runner
	repo     *AdminRepository
}

func NewService(txRunner database.Runner, repo *AdminRepository) *AdminService {
	return &AdminService{
		txRunner: txRunner,
		repo:     repo,
	}
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

func (s *AdminService) ListRecentLedgers(ctx context.Context, limit int) ([]InventoryLedgerItem, error) {
	var list []InventoryLedgerItem
	err := s.txRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		list, err = s.repo.ListRecentLedgers(ctx, db, limit)
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
