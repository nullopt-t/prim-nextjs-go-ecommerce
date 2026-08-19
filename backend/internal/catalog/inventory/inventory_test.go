package inventory_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/inventory"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type InventoryTestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *database.DB
	service     *inventory.InventoryService
	repo        *inventory.InventoryRepository
	categoryID  uuid.UUID
	productID   uuid.UUID
}

func (s *InventoryTestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("prim_inventory_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(s.T(), err)
	s.pgContainer = pgContainer

	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)

	db, err := database.ConnectDB(ctx, config.Config{
		DBCfg: config.DatabaseConfig{
			DBHost:     host,
			DBPort:     port.Port(),
			DBUser:     "testuser",
			DBPassword: "testpass",
			DBName:     "prim_inventory_test",
		},
	})
	require.NoError(s.T(), err)
	s.db = db

	schemaBytes, err := os.ReadFile("../../../migrations/000001_init_sechema.up.sql")
	require.NoError(s.T(), err)
	_, err = db.Exec(ctx, string(schemaBytes))
	require.NoError(s.T(), err)

	txRunner := database.NewTxRunner(s.db)
	logger := log.NewConsoleLogger()
	s.repo = inventory.NewRepository()
	s.service = inventory.NewService(logger, txRunner, s.repo)

	// Create base Category and Product for Foreign Key relations
	s.categoryID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO product_categories (id, public_id, name)
		VALUES ($1, $2, $3)
	`, s.categoryID, uuid.New(), "Test Category")
	require.NoError(s.T(), err)

	s.productID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO products (id, category_id, slug, title, status, product_type)
		VALUES ($1, $2, $3, $4, 'published', 'simple')
	`, s.productID, s.categoryID, "test-product-"+uuid.NewString(), "Test Product")
	require.NoError(s.T(), err)
}

func (s *InventoryTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *InventoryTestSuite) createTestVariant() uuid.UUID {
	variantID := uuid.New()
	sku := "SKU-" + uuid.NewString()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO product_variants (id, sku, product_id, title)
		VALUES ($1, $2, $3, $4)
	`, variantID, sku, s.productID, "Test Variant "+sku)
	require.NoError(s.T(), err)
	return variantID
}

func (s *InventoryTestSuite) createTestCart() uuid.UUID {
	cartID := uuid.New()
	sessionID := "session-" + uuid.NewString()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO carts (id, session_id)
		VALUES ($1, $2)
	`, cartID, sessionID)
	require.NoError(s.T(), err)
	return cartID
}

// =========================================================================
// 1. Core Stock Calculation (ATP & On-Hand)
// =========================================================================

func (s *InventoryTestSuite) Test01_BaselineStockCalculation() {
	ctx := context.Background()
	variantID := s.createTestVariant()

	// 1. Restock +100
	stock, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  100,
		Reason:    "restock",
	})
	s.Require().NoError(err)
	s.Equal(100, stock.OnHandQuantity)
	s.Equal(0, stock.ReservedQuantity)
	s.Equal(100, stock.AvailableQuantity)
	s.True(stock.IsInStock)

	// 2. Negative ledger (adjustment / damage) -20
	stock, err = s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  -20,
		Reason:    "adjustment",
	})
	s.Require().NoError(err)
	s.Equal(80, stock.OnHandQuantity)
	s.Equal(0, stock.ReservedQuantity)
	s.Equal(80, stock.AvailableQuantity)
	s.True(stock.IsInStock)

	// 3. Negative ledger (sale) -10
	stock, err = s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  -10,
		Reason:    "sale",
	})
	s.Require().NoError(err)
	s.Equal(70, stock.OnHandQuantity)
	s.Equal(0, stock.ReservedQuantity)
	s.Equal(70, stock.AvailableQuantity)
}

