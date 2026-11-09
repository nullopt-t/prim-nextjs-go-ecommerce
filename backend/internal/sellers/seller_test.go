package sellers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nullopt-t/prim-backend/internal/catalog/category/attribute"
	"github.com/nullopt-t/prim-backend/internal/catalog/coupon"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/object"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/sellers"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// SellerServiceSuite is an integration suite for the seller domain's service
// logic, backed by real PostgreSQL via testcontainers (mirroring the checkout
// suite's scaffolding). Money-critical flows (lifecycle guards, balance
// invariant, conditional payout transitions) are covered here.
// fakeCouponCatalog is an in-memory CouponCatalog double for promotion tests.
type fakeCouponCatalog struct {
	mu      sync.Mutex
	coupons map[uuid.UUID]*model.Coupon
}

func newFakeCouponCatalog() *fakeCouponCatalog {
	return &fakeCouponCatalog{coupons: map[uuid.UUID]*model.Coupon{}}
}

func (f *fakeCouponCatalog) ListSellerCoupons(_ context.Context, sellerID uuid.UUID, _ int, _ *pagination.Keyset) ([]*model.Coupon, *pagination.CursorPage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var list []*model.Coupon
	for _, c := range f.coupons {
		if c.SellerID != nil && *c.SellerID == sellerID {
			list = append(list, c)
		}
	}
	return list, &pagination.CursorPage{}, nil
}

func (f *fakeCouponCatalog) CreateSellerCoupon(_ context.Context, sellerID uuid.UUID, in coupon.CreateCouponInput) (*model.Coupon, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := &model.Coupon{
		ID:            uuid.New(),
		Code:          in.Code,
		DiscountType:  in.DiscountType,
		DiscountValue: in.DiscountValue,
		Currency:      in.Currency,
		MinCartAmount: in.MinCartAmount,
		UsageLimit:    in.UsageLimit,
		PerUserLimit:  in.PerUserLimit,
		StartsAt:      in.StartsAt,
		ExpiresAt:     in.ExpiresAt,
		Status:        in.Status,
		SellerID:      &sellerID,
	}
	f.coupons[c.ID] = c
	return c, nil
}

func (f *fakeCouponCatalog) SetSellerCouponStatus(_ context.Context, sellerID, couponID uuid.UUID, status model.CouponStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.coupons[couponID]
	if !ok || c.SellerID == nil || *c.SellerID != sellerID {
		return database.ErrNotFound
	}
	c.Status = status
	return nil
}

func (f *fakeCouponCatalog) DeleteSellerCoupon(_ context.Context, sellerID, couponID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.coupons[couponID]
	if !ok || c.SellerID == nil || *c.SellerID != sellerID {
		return database.ErrNotFound
	}
	delete(f.coupons, couponID)
	return nil
}

// fakeSellerAttributeStore is an in-memory SellerAttributeStore double.
type fakeSellerAttributeStore struct {
	mu           sync.Mutex
	items        map[uuid.UUID]*attribute.AttributeItem
	platformDefs []attribute.AttributeItem
}

func newFakeSellerAttributeStore() *fakeSellerAttributeStore {
	return &fakeSellerAttributeStore{items: map[uuid.UUID]*attribute.AttributeItem{}}
}

func (f *fakeSellerAttributeStore) ListByCategory(_ context.Context, categoryID uuid.UUID) ([]attribute.AttributeItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var list []attribute.AttributeItem
	for _, def := range f.platformDefs {
		if def.CategoryID != nil && *def.CategoryID == categoryID {
			list = append(list, def)
		}
	}
	return list, nil
}

func (f *fakeSellerAttributeStore) ListBySellerCategory(_ context.Context, sellerID, categoryID uuid.UUID) ([]attribute.AttributeItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var list []attribute.AttributeItem
	for _, item := range f.items {
		if item.CategoryID != nil && *item.CategoryID == categoryID {
			list = append(list, *item)
		}
	}
	return list, nil
}

func (f *fakeSellerAttributeStore) CreateSellerAttribute(_ context.Context, sellerID uuid.UUID, ai *attribute.AttributeItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items[ai.ID] = ai
	return nil
}

func (f *fakeSellerAttributeStore) UpdateSellerAttribute(_ context.Context, sellerID, id uuid.UUID, fields attribute.UpdateAttributeFields) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.items[id]
	if !ok {
		return database.ErrNotFound
	}
	if fields.Label != nil {
		item.Label = *fields.Label
	}
	return nil
}

func (f *fakeSellerAttributeStore) DeleteSellerAttribute(_ context.Context, sellerID, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.items[id]; !ok {
		return database.ErrNotFound
	}
	delete(f.items, id)
	return nil
}

// fakeBundleStore is an in-memory BundleStore double.
type fakeBundleStore struct {
	mu      sync.Mutex
	bundles map[uuid.UUID]*sellers.SellerBundle
}

func newFakeBundleStore() *fakeBundleStore {
	return &fakeBundleStore{bundles: map[uuid.UUID]*sellers.SellerBundle{}}
}

