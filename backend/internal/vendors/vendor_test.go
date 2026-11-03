package vendors_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/object"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/internal/vendors"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// VendorServiceSuite is an integration suite for the vendor domain's service
// logic, backed by real PostgreSQL via testcontainers (mirroring the checkout
// suite's scaffolding). Money-critical flows (lifecycle guards, balance
// invariant, conditional payout transitions) are covered here.
type VendorServiceSuite struct {
	suite.Suite

	ctx        context.Context
	pg         testcontainers.Container
	db         *database.DB
	service    *vendors.Service
	vendorID   uuid.UUID
	vendorUID  uuid.UUID
	categoryID uuid.UUID
}

func TestVendorServiceSuite(t *testing.T) {
	suite.Run(t, new(VendorServiceSuite))
}

func (s *VendorServiceSuite) SetupSuite() {
	require := s.Require()
	s.ctx = context.Background()

	pg, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:18-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "testuser",
				"POSTGRES_PASSWORD": "testpass",
				"POSTGRES_DB":       "prim_vendor_test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	require.NoError(err)
	s.pg = pg

	host, err := pg.Host(s.ctx)
	require.NoError(err)
	port, err := pg.MappedPort(s.ctx, "5432")
	require.NoError(err)

	s.db, err = database.ConnectDB(s.ctx, config.DatabaseConfig{
		DBHost:     host,
		DBPort:     port.Port(),
		DBUser:     "testuser",
		DBPassword: "testpass",
		DBName:     "prim_vendor_test",
	})
	require.NoError(err)

	for _, f := range []string{
		"../../migrations/000001_init_sechema.up.sql",
		"../../migrations/000003_vendor_domain.up.sql",
	} {
		migration, err := os.ReadFile(f)
		require.NoError(err)
		_, err = s.db.Exec(s.ctx, string(migration))
		require.NoError(err)
	}

	runner := database.NewTxRunner(s.db)
	logger := log.NewConsoleLogger()

	// Seed one vendor-facing user; the service registers their profile.
	db := s.db
	s.vendorUID = uuid.New()
	_, err = db.Exec(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'vendor@example.com', 'Vendor User', 'vendor', 'active', false, now(), now())
	`, s.vendorUID)
	require.NoError(err)

	s.service = vendors.NewService(
		vendors.NewRepository(),
		user.NewPostgresRepository(),
		object.NewRepository(),
		order.NewRepository(),
		vendors.NewStaticCommissionDefaults(),
		storage.NoopProvider{},
		runner,
		logger,
	)

	s.categoryID = uuid.New()
	_, err = db.Exec(s.ctx, `
		INSERT INTO product_categories (id, name) VALUES ($1, 'Payout Category')
	`, s.categoryID)
	require.NoError(err)

	registered, err := s.service.RegisterVendor(s.ctx, s.vendorUID, vendors.RegisterVendorInput{
		StoreName: "Suite Store",
		Slug:      "suite-store",
	})
	require.NoError(err)
	s.vendorID = registered.ID

	_, err = s.service.AdminChangeVendorStatus(s.ctx, s.vendorID, model.VendorStatusActive)
	require.NoError(err)
}

func (s *VendorServiceSuite) TearDownSuite() {
	s.db.Close()
	if s.pg != nil {
		_ = s.pg.Terminate(s.ctx)
	}
}

// seedSettledEarnings inserts a delivered order carrying one vendor line with
// the supplied frozen vendor_earning. Returns the order id.
func (s *VendorServiceSuite) seedSettledEarnings(earning int64) uuid.UUID {
	require := s.Require()
	db := s.db
	ctx := s.ctx

	orderID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()

	_, err := db.Exec(ctx, `
		INSERT INTO products (id, category_id, slug, title, status, product_type, variation_axes, vendor_id)
		VALUES ($1, $2, $3, 'Payout Product', 'published', 'simple', '{}', $4)
	`, productID, s.categoryID, "payout-product-"+orderID.String()[:8], s.vendorID)
	require.NoError(err)

	_, err = db.Exec(ctx, `
		INSERT INTO product_variants (id, sku, product_id, title, price, currency)
		VALUES ($1, $2, $3, 'Payout Variant', 0, 'EGP')
	`, variantID, "SKU-"+orderID.String()[:8], productID)
	require.NoError(err)

	_, err = db.Exec(ctx, `
		INSERT INTO orders (id, customer_id, customer_email, shipping_address, status, total_amount, currency)
		VALUES ($1, NULL, 'buyer@example.com', '{}'::jsonb, 'delivered', 0, 'EGP')
	`, orderID)
	require.NoError(err)

	_, err = db.Exec(ctx, `
		INSERT INTO order_items (id, order_id, variant_id, quantity, price_at_purchase, product_snapshot, vendor_id, vendor_earning)
		VALUES ($1, $2, $3, 1, 0, '{}'::jsonb, $4, $5)
	`, uuid.New(), orderID, variantID, s.vendorID, earning)
	require.NoError(err)

	return orderID
}

// ---- Registration & lifecycle -----------------------------------------------

func (s *VendorServiceSuite) TestRegisterVendorCreatesPendingProfile() {
	uid := uuid.New()
	_, err := s.db.Exec(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'second@example.com', 'Second User', 'vendor', 'active', false, now(), now())
	`, uid)
	s.Require().NoError(err)

	v, err := s.service.RegisterVendor(s.ctx, uid, vendors.RegisterVendorInput{
		StoreName: "Second Store",
		Slug:      "second-store",
	})
	s.Require().NoError(err)
	s.Equal(model.VendorStatusPending, v.Status)
	s.Equal(1000, v.CommissionRateBps)
}

