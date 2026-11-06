package order_test

// Database-backed order lifecycle tests. These exercise the real order
// repository and inventory writer against PostgreSQL so cancellation and
// payment transitions are verified at the SQL/transaction boundary (row locks,
// compensating ledger entries, and rollback), not only with fakes.

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

type OrderLifecycleSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *database.DB
	runner      database.Runner
	inventory   *inventory.InventoryService
	service     *order.OrderService
	categoryID  uuid.UUID
	productID   uuid.UUID
}

func (s *OrderLifecycleSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_order_lifecycle_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	require.NoError(s.T(), err)
	s.pgContainer = pgContainer

	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)

	db, err := database.ConnectDB(ctx, config.DatabaseConfig{
		DBHost: host, DBPort: port.Port(),
		DBUser: "testuser", DBPassword: "testpass", DBName: "prim_order_lifecycle_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	for _, migration := range []string{
		"../../migrations/000001_init_sechema.up.sql",
		"../../migrations/000003_vendor_domain.up.sql",
	} {
		schema, readErr := os.ReadFile(migration)
		require.NoError(s.T(), readErr)
		_, execErr := db.Exec(ctx, string(schema))
		require.NoError(s.T(), execErr)
	}

	s.runner = database.NewTxRunner(db)
	logger := log.NewNoopLogger()
	inventoryRepo := inventory.NewRepository()
	s.inventory = inventory.NewService(logger, s.runner, inventoryRepo, storage.NoopProvider{}, nil)
	s.service = order.NewService(
		s.runner,
		order.NewRepository(),
		logger,
		inventoryRepo,
		s.inventory,
		user.NewPostgresRepository(),
		setting.NewService(s.runner, setting.NewRepository()),
		&fakeOrderNotifier{},
	)

	s.categoryID = uuid.New()
	_, err = db.Exec(ctx, `INSERT INTO product_categories (id, name) VALUES ($1, $2)`, s.categoryID, "Lifecycle")
	require.NoError(s.T(), err)

	s.productID = uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO products (id, category_id, slug, title, status, product_type, variation_axes)
		VALUES ($1, $2, $3, $4, 'published', 'variable', ARRAY['color']::text[])
	`, s.productID, s.categoryID, "lifecycle-"+uuid.NewString(), "Lifecycle Product")
	require.NoError(s.T(), err)
}

func (s *OrderLifecycleSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *OrderLifecycleSuite) seedUser() uuid.UUID {
	id := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO users (id, email, full_name, role, is_email_verified, status)
		VALUES ($1, $2, $3, 'customer', true, 'active')
	`, id, "order_lifecycle_"+id.String()+"@example.com", "Lifecycle Customer")
	s.Require().NoError(err)
	return id
}

func (s *OrderLifecycleSuite) seedVariant() uuid.UUID {
	id := uuid.New()
	sku := "ORD-" + uuid.NewString()
	attributes := `{"color":"` + sku + `"}`
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO product_variants (id, sku, product_id, title, price, currency, attributes)
		VALUES ($1, $2, $3, 'Default', 1000, 'EGP', $4::jsonb)
	`, id, sku, s.productID, attributes)
	s.Require().NoError(err)
	return id
}

func (s *OrderLifecycleSuite) stockVariant(variantID uuid.UUID, quantity int) {
	_, err := s.inventory.AdjustStock(context.Background(), inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  quantity,
		Reason:    "restock",
	})
	s.Require().NoError(err)
}

func (s *OrderLifecycleSuite) seedOrder(customerID uuid.UUID, status model.OrderStatus) uuid.UUID {
	orderID := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO orders (id, customer_id, customer_email, shipping_address, status, total_amount, currency)
		VALUES ($1, $2, 'lifecycle@example.com', '{}'::jsonb, $3, 1000, 'EGP')
	`, orderID, customerID, string(status))
	s.Require().NoError(err)
	return orderID
}

func (s *OrderLifecycleSuite) seedOrderItem(orderID, variantID uuid.UUID, quantity int) {
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO order_items (id, order_id, variant_id, quantity, price_at_purchase, product_snapshot)
		VALUES ($1, $2, $3, $4, 1000, '{}'::jsonb)
	`, uuid.New(), orderID, variantID, quantity)
	s.Require().NoError(err)
}

func (s *OrderLifecycleSuite) seedPayment(orderID uuid.UUID, status string) uuid.UUID {
	id := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO payments (id, order_id, amount, currency, status, provider, provider_transaction_id)
		VALUES ($1, $2, 1000, 'EGP', $3, 'test', $4)
	`, id, orderID, status, "TX-"+uuid.NewString())
	s.Require().NoError(err)
	return id
}

func (s *OrderLifecycleSuite) paymentStatus(paymentID uuid.UUID) string {
	var status string
	s.Require().NoError(s.db.QueryRow(context.Background(),
		`SELECT status::text FROM payments WHERE id = $1`, paymentID).Scan(&status))
	return status
}

func (s *OrderLifecycleSuite) orderStatus(orderID uuid.UUID) string {
	var status string
	s.Require().NoError(s.db.QueryRow(context.Background(),
		`SELECT status::text FROM orders WHERE id = $1`, orderID).Scan(&status))
	return status
}

func (s *OrderLifecycleSuite) ledgerCount(variantID uuid.UUID) int {
	var count int
	s.Require().NoError(s.db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM inventory_ledgers WHERE variant_id = $1`, variantID).Scan(&count))
	return count
}

func (s *OrderLifecycleSuite) historyCount(orderID uuid.UUID) int {
	var count int
	s.Require().NoError(s.db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM order_status_history WHERE order_id = $1`, orderID).Scan(&count))
	return count
}