func (s *InventoryTestSuite) Test01_ReservationFiltering() {
	ctx := context.Background()
	variantID := s.createTestVariant()
	cartID := s.createTestCart()

	// Initial on-hand = 50
	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  50,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// Active reservation 1: +10 (expires in 10 mins)
	res1, err := s.service.ReserveStock(ctx, inventory.ReserveStockInput{
		VariantID: variantID,
		CartID:    &cartID,
		Quantity:  10,
		Duration:  10 * time.Minute,
	})
	s.Require().NoError(err)

	// Active reservation 2: +5 (expires in 20 mins)
	cartID2 := s.createTestCart()
	res2, err := s.service.ReserveStock(ctx, inventory.ReserveStockInput{
		VariantID: variantID,
		CartID:    &cartID2,
		Quantity:  5,
		Duration:  20 * time.Minute,
	})
	s.Require().NoError(err)

	// Stock check: OnHand = 50, Reserved = 15, Available = 35
	stock, err := s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(50, stock.OnHandQuantity)
	s.Equal(15, stock.ReservedQuantity)
	s.Equal(35, stock.AvailableQuantity)

	// Released reservation: release res1
	err = s.service.ReleaseReservation(ctx, res1.ID)
	s.Require().NoError(err)

	// Stock check: OnHand = 50, Reserved = 5, Available = 45
	stock, err = s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(50, stock.OnHandQuantity)
	s.Equal(5, stock.ReservedQuantity)
	s.Equal(45, stock.AvailableQuantity)

	// Expired reservation: manually insert an expired reservation (expired 1 hour ago)
	expiredResID := uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO inventory_reservations (id, variant_id, cart_id, quantity, expires_at, created_at, released_at)
		VALUES ($1, $2, $3, $4, now() - INTERVAL '1 hour', now() - INTERVAL '2 hours', NULL)
	`, expiredResID, variantID, cartID, 15)
	s.Require().NoError(err)

	// Stock check: expired reservation is completely ignored by deduction!
	stock, err = s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(50, stock.OnHandQuantity)
	s.Equal(5, stock.ReservedQuantity) // only res2 is active
	s.Equal(45, stock.AvailableQuantity)

	_ = res2
}

func (s *InventoryTestSuite) Test01_ColdVariantZeroState() {
	ctx := context.Background()
	coldVariantID := s.createTestVariant()

	stock, err := s.service.GetStockByVariantID(ctx, coldVariantID)
	s.Require().NoError(err)
	s.NotNil(stock)
	s.Equal(0, stock.OnHandQuantity)
	s.Equal(0, stock.ReservedQuantity)
	s.Equal(0, stock.AvailableQuantity)
	s.False(stock.IsInStock)

	// Batch get with cold variants
	stocks, err := s.service.GetStockForVariants(ctx, []uuid.UUID{coldVariantID, uuid.New()})
	s.Require().NoError(err)
	s.NotNil(stocks[coldVariantID])
	s.Equal(0, stocks[coldVariantID].OnHandQuantity)
	s.Equal(0, stocks[coldVariantID].AvailableQuantity)
	s.False(stocks[coldVariantID].IsInStock)
}

// =========================================================================
// 2. Reservation Lifecycle (Cart / Checkout Holds)
// =========================================================================

func (s *InventoryTestSuite) Test02_ReservationLifecycle() {
	ctx := context.Background()
	variantID := s.createTestVariant()
	cartID := s.createTestCart()

	// Stock = 20
	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  20,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// 1. Successful reservation of 5
	res, err := s.service.ReserveStock(ctx, inventory.ReserveStockInput{
		VariantID: variantID,
		CartID:    &cartID,
		Quantity:  5,
		Duration:  15 * time.Minute,
	})
	s.Require().NoError(err)
	s.NotNil(res)
	s.True(res.ExpiresAt.After(time.Now()))
	s.Equal(5, res.Quantity)

	// 2. Insufficient stock rejection: available is 15, requesting 16
	_, err = s.service.ReserveStock(ctx, inventory.ReserveStockInput{
		VariantID: variantID,
		Quantity:  16,
		Duration:  15 * time.Minute,
	})
	s.Require().Error(err)

	// 3. Reservation extension / upsert: User adds 3 more items to the same cart
	resUpdated, err := s.service.ReserveStock(ctx, inventory.ReserveStockInput{
		VariantID: variantID,
		CartID:    &cartID,
		Quantity:  3,
		Duration:  20 * time.Minute,
	})
	s.Require().NoError(err)
	s.Equal(res.ID, resUpdated.ID)
	s.Equal(8, resUpdated.Quantity) // 5 + 3 = 8

	stock, err := s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(20, stock.OnHandQuantity)
	s.Equal(8, stock.ReservedQuantity)
	s.Equal(12, stock.AvailableQuantity)

	// 4. Explicit Release (item removal from cart)
	err = s.service.ReleaseCartReservations(ctx, cartID)
	s.Require().NoError(err)

	stock, err = s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(20, stock.OnHandQuantity)
	s.Equal(0, stock.ReservedQuantity)
	s.Equal(20, stock.AvailableQuantity)
}

func (s *InventoryTestSuite) Test02_OrderFulfillmentAtomicConversion() {
	ctx := context.Background()
	variantID := s.createTestVariant()
	cartID := s.createTestCart()

	// Initial on_hand = 100
	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  100,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// User reserves 7 items
	res, err := s.service.ReserveStock(ctx, inventory.ReserveStockInput{
		VariantID: variantID,
		CartID:    &cartID,
		Quantity:  7,
		Duration:  15 * time.Minute,
	})
	s.Require().NoError(err)

	stockBefore, err := s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(100, stockBefore.OnHandQuantity)
	s.Equal(7, stockBefore.ReservedQuantity)
	s.Equal(93, stockBefore.AvailableQuantity)

	// Atomic Order Fulfillment / Commit Reservation
	orderRef := "ORD-9901"
	err = s.service.CommitReservation(ctx, res.ID, &orderRef)
	s.Require().NoError(err)

	// Verify post-checkout state:
	// on_hand drops by 7 (100 -> 93)
	// active reservations drop by 7 (7 -> 0)
	// available remains strictly identical (93 == 93)
	stockAfter, err := s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(93, stockAfter.OnHandQuantity)
	s.Equal(0, stockAfter.ReservedQuantity)
	s.Equal(93, stockAfter.AvailableQuantity)
	s.Equal(stockBefore.AvailableQuantity, stockAfter.AvailableQuantity)
}

// =========================================================================
// 3. Concurrency & Race Conditions (Critical)
// =========================================================================

func (s *InventoryTestSuite) Test03_ThunderingHerdOverselling() {
	ctx := context.Background()
	variantID := s.createTestVariant()

	// Setup: on_hand = 1
	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  1,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// Fire 50 concurrent reservation requests of qty = 1 simultaneously
	const concurrency = 50
	var successCount int32
	var failCount int32
	var wg sync.WaitGroup
	wg.Add(concurrency)

	startGate := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			<-startGate // Align goroutines for simultaneous execution

			_, rErr := s.service.ReserveStock(context.Background(), inventory.ReserveStockInput{
				VariantID: variantID,
				Quantity:  1,
				Duration:  10 * time.Minute,
			})
			if rErr == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}()
	}

	close(startGate)
	wg.Wait()

	// Exactly 1 succeeds, 49 fail with insufficient stock
	s.Equal(int32(1), successCount, "Exactly one reservation should succeed")
	s.Equal(int32(concurrency-1), failCount, "All other concurrent reservations must fail")

	// Final available must be 0
	stock, err := s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(1, stock.OnHandQuantity)
	s.Equal(1, stock.ReservedQuantity)
	s.Equal(0, stock.AvailableQuantity)
	s.False(stock.IsInStock)
}

func (s *InventoryTestSuite) Test03_ConcurrentCheckoutsLowStock() {
	ctx := context.Background()
	variantID := s.createTestVariant()

	// Setup on_hand = 10
	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  10,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// Create 10 valid reservations of qty = 1
	reservations := make([]*model.InventoryReservation, 10)
	for i := 0; i < 10; i++ {
		res, err := s.service.ReserveStock(ctx, inventory.ReserveStockInput{
			VariantID: variantID,
			Quantity:  1,
			Duration:  10 * time.Minute,
		})
		s.Require().NoError(err)
		reservations[i] = res
	}

	// 10 concurrent checkouts at the exact same millisecond
	var wg sync.WaitGroup
	wg.Add(10)
	startGate := make(chan struct{})

	for i := 0; i < 10; i++ {
		resID := reservations[i].ID
		orderRef := fmt.Sprintf("ORD-BATCH-%d", i)
		go func() {
			defer wg.Done()
			<-startGate
			err := s.service.CommitReservation(context.Background(), resID, &orderRef)
			s.Require().NoError(err)
		}()
	}

	close(startGate)
	wg.Wait()

	stock, err := s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(0, stock.OnHandQuantity)
	s.Equal(0, stock.ReservedQuantity)
	s.Equal(0, stock.AvailableQuantity)
}

func (s *InventoryTestSuite) Test03_CheckoutVsExpirationRace() {
	ctx := context.Background()
	variantID := s.createTestVariant()

	// Restock 5
	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  5,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// Insert an expired reservation
	expiredResID := uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO inventory_reservations (id, variant_id, quantity, expires_at, created_at, released_at)
		VALUES ($1, $2, $3, now() - INTERVAL '1 second', now() - INTERVAL '10 seconds', NULL)
	`, expiredResID, variantID, 2)
	s.Require().NoError(err)

	// Attempting to commit an expired reservation must fail safely and roll back
	ref := "ORD-EXPIRED"
	err = s.service.CommitReservation(ctx, expiredResID, &ref)
	s.Require().Error(err)

	// Ledger should remain 5 and available should remain 5
	stock, err := s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(5, stock.OnHandQuantity)
	s.Equal(5, stock.AvailableQuantity)
}

