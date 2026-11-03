package vendors_test

import (
	"context"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/object"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/internal/vendors"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
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
	catalog    *fakeProductCatalog
	variants   *fakeVariantCatalog
	stock      *fakeStockAdjuster
}

// fakeVariantCatalog is an in-memory VariantCatalog double used to prove the
// ownership gate before delegation.
type fakeVariantCatalog struct {
	mu       sync.Mutex
	variants map[uuid.UUID]*model.ProductVariant
}

func newFakeVariantCatalog() *fakeVariantCatalog {
	return &fakeVariantCatalog{variants: map[uuid.UUID]*model.ProductVariant{}}
}

func (f *fakeVariantCatalog) CreateVariant(_ context.Context, in *variant.CreateVariantInput) (*model.ProductVariant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v := &model.ProductVariant{
		ID:         uuid.New(),
		SKU:        uuid.New().String()[:8],
		ProductID:  in.ProductID,
		Title:      in.Title,
		Price:      in.Price,
		Attributes: in.Attributes,
		IsDefault:  in.IsDefault,
	}
	if in.SKU != nil {
		v.SKU = *in.SKU
	}
	f.variants[v.ID] = v
	return v, nil
}

func (f *fakeVariantCatalog) UpdateVariant(_ context.Context, variantID uuid.UUID, in variant.UpdateVariantInput) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.variants[variantID]
	if !ok {
		return database.ErrNotFound
	}
	if in.Title != nil {
		v.Title = *in.Title
	}
	return nil
}

func (f *fakeVariantCatalog) DeleteVariantByID(_ context.Context, variantID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.variants[variantID]; !ok {
		return database.ErrNotFound
	}
	delete(f.variants, variantID)
	return nil
}

func (f *fakeVariantCatalog) ListVariantsByProductID(_ context.Context, productID uuid.UUID, _ *pagination.ListQuery, _ bool) (*pagination.PagedResult[model.ProductVariant], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	items := make([]*model.ProductVariant, 0)
	for _, v := range f.variants {
		if v.ProductID == productID {
			items = append(items, v)
		}
	}
	return &pagination.PagedResult[model.ProductVariant]{Items: items}, nil
}

func (f *fakeVariantCatalog) GetVariantByID(_ context.Context, variantID uuid.UUID) (*model.ProductVariant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.variants[variantID]
	if !ok {
		return nil, database.ErrNotFound
	}
	return v, nil
}

// fakeStockAdjuster records the delegated stock movement.
type fakeStockAdjuster struct {
	mu      sync.Mutex
	adjusts []inventory.AdjustStockInput
}

func (f *fakeStockAdjuster) AdjustStock(_ context.Context, in inventory.AdjustStockInput) (*model.InventoryStock, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.adjusts = append(f.adjusts, in)
	return &model.InventoryStock{VariantID: in.VariantID, OnHandQuantity: in.Quantity, AvailableQuantity: in.Quantity, IsInStock: in.Quantity > 0}, nil
}

// fakeProductCatalog is an in-memory ProductCatalog double. Catalog rules are
// covered by the product domain; this fake exists to prove vendor ownership,
// merchant-standing, and delegation behavior.
type fakeProductCatalog struct {
	mu       sync.Mutex
	products map[uuid.UUID]*model.Product
}

func newFakeProductCatalog() *fakeProductCatalog {
	return &fakeProductCatalog{products: map[uuid.UUID]*model.Product{}}
}

func (f *fakeProductCatalog) CreateProductAsDraft(_ context.Context, input product.CreateProductInput) (*model.Product, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	slug := "product"
	if input.Slug != nil && *input.Slug != "" {
		slug = *input.Slug
	}
	now := time.Now()
	p := &model.Product{
		ID:          uuid.New(),
		VendorID:    input.VendorID,
		Slug:        slug,
		Title:       input.Title,
		Status:      model.PublicationStatusDraft,
		ProductType: model.ProductTypeSimple,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	f.products[p.ID] = p
	return p, nil
}

func (f *fakeProductCatalog) UpdateProduct(_ context.Context, productID uuid.UUID, input product.UpdateProductInput) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.products[productID]
	if !ok || p.DeletedAt != nil {
		return database.ErrNotFound
	}
	if input.Title != nil {
		p.Title = *input.Title
	}
	return nil
}