func (f *fakeBundleStore) CreateBundle(_ context.Context, sellerID uuid.UUID, bundlePrice int64, currency string, productID uuid.UUID, items []sellers.BundleItem) (*sellers.SellerBundle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	bundle := &sellers.SellerBundle{
		ID:          uuid.New(),
		ProductID:   productID,
		BundlePrice: bundlePrice,
		Currency:    currency,
		Items:       items,
	}
	f.bundles[bundle.ID] = bundle
	return bundle, nil
}

func (f *fakeBundleStore) ListBundles(_ context.Context, sellerID uuid.UUID) ([]sellers.SellerBundle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var list []sellers.SellerBundle
	for _, bundle := range f.bundles {
		list = append(list, *bundle)
	}
	return list, nil
}

func (f *fakeBundleStore) UpdateBundle(_ context.Context, sellerID, bundleID uuid.UUID, bundlePrice *int64, currency *string, replaceItems bool, items []sellers.BundleItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	bundle, ok := f.bundles[bundleID]
	if !ok {
		return database.ErrNotFound
	}
	if bundlePrice != nil {
		bundle.BundlePrice = *bundlePrice
	}
	if replaceItems {
		bundle.Items = items
	}
	return nil
}

func (f *fakeBundleStore) IsBundleProduct(_ context.Context, productID uuid.UUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, bundle := range f.bundles {
		if bundle.ProductID == productID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeBundleStore) GetBundleProductID(_ context.Context, sellerID, bundleID uuid.UUID) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	bundle, ok := f.bundles[bundleID]
	if !ok {
		return uuid.Nil, database.ErrNotFound
	}
	return bundle.ProductID, nil
}

func (f *fakeBundleStore) DeleteBundle(_ context.Context, sellerID, bundleID uuid.UUID) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	bundle, ok := f.bundles[bundleID]
	if !ok {
		return uuid.Nil, database.ErrNotFound
	}
	delete(f.bundles, bundleID)
	return bundle.ProductID, nil
}