// =========================================================================
// 4. Background Expiration & Janitor Jobs
// =========================================================================

func (s *InventoryTestSuite) Test04_SoftExpiryVsHardCleanup() {
	ctx := context.Background()
	variantID := s.createTestVariant()

	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  30,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// Insert expired reservation (soft expiry)
	resID := uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO inventory_reservations (id, variant_id, quantity, expires_at, created_at, released_at)
		VALUES ($1, $2, $3, now() - INTERVAL '5 minutes', now() - INTERVAL '20 minutes', NULL)
	`, resID, variantID, 10)
	s.Require().NoError(err)

	// Read query immediately considers stock available even before janitor runs
	stock, err := s.service.GetStockByVariantID(ctx, variantID)
	s.Require().NoError(err)
	s.Equal(30, stock.OnHandQuantity)
	s.Equal(0, stock.ReservedQuantity)
	s.Equal(30, stock.AvailableQuantity)

	// Janitor Sweep
	releasedCount, err := s.service.ReleaseExpiredReservations(ctx)
	s.Require().NoError(err)
	s.True(releasedCount >= 1)

	// Verify hard cleanup in DB (released_at is now set)
	var releasedAt *time.Time
	err = s.db.QueryRow(ctx, `SELECT released_at FROM inventory_reservations WHERE id = $1`, resID).Scan(&releasedAt)
	s.Require().NoError(err)
	s.NotNil(releasedAt)
}

// =========================================================================
// 5. Adjustments & Edge Cases
// =========================================================================

func (s *InventoryTestSuite) Test05_WriteOffsAndDamage() {
	ctx := context.Background()
	variantID := s.createTestVariant()

	// Restock 100
	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  100,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// Write-off / damage adjustment -15
	damageRef := "DAMAGE-LOG-44"
	stock, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID:   variantID,
		Quantity:    -15,
		Reason:      "adjustment",
		ReferenceID: &damageRef,
	})
	s.Require().NoError(err)
	s.Equal(85, stock.OnHandQuantity)
	s.Equal(85, stock.AvailableQuantity)
}

func (s *InventoryTestSuite) Test05_ReturnsAndRefunds() {
	ctx := context.Background()
	variantID := s.createTestVariant()

	// Initial stock 10
	_, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  10,
		Reason:    "restock",
	})
	s.Require().NoError(err)

	// Process customer return +2
	returnRef := "RET-505"
	stock, err := s.service.AdjustStock(ctx, inventory.AdjustStockInput{
		VariantID:   variantID,
		Quantity:    2,
		Reason:      "return",
		ReferenceID: &returnRef,
	})
	s.Require().NoError(err)
	s.Equal(12, stock.OnHandQuantity)
	s.Equal(12, stock.AvailableQuantity)
}

func (s *InventoryTestSuite) Test05_DatabaseConstraintViolations() {
	ctx := context.Background()
	variantID := s.createTestVariant()

	// 1. quantity = 0 into inventory_ledgers -> CHECK (quantity <> 0) fails
	_, err := s.db.Exec(ctx, `
		INSERT INTO inventory_ledgers (id, variant_id, quantity, reason)
		VALUES ($1, $2, 0, 'restock')
	`, uuid.New(), variantID)
	s.Require().Error(err)
	var pgErr *pgconn.PgError
	s.True(s.ErrorAs(err, &pgErr))
	s.Equal("23514", pgErr.Code) // check_violation

	// 2. quantity <= 0 into inventory_reservations -> CHECK (quantity > 0) fails
	_, err = s.db.Exec(ctx, `
		INSERT INTO inventory_reservations (id, variant_id, quantity, expires_at)
		VALUES ($1, $2, -5, now() + INTERVAL '10 minutes')
	`, uuid.New(), variantID)
	s.Require().Error(err)
	s.True(s.ErrorAs(err, &pgErr))
	s.Equal("23514", pgErr.Code) // check_violation

	// 3. Invalid variant_id -> Foreign Key constraint failure
	nonExistentVariantID := uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO inventory_ledgers (id, variant_id, quantity, reason)
		VALUES ($1, $2, 10, 'restock')
	`, uuid.New(), nonExistentVariantID)
	s.Require().Error(err)
	s.True(s.ErrorAs(err, &pgErr))
	s.Equal("23503", pgErr.Code) // foreign_key_violation
}

func TestInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryTestSuite))
}
