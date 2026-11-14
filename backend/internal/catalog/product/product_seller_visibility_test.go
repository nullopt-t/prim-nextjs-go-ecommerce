package product_test

// Storefront visibility for seller-owned products. A suspended or rejected
// seller's published products must not be listable or reachable by slug, while
// admin listings (IncludeUnpublished) still see them.

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type ProductSellerVisibilitySuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *database.DB
	runner      database.Runner
	repo        *product.ProductRepository
	categoryID  uuid.UUID
}

func (s *ProductSellerVisibilitySuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_product_visibility_test"),
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
		DBUser: "testuser", DBPassword: "testpass", DBName: "prim_product_visibility_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	for _, migration := range []string{
		"../../../migrations/000001_init.up.sql",
	} {
		schema, readErr := os.ReadFile(migration)
		require.NoError(s.T(), readErr)
		_, execErr := db.ExecContext(ctx, string(schema))
		require.NoError(s.T(), execErr)
	}

	s.runner = database.NewTxRunner(s.db, log.NewNoopLogger())
	s.repo = product.NewProductRepository()
	s.categoryID = uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO product_categories (id, name) VALUES ($1, $2)`, s.categoryID, "Visibility")
	require.NoError(s.T(), err)
}

func (s *ProductSellerVisibilitySuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *ProductSellerVisibilitySuite) seedProduct(slug string, sellerID *uuid.UUID) uuid.UUID {
	id := uuid.New()
	var seller any
	if sellerID != nil {
		seller = *sellerID
	}
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO products (id, category_id, slug, title, status, product_type, variation_axes, seller_id)
		VALUES ($1, $2, $3, $4, 'published', 'variable', ARRAY['color']::text[], $5)
	`, id, s.categoryID, slug, "Product "+slug, seller)
	s.Require().NoError(err)
	return id
}

func (s *ProductSellerVisibilitySuite) seedSeller(status model.SellerStatus) uuid.UUID {
	userID := uuid.New()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO users (id, phone, full_name, role, phone_verified, status, country_code)
		VALUES ($1, $2, 'Seller User', 'seller', true, 'active', '+20')
	`, userID, "seller_"+userID.String()+"@example.com")
	s.Require().NoError(err)

	sellerID := uuid.New()
	_, err = s.db.ExecContext(context.Background(), `
		INSERT INTO sellers (id, user_id, store_name, slug, status)
		VALUES ($1, $2, $3, $4, $5)
	`, sellerID, userID, "Store "+sellerID.String()[:8], "store-"+sellerID.String()[:8], string(status))
	s.Require().NoError(err)
	return sellerID
}

func (s *ProductSellerVisibilitySuite) TestStorefrontHidesSuspendedSellerProducts() {
	activeSeller := s.seedSeller(model.SellerStatusActive)
	suspendedSeller := s.seedSeller(model.SellerStatusSuspended)
	rejectedSeller := s.seedSeller(model.SellerStatusRejected)

	platformSlug := "platform-" + uuid.NewString()
	activeSlug := "active-" + uuid.NewString()
	suspendedSlug := "suspended-" + uuid.NewString()
	rejectedSlug := "rejected-" + uuid.NewString()

	s.seedProduct(platformSlug, nil)
	s.seedProduct(activeSlug, &activeSeller)
	s.seedProduct(suspendedSlug, &suspendedSeller)
	s.seedProduct(rejectedSlug, &rejectedSeller)

	var list *pagination.PagedResult[product.ProductCardReadModel]
	err := s.runner.WithExecutor(context.Background(), func(ctx context.Context) error {
		var listErr error
		list, listErr = s.repo.List(ctx, &product.ListProductsQuery{}, false)
		return listErr
	})
	s.Require().NoError(err)

	slugs := make(map[string]bool)
	for _, item := range list.Items {
		slugs[item.Slug] = true
	}
	s.True(slugs[platformSlug], "platform products stay visible")
	s.True(slugs[activeSlug], "active seller products stay visible")
	s.False(slugs[suspendedSlug], "suspended seller products must be hidden")
	s.False(slugs[rejectedSlug], "rejected seller products must be hidden")

	// Public detail lookups must agree with the listing.
	err = s.runner.WithExecutor(context.Background(), func(ctx context.Context) error {
		_, getErr := s.repo.GetBySlug(ctx, suspendedSlug)
		return getErr
	})
	s.Require().Error(err)
	err = s.runner.WithExecutor(context.Background(), func(ctx context.Context) error {
		_, getErr := s.repo.GetBySlug(ctx, activeSlug)
		return getErr
	})
	s.Require().NoError(err)

	// Admin listings still include every published product.
	var adminList *pagination.PagedResult[product.ProductCardReadModel]
	err = s.runner.WithExecutor(context.Background(), func(ctx context.Context) error {
		var listErr error
		adminList, listErr = s.repo.List(ctx, &product.ListProductsQuery{IncludeUnpublished: true}, false)
		return listErr
	})
	s.Require().NoError(err)
	adminSlugs := make(map[string]bool)
	for _, item := range adminList.Items {
		adminSlugs[item.Slug] = true
	}
	s.True(adminSlugs[suspendedSlug])
	s.True(adminSlugs[rejectedSlug])
}

func TestProductSellerVisibilitySuite(t *testing.T) {
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon unavailable; skipping PostgreSQL integration suite")
	}
	suite.Run(t, new(ProductSellerVisibilitySuite))
}