type SellerServiceSuite struct {
	suite.Suite

	ctx        context.Context
	pg         testcontainers.Container
	db         *database.DB
	service    *sellers.Service
	sellerID   uuid.UUID
	sellerUID  uuid.UUID
	categoryID uuid.UUID
	catalog    *fakeProductCatalog
	variants   *fakeVariantCatalog
	coupons    *fakeCouponCatalog
	attrs      *fakeSellerAttributeStore
	bundles    *fakeBundleStore
	stock      *fakeStockAdjuster
	runner     database.Runner
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
// covered by the product domain; this fake exists to prove seller ownership,
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
		SellerID:    input.SellerID,
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

func (f *fakeProductCatalog) GetSellerProductStats(_ context.Context, sellerID uuid.UUID, _ int) (*model.SellerProductStats, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	stats := new(model.SellerProductStats)
	for _, p := range f.products {
		if p.SellerID == nil || *p.SellerID != sellerID {
			continue
		}
		stats.Total++
		switch p.Status {
		case model.PublicationStatusPublished:
			stats.Published++
		case model.PublicationStatusDraft:
			stats.Draft++
		case model.PublicationStatusArchived:
			stats.Archived++
		}
	}
	return stats, nil
}

func (f *fakeProductCatalog) List(_ context.Context, q *product.ListProductsQuery, _ bool) (*pagination.PagedResult[product.ProductCardReadModel], error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	owned := map[uuid.UUID]bool{}
	for _, id := range q.SellerUUIDs {
		owned[id] = true
	}

	items := make([]*product.ProductCardReadModel, 0, len(f.products))
	for _, p := range f.products {
		if p.SellerID == nil || !owned[*p.SellerID] {
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

func TestSellerServiceSuite(t *testing.T) {
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon unavailable; skipping PostgreSQL integration suite")
	}
	suite.Run(t, new(SellerServiceSuite))
}

func (s *SellerServiceSuite) SetupSuite() {
	require := s.Require()
	s.ctx = context.Background()

	pg, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:18-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "testuser",
				"POSTGRES_PASSWORD": "testpass",
				"POSTGRES_DB":       "prim_seller_test",
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
		DBName:     "prim_seller_test",
	})
	require.NoError(err)

	for _, f := range []string{
		"../../migrations/000001_init.up.sql",
		"../../migrations/000005_order_fulfillment.up.sql",
	} {
		migration, err := os.ReadFile(f)
		require.NoError(err)
		_, err = s.db.ExecContext(s.ctx, string(migration))
		require.NoError(err)
	}

	logger := log.NewConsoleLogger()
	runner := database.NewTxRunner(s.db, logger)
	s.runner = runner

	// Seed one seller-facing user; the service registers their profile.
	db := s.db
	s.sellerUID = uuid.New()
	_, err = db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'seller@example.com', 'Seller User', 'seller', 'active', false, now(), now())
	`, s.sellerUID)
	require.NoError(err)

	s.catalog = newFakeProductCatalog()
	s.variants = newFakeVariantCatalog()
	s.coupons = newFakeCouponCatalog()
	s.attrs = newFakeSellerAttributeStore()
	s.bundles = newFakeBundleStore()
	s.stock = &fakeStockAdjuster{}

	s.service = sellers.NewService(
		sellers.NewRepository(),
		user.NewPostgresRepository(),
		object.NewService(runner, object.NewRepository(), storage.NoopProvider{}),
		order.NewRepository(),
		order.NewRepository(),
		s.catalog,
		s.variants,
		s.coupons,
		s.attrs,
		s.bundles,
		s.stock,
		sellers.NewStaticCommissionDefaults(),
		storage.NoopProvider{},
		runner,
		logger,
	)

	s.categoryID = uuid.New()
	_, err = db.ExecContext(s.ctx, `
		INSERT INTO product_categories (id, name) VALUES ($1, 'Payout Category')
	`, s.categoryID)
	require.NoError(err)

	registered, err := s.service.RegisterSeller(s.ctx, s.sellerUID, sellers.RegisterSellerInput{
		StoreName: "Suite Store",
		Slug:      "suite-store",
	})
	require.NoError(err)
	s.sellerID = registered.ID

	_, err = s.service.AdminChangeSellerStatus(s.ctx, s.sellerID, model.SellerStatusActive, "", s.sellerUID)
	require.NoError(err)
}

func (s *SellerServiceSuite) TearDownSuite() {
	s.db.Close()
	if s.pg != nil {
		_ = s.pg.Terminate(s.ctx)
	}
}

// seedSettledEarnings inserts a delivered order carrying one seller line with
// the supplied frozen seller_earning. Returns the order id.
func (s *SellerServiceSuite) seedSettledEarnings(earning int64) uuid.UUID {
	return s.seedSettledEarningsFor(s.sellerID, earning)
}

func (s *SellerServiceSuite) seedSettledEarningsFor(sellerID uuid.UUID, earning int64) uuid.UUID {
	require := s.Require()
	db := s.db
	ctx := s.ctx

	orderID := uuid.New()
	productID := uuid.New()
	variantID := uuid.New()

	_, err := db.ExecContext(ctx, `
		INSERT INTO products (id, category_id, slug, title, status, product_type, variation_axes, seller_id)
		VALUES ($1, $2, $3, 'Payout Product', 'published', 'simple', '{}', $4)
	`, productID, s.categoryID, "payout-product-"+orderID.String()[:8], sellerID)
	require.NoError(err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO product_variants (id, sku, product_id, title, price, currency)
		VALUES ($1, $2, $3, 'Payout Variant', 0, 'EGP')
	`, variantID, "SKU-"+orderID.String()[:8], productID)
	require.NoError(err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, customer_email, shipping_address, status, total_amount, currency)
		VALUES ($1, NULL, 'buyer@example.com', '{}'::jsonb, 'delivered', 0, 'EGP')
	`, orderID)
	require.NoError(err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO order_items (id, order_id, variant_id, quantity, price_at_purchase, product_snapshot, seller_id, seller_earning)
		VALUES ($1, $2, $3, 1, $5, '{}'::jsonb, $4, $5)
	`, uuid.New(), orderID, variantID, sellerID, earning)
	require.NoError(err)

	return orderID
}

// ---- Registration & lifecycle -----------------------------------------------

func (s *SellerServiceSuite) TestRegisterSellerCreatesPendingProfile() {
	uid := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'second@example.com', 'Second User', 'seller', 'active', false, now(), now())
	`, uid)
	s.Require().NoError(err)

	v, err := s.service.RegisterSeller(s.ctx, uid, sellers.RegisterSellerInput{
		StoreName: "Second Store",
		Slug:      "second-store",
	})
	s.Require().NoError(err)
	s.Equal(model.SellerStatusPending, v.Status)
	s.Equal(1000, v.CommissionRateBps)
}

func (s *SellerServiceSuite) TestRegisterSellerIsOnePerUser() {
	_, err := s.service.RegisterSeller(s.ctx, s.sellerUID, sellers.RegisterSellerInput{
		StoreName: "Suite Store Two",
		Slug:      "suite-store-two",
	})
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusConflict, apiErr.Status)
}

func (s *SellerServiceSuite) TestRegisterSellerSlugCollision() {
	uid := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'third@example.com', 'Third User', 'seller', 'active', false, now(), now())
	`, uid)
	s.Require().NoError(err)

	_, err = s.service.RegisterSeller(s.ctx, uid, sellers.RegisterSellerInput{
		StoreName: "Copycat Store",
		Slug:      "suite-store",
	})
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusConflict, apiErr.Status)
}