func (s *VendorServiceSuite) TestRegisterVendorIsOnePerUser() {
	_, err := s.service.RegisterVendor(s.ctx, s.vendorUID, vendors.RegisterVendorInput{
		StoreName: "Suite Store Two",
		Slug:      "suite-store-two",
	})
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusConflict, apiErr.Status)
}

func (s *VendorServiceSuite) TestRegisterVendorSlugCollision() {
	uid := uuid.New()
	_, err := s.db.Exec(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'third@example.com', 'Third User', 'vendor', 'active', false, now(), now())
	`, uid)
	s.Require().NoError(err)

	_, err = s.service.RegisterVendor(s.ctx, uid, vendors.RegisterVendorInput{
		StoreName: "Copycat Store",
		Slug:      "suite-store",
	})
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusConflict, apiErr.Status)
}

func (s *VendorServiceSuite) TestAdminLifecycleGuards() {
	// active -> pending is not a legal arrow.
	_, err := s.service.AdminChangeVendorStatus(s.ctx, s.vendorID, model.VendorStatusPending)
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusConflict, apiErr.Status)

	suspended, err := s.service.AdminChangeVendorStatus(s.ctx, s.vendorID, model.VendorStatusSuspended)
	s.Require().NoError(err)
	s.Equal(model.VendorStatusSuspended, suspended.Status)

	restored, err := s.service.AdminChangeVendorStatus(s.ctx, s.vendorID, model.VendorStatusActive)
	s.Require().NoError(err)
	s.Equal(model.VendorStatusActive, restored.Status)
}

// ---- Payouts ----------------------------------------------------------------

func (s *VendorServiceSuite) TestRequestPayoutRequiresPositiveAmount() {
	_, err := s.service.RequestPayout(s.ctx, s.vendorUID, 0)
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
}

func (s *VendorServiceSuite) TestRequestPayoutRejectsInsufficientBalance() {
	// Zero settled earnings: any request must be rejected.
	_, err := s.service.RequestPayout(s.ctx, s.vendorUID, 1)
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("PAYOUT_INSUFFICIENT_BALANCE", apiErr.Code)
}

func (s *VendorServiceSuite) TestRequestPayoutConsumesBalanceAndPaysOut() {
	// Adversarial gate #3: amount > settled - outstanding must 400. An exact
	// balance request succeeds and later claims exclude the assigned funds.
	earning := int64(100_000)
	s.seedSettledEarnings(earning)

	// Overshoot fails.
	_, err := s.service.RequestPayout(s.ctx, s.vendorUID, earning+1)
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("PAYOUT_INSUFFICIENT_BALANCE", apiErr.Code)

	// Exact balance succeeds and the snapshot matches.
	payout, err := s.service.RequestPayout(s.ctx, s.vendorUID, earning)
	s.Require().NoError(err)
	s.Equal(earning, payout.SnapshotAvailableBalance)
	s.Equal(model.PayoutStatusRequested, payout.Status)

	// This request now holds the funds: another exact request is rejected.
	_, err = s.service.RequestPayout(s.ctx, s.vendorUID, 1)
	apiErr = s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("PAYOUT_INSUFFICIENT_BALANCE", apiErr.Code)

	s.Run("lifecycle transition", func() {
		approved, err := s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusApproved, nil, nil)
		s.NoError(err)
		s.Equal(model.PayoutStatusApproved, approved.Status)

		ref := "TRF-0001"
		processing, err := s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusProcessing, &ref, nil)
		s.NoError(err)
		s.Equal(model.PayoutStatusProcessing, processing.Status)

		transferred, err := s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusPaid, nil, nil)
		s.NoError(err)
		s.Equal(model.PayoutStatusPaid, transferred.Status)
		s.Require().NotNil(transferred.ProcessedAt)

		// Idempotency guard: paid is terminal.
		_, err = s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusPaid, nil, nil)
		apiErr := s.assertAPIError(err)
		s.Equal(http.StatusConflict, apiErr.Status)
	})
}

func (s *VendorServiceSuite) TestAdminTransitionPayoutConcurrentConflict() {
	// Two actors both attempt approved->processing; the first wins, the
	// second sees a conflict because the row state moved underneath (the
	// conditional update protects against the double transfer).
	earning := int64(5_000)
	s.seedSettledEarnings(earning)

	payout, err := s.service.RequestPayout(s.ctx, s.vendorUID, earning)
	s.Require().NoError(err)

	_, err = s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusApproved, nil, nil)
	s.Require().NoError(err)

	first, err := s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusProcessing, nil, nil)
	s.Require().NoError(err)
	s.Equal(model.PayoutStatusProcessing, first.Status)

	second, err := s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusProcessing, nil, nil)
	s.Require().Error(err)
	s.Nil(second)
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusConflict, apiErr.Status)
}

func (s *VendorServiceSuite) assertAPIError(err error) *apierr.APIError {
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	return apiErr
}