func (s *OrderLifecycleSuite) TestCancelOrder_AfterShipmentRejected() {
	customerID := s.seedUser()
	variantID := s.seedVariant()
	s.stockVariant(variantID, 5)
	orderID := s.seedOrder(customerID, model.OrderStatusShipped)
	s.seedOrderItem(orderID, variantID, 1)
	ledgersBefore := s.ledgerCount(variantID)

	err := s.service.CancelOrder(context.Background(), orderID, customerID)
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(400, apiErr.Status)
	s.Equal("INVALID_ORDER_STATUS", apiErr.Code)

	s.Equal(string(model.OrderStatusShipped), s.orderStatus(orderID))
	s.Equal(ledgersBefore, s.ledgerCount(variantID), "a rejected cancellation must not restock")
}

func (s *OrderLifecycleSuite) TestCancelOrder_ForeignCustomerForbidden() {
	ownerID := s.seedUser()
	attackerID := s.seedUser()
	orderID := s.seedOrder(ownerID, model.OrderStatusPending)

	err := s.service.CancelOrder(context.Background(), orderID, attackerID)
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(403, apiErr.Status)
	s.Equal(string(model.OrderStatusPending), s.orderStatus(orderID))
}

func (s *OrderLifecycleSuite) TestCancelOrder_AlreadyCancelledIsIdempotent() {
	customerID := s.seedUser()
	variantID := s.seedVariant()
	s.stockVariant(variantID, 10)
	orderID := s.seedOrder(customerID, model.OrderStatusPending)
	s.seedOrderItem(orderID, variantID, 2)

	s.Require().NoError(s.service.CancelOrder(context.Background(), orderID, customerID))
	ledgerAfterFirst := s.ledgerCount(variantID)
	historyAfterFirst := s.historyCount(orderID)

	// A second cancellation must be a no-op: no extra history row and, above
	// all, no second compensating restock.
	s.Require().NoError(s.service.CancelOrder(context.Background(), orderID, customerID))
	s.Equal(ledgerAfterFirst, s.ledgerCount(variantID))
	s.Equal(historyAfterFirst, s.historyCount(orderID))
	s.Equal(string(model.OrderStatusCancelled), s.orderStatus(orderID))
}

func (s *OrderLifecycleSuite) TestCancelOrder_RestocksPaidOrderOnce() {
	customerID := s.seedUser()
	variantID := s.seedVariant()
	s.stockVariant(variantID, 10)
	orderID := s.seedOrder(customerID, model.OrderStatusPaid)
	s.seedOrderItem(orderID, variantID, 3)

	s.Require().NoError(s.service.CancelOrder(context.Background(), orderID, customerID))
	s.Require().NoError(s.service.CancelOrder(context.Background(), orderID, customerID))

	var restocked int
	s.Require().NoError(s.db.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(quantity), 0)
		FROM inventory_ledgers
		WHERE variant_id = $1 AND reason = 'return'
	`, variantID).Scan(&restocked))
	s.Equal(3, restocked, "exactly one compensating restock is written")
}

func (s *OrderLifecycleSuite) TestRefundPayment_DBTransitionDoesNotTouchStockOrOrder() {
	customerID := s.seedUser()
	variantID := s.seedVariant()
	s.stockVariant(variantID, 10)
	orderID := s.seedOrder(customerID, model.OrderStatusPaid)
	s.seedOrderItem(orderID, variantID, 1)
	paymentID := s.seedPayment(orderID, "captured")
	ledgersBefore := s.ledgerCount(variantID)

	s.Require().NoError(s.service.RefundPayment(context.Background(), paymentID))
	s.Equal("refunded", s.paymentStatus(paymentID))

	// Pinned decision: refunding a payment changes only the payment ledger.
	// Stock and order status changes are owned by cancellation / status
	// transitions, so a refund must not silently restock or rewrite the order.
	s.Equal(string(model.OrderStatusPaid), s.orderStatus(orderID))
	s.Equal(ledgersBefore, s.ledgerCount(variantID))

	err := s.service.RefundPayment(context.Background(), paymentID)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(409, apiErr.Status)
}

func (s *OrderLifecycleSuite) TestRefundPayment_ConcurrentDoubleRefundOnlySucceedsOnce() {
	customerID := s.seedUser()
	orderID := s.seedOrder(customerID, model.OrderStatusPaid)
	paymentID := s.seedPayment(orderID, "captured")

	start := make(chan struct{})
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			results <- s.service.RefundPayment(context.Background(), paymentID)
		}()
	}
	close(start)

	successes, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			successes++
			continue
		}
		var apiErr *apierr.APIError
		s.Require().ErrorAs(err, &apiErr)
		s.Equal(409, apiErr.Status)
		conflicts++
	}
	s.Equal(1, successes)
	s.Equal(1, conflicts)
	s.Equal("refunded", s.paymentStatus(paymentID))
}

func (s *OrderLifecycleSuite) TestUpdatePaymentStatus_CaptureFailureRollsBackPayment() {
	customerID := s.seedUser()
	variantID := s.seedVariant()
	orderID := s.seedOrder(customerID, model.OrderStatusPending)
	paymentID := s.seedPayment(orderID, "pending")

	// A live hold was written without any on-hand stock behind it, so
	// committing the sale fails on insufficient physical stock.
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO inventory_reservations (id, variant_id, order_id, quantity, expires_at)
		VALUES ($1, $2, $3, 5, now() + interval '15 minutes')
	`, uuid.New(), variantID, orderID)
	s.Require().NoError(err)

	err = s.service.UpdatePaymentStatus(context.Background(), paymentID, "captured")
	s.Require().Error(err)

	// The failed capture must roll back the payment status with it.
	s.Equal("pending", s.paymentStatus(paymentID))
}

func TestOrderLifecycleSuite(t *testing.T) {
	suite.Run(t, new(OrderLifecycleSuite))
}