func (s *SellerServiceSuite) TestAdminLifecycleGuards() {
	// active -> pending is not a legal arrow.
	_, err := s.service.AdminChangeSellerStatus(s.ctx, s.sellerID, model.SellerStatusPending, "", s.sellerUID)
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusConflict, apiErr.Status)

	suspended, err := s.service.AdminChangeSellerStatus(s.ctx, s.sellerID, model.SellerStatusSuspended, "Policy review", s.sellerUID)
	s.Require().NoError(err)
	s.Equal(model.SellerStatusSuspended, suspended.Status)

	restored, err := s.service.AdminChangeSellerStatus(s.ctx, s.sellerID, model.SellerStatusActive, "", s.sellerUID)
	s.Require().NoError(err)
	s.Equal(model.SellerStatusActive, restored.Status)
}

// ---- Payouts ----------------------------------------------------------------

func (s *SellerServiceSuite) TestRequestPayoutRequiresPositiveAmount() {
	_, err := s.service.RequestPayout(s.ctx, s.sellerUID, 0, "")
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
}

func (s *SellerServiceSuite) TestRequestPayoutRejectsInsufficientBalance() {
	// Zero settled earnings: any request must be rejected.
	_, err := s.service.RequestPayout(s.ctx, s.sellerUID, 1, "")
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("PAYOUT_INSUFFICIENT_BALANCE", apiErr.Code)
}