func (f *fakeProductCatalog) SoftDeleteProduct(_ context.Context, productID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.products[productID]; !ok {
		return database.ErrNotFound
	}
	delete(f.products, productID)
	return nil
}

func (f *fakeProductCatalog) PublishProduct(_ context.Context, productID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.products[productID]
	if !ok {
		return database.ErrNotFound
	}
	p.Status = model.PublicationStatusPublished
	return nil
}

func (f *fakeProductCatalog) ArchiveProduct(_ context.Context, productID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.products[productID]
	if !ok {
		return database.ErrNotFound
	}
	p.Status = model.PublicationStatusArchived
	return nil
}

func (f *fakeProductCatalog) GetByID(_ context.Context, productID uuid.UUID) (*model.Product, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.products[productID]
	if !ok {
		return nil, database.ErrNotFound
	}
	return p, nil
}

func (f *fakeProductCatalog) List(_ context.Context, q *product.ListProductsQuery, _ bool) (*pagination.PagedResult[product.ProductCardReadModel], error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	owned := map[uuid.UUID]bool{}
	for _, id := range q.VendorUUIDs {
		owned[id] = true
	}

	items := make([]*product.ProductCardReadModel, 0, len(f.products))
	for _, p := range f.products {
		if p.VendorID == nil || !owned[*p.VendorID] {
			continue
		}
		if !q.IncludeUnpublished && p.Status != model.PublicationStatusPublished {
			continue
		}
		items = append(items, &product.ProductCardReadModel{
			ID:          p.ID,
			Slug:        p.Slug,
			Title:       p.Title,
			Status:      p.Status,
			ProductType: p.ProductType,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	return &pagination.PagedResult[product.ProductCardReadModel]{Items: items}, nil
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

	s.catalog = newFakeProductCatalog()
	s.variants = newFakeVariantCatalog()
	s.stock = &fakeStockAdjuster{}

	s.service = vendors.NewService(
		vendors.NewRepository(),
		user.NewPostgresRepository(),
		object.NewRepository(),
		order.NewRepository(),
		s.catalog,
		s.variants,
		s.stock,
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

// ---- Merchant-owned catalog -------------------------------------------------

func (s *VendorServiceSuite) TestVendorCreatesOwnedDraftAndListsIt() {
	slug := "vendor-headphones"
	created, err := s.service.CreateProduct(s.ctx, s.vendorUID, product.CreateProductInput{
		CategoryID: s.categoryID,
		Slug:       &slug,
		Title:      "Vendor Headphones",
	})
	s.Require().NoError(err)
	s.Require().NotNil(created.VendorID)
	s.Equal(s.vendorID, *created.VendorID)
	s.Equal(model.PublicationStatusDraft, created.Status)

	page, err := s.service.ListOwnProducts(s.ctx, s.vendorUID, nil)
	s.Require().NoError(err)

	found := false
	for _, item := range page.Items {
		if item.ID == created.ID {
			found = true
		}
	}
	s.True(found, "the vendor's own draft must appear in their listing")
}

func (s *VendorServiceSuite) TestVendorListingExcludesForeignAndPlatformProducts() {
	foreignVendor := uuid.New()
	foreign := &model.Product{ID: uuid.New(), VendorID: &foreignVendor, Title: "Foreign"}
	platform := &model.Product{ID: uuid.New(), Title: "Platform"}
	s.catalog.mu.Lock()
	s.catalog.products[foreign.ID] = foreign
	s.catalog.products[platform.ID] = platform
	s.catalog.mu.Unlock()

	page, err := s.service.ListOwnProducts(s.ctx, s.vendorUID, nil)
	s.Require().NoError(err)
	for _, item := range page.Items {
		s.NotEqual(foreign.ID, item.ID)
		s.NotEqual(platform.ID, item.ID)
	}
}

func (s *VendorServiceSuite) TestVendorCannotMutateForeignProduct() {
	foreignVendor := uuid.New()
	foreign := &model.Product{ID: uuid.New(), VendorID: &foreignVendor, Title: "Foreign"}
	s.catalog.mu.Lock()
	s.catalog.products[foreign.ID] = foreign
	s.catalog.mu.Unlock()

	title := "Hijacked"
	err := s.service.UpdateProduct(s.ctx, s.vendorUID, foreign.ID, product.UpdateProductInput{Title: &title})
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	err = s.service.DeleteProduct(s.ctx, s.vendorUID, foreign.ID)
	apiErr = s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	s.Equal("Foreign", foreign.Title)
}

func (s *VendorServiceSuite) TestPlatformProductIsInvisibleToVendor() {
	platform := &model.Product{ID: uuid.New(), Title: "Platform"}
	s.catalog.mu.Lock()
	s.catalog.products[platform.ID] = platform
	s.catalog.mu.Unlock()

	err := s.service.PublishProduct(s.ctx, s.vendorUID, platform.ID)
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)
}

func (s *VendorServiceSuite) TestSuspendedVendorCannotCreateProduct() {
	_, err := s.service.AdminChangeVendorStatus(s.ctx, s.vendorID, model.VendorStatusSuspended)
	s.Require().NoError(err)

	slug := "suspended-listing"
	_, err = s.service.CreateProduct(s.ctx, s.vendorUID, product.CreateProductInput{
		CategoryID: s.categoryID,
		Slug:       &slug,
		Title:      "Suspended Listing",
	})
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusForbidden, apiErr.Status)

	_, err = s.service.AdminChangeVendorStatus(s.ctx, s.vendorID, model.VendorStatusActive)
	s.Require().NoError(err)
}

func (s *VendorServiceSuite) TestVendorCreatesVariantUnderOwnedProduct() {
	slug := "variant-owner-product"
	created, err := s.service.CreateProduct(s.ctx, s.vendorUID, product.CreateProductInput{
		CategoryID: s.categoryID,
		Slug:       &slug,
		Title:      "Variant Owner",
	})
	s.Require().NoError(err)

	price := int64(12500)
	v, err := s.service.CreateVariant(s.ctx, s.vendorUID, created.ID, &variant.CreateVariantInput{
		Title: "Black",
		Price: &price,
	})
	s.Require().NoError(err)
	s.Equal(created.ID, v.ProductID)

	page, err := s.service.ListProductVariants(s.ctx, s.vendorUID, created.ID, nil)
	s.Require().NoError(err)
	s.Len(page.Items, 1)

	stock, err := s.service.AdjustStock(s.ctx, s.vendorUID, inventory.AdjustStockInput{
		VariantID: v.ID,
		Quantity:  25,
		Reason:    "restock",
	})
	s.Require().NoError(err)
	s.Equal(25, stock.OnHandQuantity)
	s.Len(s.stock.adjusts, 1)
}

func (s *VendorServiceSuite) TestVendorCannotTouchForeignVariant() {
	foreignVendor := uuid.New()
	foreignProduct := &model.Product{ID: uuid.New(), VendorID: &foreignVendor, Title: "Foreign"}
	foreignVariant := &model.ProductVariant{ID: uuid.New(), ProductID: foreignProduct.ID, SKU: "F-1", Title: "Foreign Variant"}
	s.catalog.mu.Lock()
	s.catalog.products[foreignProduct.ID] = foreignProduct
	s.catalog.mu.Unlock()
	s.variants.mu.Lock()
	s.variants.variants[foreignVariant.ID] = foreignVariant
	s.variants.mu.Unlock()

	title := "Hijacked"
	err := s.service.UpdateVariant(s.ctx, s.vendorUID, foreignVariant.ID, variant.UpdateVariantInput{Title: &title})
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	err = s.service.DeleteVariant(s.ctx, s.vendorUID, foreignVariant.ID)
	apiErr = s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	_, err = s.service.AdjustStock(s.ctx, s.vendorUID, inventory.AdjustStockInput{
		VariantID: foreignVariant.ID,
		Quantity:  10,
		Reason:    "restock",
	})
	apiErr = s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	s.Equal("Foreign Variant", foreignVariant.Title)
	s.Empty(s.stock.adjusts)
}