func (s *SellerServiceSuite) TestRequestPayoutConsumesBalanceAndPaysOut() {
	// Adversarial gate #3: amount > settled - outstanding must 400. An exact
	// balance request succeeds and later claims exclude the assigned funds.
	earning := int64(100_000)
	s.seedSettledEarnings(earning)

	// Overshoot fails.
	_, err := s.service.RequestPayout(s.ctx, s.sellerUID, earning+1, "")
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("PAYOUT_INSUFFICIENT_BALANCE", apiErr.Code)

	// Exact balance succeeds and the snapshot matches.
	payout, err := s.service.RequestPayout(s.ctx, s.sellerUID, earning, "")
	s.Require().NoError(err)
	s.Equal(earning, payout.SnapshotAvailableBalance)
	s.Equal(model.PayoutStatusRequested, payout.Status)

	// This request now holds the funds: another exact request is rejected.
	_, err = s.service.RequestPayout(s.ctx, s.sellerUID, 1, "")
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

func (s *SellerServiceSuite) TestAdminTransitionPayoutConcurrentConflict() {
	// Two actors both attempt approved->processing; the first wins, the
	// second sees a conflict because the row state moved underneath (the
	// conditional update protects against the double transfer).
	earning := int64(5_000)
	s.seedSettledEarnings(earning)

	payout, err := s.service.RequestPayout(s.ctx, s.sellerUID, earning, "")
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

func (s *SellerServiceSuite) TestCustomerCanRegisterThroughHTTPRoutes() {
	uid := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'router-merchant@example.com', 'Router Merchant', 'customer', 'active', true, now(), now())
	`, uid)
	s.Require().NoError(err)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	sellers.NewRouter(sellers.NewHandler(s.service, stubMediaUploader{}), staticSessionStore{
		session: &model.UserSession{
			ID:       "router-session",
			UserID:   uid,
			UserRole: model.CustomerRole,
			Type:     model.SessionTypeAuthenticated,
		},
	}).MapRoutes(engine.Group("/api/v1"))

	cookie := &http.Cookie{Name: middleware.SessionCookieName, Value: "router-session"}
	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/seller/register",
		bytes.NewBufferString(`{"storeName":"HTTP Store","slug":"http-store"}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.AddCookie(cookie)
	createResponse := httptest.NewRecorder()
	engine.ServeHTTP(createResponse, createRequest)
	s.Require().Equal(http.StatusCreated, createResponse.Code, createResponse.Body.String())

	profileRequest := httptest.NewRequest(http.MethodGet, "/api/v1/seller/profile", nil)
	profileRequest.AddCookie(cookie)
	profileResponse := httptest.NewRecorder()
	engine.ServeHTTP(profileResponse, profileRequest)
	s.Require().Equal(http.StatusOK, profileResponse.Code, profileResponse.Body.String())

	var payload struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	s.Require().NoError(json.Unmarshal(profileResponse.Body.Bytes(), &payload))
	s.Equal(string(model.SellerStatusPending), payload.Data.Status)

	var role model.UserRole
	s.Require().NoError(s.db.QueryRowContext(s.ctx, `SELECT role FROM users WHERE id = $1`, uid).Scan(&role))
	s.Equal(model.CustomerRole, role)
}

func (s *SellerServiceSuite) TestCustomerRoleCanRegisterAndReadProfile() {
	uid := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'customer-merchant@example.com', 'Customer Merchant', 'customer', 'active', true, now(), now())
	`, uid)
	s.Require().NoError(err)

	created, err := s.service.RegisterSeller(s.ctx, uid, sellers.RegisterSellerInput{
		StoreName: "Customer Store",
		Slug:      "customer-store",
	})
	s.Require().NoError(err)
	s.Equal(model.SellerStatusPending, created.Status)

	profile, err := s.service.GetSellerByUser(s.ctx, uid)
	s.Require().NoError(err)
	s.Equal(created.ID, profile.ID)

	var role model.UserRole
	s.Require().NoError(s.db.QueryRowContext(s.ctx, `SELECT role FROM users WHERE id = $1`, uid).Scan(&role))
	s.Equal(model.CustomerRole, role, "seller capability must not mutate the staff role")
}

func (s *SellerServiceSuite) TestRejectedApplicationCanBeResubmitted() {
	uid, seller := s.createSeller(model.SellerStatusPending)

	rejected, err := s.service.AdminChangeSellerStatus(
		s.ctx, seller.ID, model.SellerStatusRejected, "Please provide a tax registration", s.sellerUID,
	)
	s.Require().NoError(err)
	s.Equal("Please provide a tax registration", rejected.StatusReason)
	s.Require().NotNil(rejected.ReviewedBy)
	s.Equal(s.sellerUID, *rejected.ReviewedBy)
	s.Require().NotNil(rejected.ReviewedAt)

	resubmitted, err := s.service.RegisterSeller(s.ctx, uid, sellers.RegisterSellerInput{
		StoreName:   "Resubmitted Store",
		Slug:        "resubmitted-store",
		Description: "Updated documents supplied",
	})
	s.Require().NoError(err)
	s.Equal(model.SellerStatusPending, resubmitted.Status)
	s.Empty(resubmitted.StatusReason)
	s.Nil(resubmitted.ReviewedBy)
	s.Nil(resubmitted.ReviewedAt)
	s.Equal("resubmitted-store", resubmitted.Slug)

	var role model.UserRole
	s.Require().NoError(s.db.QueryRowContext(s.ctx, `SELECT role FROM users WHERE id = $1`, uid).Scan(&role))
	s.Equal(model.CustomerRole, role)

	pending := model.SellerStatusPending
	queue, err := s.service.AdminListSellers(
		s.ctx, &pending, &pagination.ListQuery{Page: 1, PageSize: 100},
	)
	s.Require().NoError(err)
	found := false
	for _, item := range queue.Items {
		if item.ID == resubmitted.ID {
			found = true
			break
		}
	}
	s.True(found, "resubmitted application must reappear in the admin pending queue")
}

func (s *SellerServiceSuite) TestRejectedProfileEditReturnsToPending() {
	uid, seller := s.createSeller(model.SellerStatusPending)
	_, err := s.service.AdminChangeSellerStatus(
		s.ctx, seller.ID, model.SellerStatusRejected, "Store name is incomplete", s.sellerUID,
	)
	s.Require().NoError(err)

	name := "Completed Store Name"
	slug := "completed-store"
	updated, err := s.service.UpdateMyProfile(s.ctx, uid, sellers.UpdateSellerInput{
		StoreName: &name,
		Slug:      &slug,
	})
	s.Require().NoError(err)
	s.Equal(model.SellerStatusPending, updated.Status)
	s.Equal(name, updated.StoreName)
	s.Equal(slug, updated.Slug)
	s.Empty(updated.StatusReason)
}

func (s *SellerServiceSuite) TestSlugAvailabilityHonoursPolicyAndOwnership() {
	available, err := s.service.CheckSlugAvailability(s.ctx, s.sellerUID, "brand-new-store")
	s.Require().NoError(err)
	s.True(available.Available)

	own, err := s.service.CheckSlugAvailability(s.ctx, s.sellerUID, "suite-store")
	s.Require().NoError(err)
	s.True(own.Available, "a merchant's current slug is available for re-application")

	reserved, err := s.service.CheckSlugAvailability(s.ctx, s.sellerUID, "admin")
	s.Require().NoError(err)
	s.False(reserved.Available)
	s.Contains(reserved.Reason, "reserved")

	taken, err := s.service.CheckSlugAvailability(s.ctx, uuid.New(), "suite-store")
	s.Require().NoError(err)
	s.False(taken.Available)
	s.Contains(taken.Reason, "taken")

	short, err := s.service.CheckSlugAvailability(s.ctx, s.sellerUID, "ab")
	s.Require().NoError(err)
	s.False(short.Available)
}

func (s *SellerServiceSuite) TestSlugConflictMapsToFieldError() {
	_, existing := s.createSeller(model.SellerStatusPending)

	uid := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, $2, 'Slug Competitor', 'customer', 'active', true, now(), now())
	`, uid, "slug-competitor-"+uid.String()[:8]+"@example.com")
	s.Require().NoError(err)

	_, err = s.service.RegisterSeller(s.ctx, uid, sellers.RegisterSellerInput{
		StoreName: "Copy",
		Slug:      existing.Slug,
	})
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusConflict, apiErr.Status)
	s.Equal("SELLER_SLUG_TAKEN", apiErr.Code)
	s.Require().NotEmpty(apiErr.Fields)
	s.Equal("slug", apiErr.Fields[0].Field)
}

func (s *SellerServiceSuite) TestStatusMatrixGuardsOperationalSurfaces() {
	uid, _ := s.createSeller(model.SellerStatusPending)
	q := &pagination.ListQuery{Page: 1, PageSize: 20}

	_, err := s.service.GetDashboard(s.ctx, uid)
	s.assertStatus(err, http.StatusForbidden)
	_, err = s.service.ListOwnProducts(s.ctx, uid, nil)
	s.assertStatus(err, http.StatusForbidden)
	_, err = s.service.ListOwnOrders(s.ctx, uid, nil, q)
	s.assertStatus(err, http.StatusForbidden)
	_, err = s.service.ListMyPayouts(s.ctx, uid)
	s.assertStatus(err, http.StatusForbidden)

	active, err := s.service.AdminChangeSellerStatus(s.ctx, s.sellerByUser(uid).ID, model.SellerStatusActive, "", s.sellerUID)
	s.Require().NoError(err)
	s.Equal(model.SellerStatusActive, active.Status)
	suspended, err := s.service.AdminChangeSellerStatus(
		s.ctx, active.ID, model.SellerStatusSuspended, "compliance hold", s.sellerUID,
	)
	s.Require().NoError(err)
	s.Equal(model.SellerStatusSuspended, suspended.Status)

	_, err = s.service.GetDashboard(s.ctx, uid)
	s.assertStatus(err, http.StatusForbidden)
	_, err = s.service.ListOwnProducts(s.ctx, uid, nil)
	s.assertStatus(err, http.StatusForbidden)
	_, err = s.service.ListOwnOrders(s.ctx, uid, nil, q)
	s.NoError(err)
	_, err = s.service.ListMyPayouts(s.ctx, uid)
	s.NoError(err)
}

func (s *SellerServiceSuite) TestPayoutTransitionsAreHeldWhileSuspended() {
	uid, seller := s.createSeller(model.SellerStatusActive)
	s.seedSettledEarningsFor(seller.ID, 10_000)

	payout, err := s.service.RequestPayout(s.ctx, uid, 5_000, "")
	s.Require().NoError(err)

	_, err = s.service.AdminChangeSellerStatus(
		s.ctx, seller.ID, model.SellerStatusSuspended, "payout hold test", s.sellerUID,
	)
	s.Require().NoError(err)

	_, err = s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusApproved, nil, nil)
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusConflict, apiErr.Status)
	s.Equal("SELLER_NOT_AUTHORIZED_FOR_ACTION", apiErr.Code)

	_, err = s.service.AdminChangeSellerStatus(s.ctx, seller.ID, model.SellerStatusActive, "", s.sellerUID)
	s.Require().NoError(err)
	approved, err := s.service.AdminTransitionPayout(s.ctx, payout.ID, model.PayoutStatusApproved, nil, nil)
	s.Require().NoError(err)
	s.Equal(model.PayoutStatusApproved, approved.Status)
}

func (s *SellerServiceSuite) TestPayoutIdempotencyKeyReplaysSameRequest() {
	uid, seller := s.createSeller(model.SellerStatusActive)
	s.seedSettledEarningsFor(seller.ID, 10_000)

	first, err := s.service.RequestPayout(s.ctx, uid, 2_500, "withdraw-1")
	s.Require().NoError(err)

	second, err := s.service.RequestPayout(s.ctx, uid, 2_500, "withdraw-1")
	s.Require().NoError(err)
	s.Equal(first.ID, second.ID)

	_, err = s.service.RequestPayout(s.ctx, uid, 3_000, "withdraw-1")
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusConflict, apiErr.Status)
	s.Equal("PAYOUT_IDEMPOTENCY_CONFLICT", apiErr.Code)
}

func (s *SellerServiceSuite) createSeller(status model.SellerStatus) (uuid.UUID, *model.Seller) {
	s.Require().NotNil(s.service)
	uid := uuid.New()
	email := "merchant-" + uid.String()[:8] + "@example.com"
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, $2, 'Lifecycle Merchant', 'customer', 'active', true, now(), now())
	`, uid, email)
	s.Require().NoError(err)

	seller, err := s.service.RegisterSeller(s.ctx, uid, sellers.RegisterSellerInput{
		StoreName: "Lifecycle Store " + uid.String()[:8],
		Slug:      "lifecycle-" + uid.String()[:8],
	})
	s.Require().NoError(err)

	switch status {
	case model.SellerStatusPending:
	case model.SellerStatusRejected:
		_, err = s.service.AdminChangeSellerStatus(s.ctx, seller.ID, status, "test rejection", s.sellerUID)
	case model.SellerStatusActive:
		_, err = s.service.AdminChangeSellerStatus(s.ctx, seller.ID, status, "", s.sellerUID)
	case model.SellerStatusSuspended:
		_, err = s.service.AdminChangeSellerStatus(s.ctx, seller.ID, model.SellerStatusActive, "", s.sellerUID)
		if err == nil {
			_, err = s.service.AdminChangeSellerStatus(s.ctx, seller.ID, status, "test suspension", s.sellerUID)
		}
	}
	s.Require().NoError(err)
	loaded, err := s.service.GetSellerByUser(s.ctx, uid)
	s.Require().NoError(err)
	return uid, loaded
}

func (s *SellerServiceSuite) sellerByUser(userID uuid.UUID) *model.Seller {
	seller, err := s.service.GetSellerByUser(s.ctx, userID)
	s.Require().NoError(err)
	return seller
}

func (s *SellerServiceSuite) assertStatus(err error, status int) {
	apiErr := s.assertAPIError(err)
	s.Equal(status, apiErr.Status)
}

type staticSessionStore struct {
	session *model.UserSession
}

func (s staticSessionStore) GetSessionByID(context.Context, string) (*model.UserSession, error) {
	return s.session, nil
}

// stubMediaUploader satisfies sellers.MediaUploader for router tests that do
// not exercise uploads.
type stubMediaUploader struct{}

func (stubMediaUploader) Upload(*gin.Context) (*object.MediaItem, error) { return nil, nil }

func (s *SellerServiceSuite) assertAPIError(err error) *apierr.APIError {
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	return apiErr
}

// seedThumbnailURL inserts a storage object row and returns the public URL
// the Noop storage provider computes for it.
func (s *SellerServiceSuite) seedThumbnailURL() string {
	key := uuid.New().String()
	obj := &model.Object{
		ID:          uuid.New(),
		ContentType: "image/png",
		FileSize:    1024,
		Status:      model.ObjectStatusUploaded,
		Key:         key,
		Bucket:      "catalog",
	}
	err := s.runner.WithTx(s.ctx, func(ctx context.Context) error {
		return object.NewRepository().Create(ctx, obj)
	})
	s.Require().NoError(err)
	return storage.NoopProvider{}.GetPublicURL(obj.Bucket, obj.Key)
}

// ---- Merchant-owned catalog -------------------------------------------------

func (s *SellerServiceSuite) TestSellerCreatesOwnedDraftAndListsIt() {
	slug := "seller-headphones"
	created, err := s.service.CreateProduct(s.ctx, s.sellerUID, s.seedThumbnailURL(), product.CreateProductInput{
		CategoryID: s.categoryID,
		Slug:       &slug,
		Title:      "Seller Headphones",
	})
	s.Require().NoError(err)
	s.Require().NotNil(created.SellerID)
	s.Equal(s.sellerID, *created.SellerID)
	s.Equal(model.PublicationStatusDraft, created.Status)

	page, err := s.service.ListOwnProducts(s.ctx, s.sellerUID, nil)
	s.Require().NoError(err)

	found := false
	for _, item := range page.Items {
		if item.ID == created.ID {
			found = true
		}
	}
	s.True(found, "the seller's own draft must appear in their listing")
}

func (s *SellerServiceSuite) TestSellerListingExcludesForeignAndPlatformProducts() {
	foreignSeller := uuid.New()
	foreign := &model.Product{ID: uuid.New(), SellerID: &foreignSeller, Title: "Foreign"}
	platform := &model.Product{ID: uuid.New(), Title: "Platform"}
	s.catalog.mu.Lock()
	s.catalog.products[foreign.ID] = foreign
	s.catalog.products[platform.ID] = platform
	s.catalog.mu.Unlock()

	page, err := s.service.ListOwnProducts(s.ctx, s.sellerUID, nil)
	s.Require().NoError(err)
	for _, item := range page.Items {
		s.NotEqual(foreign.ID, item.ID)
		s.NotEqual(platform.ID, item.ID)
	}
}

func (s *SellerServiceSuite) TestSellerCannotMutateForeignProduct() {
	foreignSeller := uuid.New()
	foreign := &model.Product{ID: uuid.New(), SellerID: &foreignSeller, Title: "Foreign"}
	s.catalog.mu.Lock()
	s.catalog.products[foreign.ID] = foreign
	s.catalog.mu.Unlock()

	title := "Hijacked"
	err := s.service.UpdateProduct(s.ctx, s.sellerUID, foreign.ID, product.UpdateProductInput{Title: &title})
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	err = s.service.DeleteProduct(s.ctx, s.sellerUID, foreign.ID)
	apiErr = s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	s.Equal("Foreign", foreign.Title)
}

func (s *SellerServiceSuite) TestPlatformProductIsInvisibleToSeller() {
	platform := &model.Product{ID: uuid.New(), Title: "Platform"}
	s.catalog.mu.Lock()
	s.catalog.products[platform.ID] = platform
	s.catalog.mu.Unlock()

	err := s.service.PublishProduct(s.ctx, s.sellerUID, platform.ID)
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)
}

func (s *SellerServiceSuite) TestSuspendedSellerCannotCreateProduct() {
	_, err := s.service.AdminChangeSellerStatus(s.ctx, s.sellerID, model.SellerStatusSuspended, "Policy review", s.sellerUID)
	s.Require().NoError(err)

	slug := "suspended-listing"
	_, err = s.service.CreateProduct(s.ctx, s.sellerUID, s.seedThumbnailURL(), product.CreateProductInput{
		CategoryID: s.categoryID,
		Slug:       &slug,
		Title:      "Suspended Listing",
	})
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusForbidden, apiErr.Status)

	_, err = s.service.AdminChangeSellerStatus(s.ctx, s.sellerID, model.SellerStatusActive, "", s.sellerUID)
	s.Require().NoError(err)
}

func (s *SellerServiceSuite) TestSellerCreatesVariantUnderOwnedProduct() {
	slug := "variant-owner-product"
	created, err := s.service.CreateProduct(s.ctx, s.sellerUID, s.seedThumbnailURL(), product.CreateProductInput{
		CategoryID: s.categoryID,
		Slug:       &slug,
		Title:      "Variant Owner",
	})
	s.Require().NoError(err)

	price := int64(12500)
	v, err := s.service.CreateVariant(s.ctx, s.sellerUID, created.ID, "", &variant.CreateVariantInput{
		Title: "Black",
		Price: &price,
	})
	s.Require().NoError(err)
	s.Equal(created.ID, v.ProductID)

	page, err := s.service.ListProductVariants(s.ctx, s.sellerUID, created.ID, nil)
	s.Require().NoError(err)
	s.Len(page.Items, 1)

	stock, err := s.service.AdjustStock(s.ctx, s.sellerUID, inventory.AdjustStockInput{
		VariantID: v.ID,
		Quantity:  25,
		Reason:    "restock",
	})
	s.Require().NoError(err)
	s.Equal(25, stock.OnHandQuantity)
	s.Len(s.stock.adjusts, 1)
}

func (s *SellerServiceSuite) TestSellerCannotTouchForeignVariant() {
	foreignSeller := uuid.New()
	foreignProduct := &model.Product{ID: uuid.New(), SellerID: &foreignSeller, Title: "Foreign"}
	foreignVariant := &model.ProductVariant{ID: uuid.New(), ProductID: foreignProduct.ID, SKU: "F-1", Title: "Foreign Variant"}
	s.catalog.mu.Lock()
	s.catalog.products[foreignProduct.ID] = foreignProduct
	s.catalog.mu.Unlock()
	s.variants.mu.Lock()
	s.variants.variants[foreignVariant.ID] = foreignVariant
	s.variants.mu.Unlock()

	title := "Hijacked"
	err := s.service.UpdateVariant(s.ctx, s.sellerUID, foreignVariant.ID, nil, variant.UpdateVariantInput{Title: &title})
	apiErr := s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	err = s.service.DeleteVariant(s.ctx, s.sellerUID, foreignVariant.ID)
	apiErr = s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	_, err = s.service.AdjustStock(s.ctx, s.sellerUID, inventory.AdjustStockInput{
		VariantID: foreignVariant.ID,
		Quantity:  10,
		Reason:    "restock",
	})
	apiErr = s.assertAPIError(err)
	s.Equal(http.StatusNotFound, apiErr.Status)

	s.Equal("Foreign Variant", foreignVariant.Title)
	s.Empty(s.stock.adjusts)
}

func (s *SellerServiceSuite) TestDashboardReportsMoneyAndCatalog() {
	// A dedicated seller keeps the counters independent of other tests.
	uid := uuid.New()
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO users (id, email, full_name, role, status, is_email_verified, created_at, updated_at)
		VALUES ($1, 'dash@example.com', 'Dash Seller', 'seller', 'active', false, now(), now())
	`, uid)
	s.Require().NoError(err)

	seller, err := s.service.RegisterSeller(s.ctx, uid, sellers.RegisterSellerInput{
		StoreName: "Dash Store",
		Slug:      "dash-store",
	})
	s.Require().NoError(err)
	_, err = s.service.AdminChangeSellerStatus(s.ctx, seller.ID, model.SellerStatusActive, "", s.sellerUID)
	s.Require().NoError(err)

	for _, slug := range []string{"dash-a", "dash-b"} {
		slug := slug
		_, err := s.service.CreateProduct(s.ctx, uid, s.seedThumbnailURL(), product.CreateProductInput{
			CategoryID: s.categoryID,
			Slug:       &slug,
			Title:      "Dash Product " + slug,
		})
		s.Require().NoError(err)
	}

	s.seedSettledEarningsFor(seller.ID, 2500)

	dashboard, err := s.service.GetDashboard(s.ctx, uid)
	s.Require().NoError(err)
	s.Equal(int64(2500), dashboard.LifetimeEarnings)
	s.Equal(int64(2500), dashboard.AvailableBalance)
	s.Equal(int64(0), dashboard.PendingEarnings)
	s.Equal(int64(2500), dashboard.GrossRevenue)
	s.Equal(1, dashboard.OrderCount)
	s.Equal(2, dashboard.ProductCount)
	s.Equal(2, dashboard.DraftProductCount)
	s.Equal(0, dashboard.PublishedProductCount)

	_, err = s.service.RequestPayout(s.ctx, uid, 1000, "")
	s.Require().NoError(err)

	dashboard, err = s.service.GetDashboard(s.ctx, uid)
	s.Require().NoError(err)
	s.Equal(int64(1000), dashboard.OutstandingPayouts)
	s.Equal(int64(1500), dashboard.AvailableBalance)
}
