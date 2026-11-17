package product

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

	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

func TestParseSuggestionParams(t *testing.T) {
	tests := []struct {
		name      string
		q         string
		limitRaw  string
		wantQuery string
		wantLimit int
	}{
		{"trims query and defaults limit", "  mouse  ", "", "mouse", 6},
		{"caps limit at ten", "mouse", "50", "mouse", 10},
		{"keeps valid limit", "mouse", "3", "mouse", 3},
		{"ignores invalid limit", "mouse", "abc", "mouse", 6},
		{"ignores non-positive limit", "mouse", "0", "mouse", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, limit := parseSuggestionParams(tt.q, tt.limitRaw)
			require.Equal(t, tt.wantQuery, query)
			require.Equal(t, tt.wantLimit, limit)
		})
	}
}

func TestResolveSuggestionThumbnails(t *testing.T) {
	items := []*ProductSuggestionResponse{
		{ID: "with-image", thumbnailBucket: "catalog", thumbnailKey: "img/a.jpg"},
		{ID: "no-image"},
	}

	resolveSuggestionThumbnails(items, storage.NoopProvider{})

	require.Equal(t, "http://localhost/catalog/img/a.jpg", items[0].Thumbnail)
	require.Empty(t, items[1].Thumbnail)
}

type ProductSuggestionsSuite struct {
	suite.Suite
	pgContainer     *postgres.PostgresContainer
	db              *database.DB
	runner          database.Runner
	repo            *ProductRepository
	categoryID      uuid.UUID
	otherCategoryID uuid.UUID
	brandID         uuid.UUID
	tagID           uuid.UUID
	objectID        uuid.UUID
}

func (s *ProductSuggestionsSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_product_suggestions_test"),
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
		DBUser: "testuser", DBPassword: "testpass", DBName: "prim_product_suggestions_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	for _, migration := range []string{
		"../../../migrations/000001_init.up.sql",
		"../../../migrations/000003_pg_trgm_search.up.sql",
	} {
		schema, readErr := os.ReadFile(migration)
		require.NoError(s.T(), readErr)
		_, execErr := db.ExecContext(ctx, string(schema))
		require.NoError(s.T(), execErr)
	}

	s.runner = database.NewTxRunner(db, log.NewNoopLogger())
	s.repo = NewProductRepository()
	s.categoryID = uuid.New()
	s.otherCategoryID = uuid.New()
	s.brandID = uuid.New()
	s.tagID = uuid.New()
	s.objectID = uuid.New()

	_, err = db.ExecContext(ctx, `INSERT INTO product_categories (id, name) VALUES ($1, $2)`, s.categoryID, "Mouse Pads")
	require.NoError(s.T(), err)
	_, err = db.ExecContext(ctx, `INSERT INTO product_categories (id, name) VALUES ($1, $2)`, s.otherCategoryID, "Other Gear")
	require.NoError(s.T(), err)
	_, err = db.ExecContext(ctx, `INSERT INTO product_brands (id, name) VALUES ($1, $2)`, s.brandID, "Mouse Brand")
	require.NoError(s.T(), err)
	_, err = db.ExecContext(ctx, `INSERT INTO product_tags (id, name) VALUES ($1, $2)`, s.tagID, "Mouse Tag")
	require.NoError(s.T(), err)
	_, err = db.ExecContext(ctx, `INSERT INTO storage_objects (id, bucket, object_key, content_type, file_size, status) VALUES ($1, 'catalog', 'img/a.jpg', 'image/jpeg', 10, 'uploaded')`, s.objectID)
	require.NoError(s.T(), err)
}

func (s *ProductSuggestionsSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *ProductSuggestionsSuite) seedProduct(slug, title string, status string, brandID, categoryID *uuid.UUID, thumbnailID *uuid.UUID, age time.Duration, deleted bool) uuid.UUID {
	id := uuid.New()
	var deletedAt any
	if deleted {
		deletedAt = time.Now()
	}
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO products (id, slug, title, status, product_type, variation_axes, brand_id, category_id, thumbnail_object_id, created_at, deleted_at)
		VALUES ($1, $2, $3, $4::product_status, 'simple', '{}'::text[], $5, $6, $7, now() - $8::interval, $9)
	`, id, slug, title, status, brandID, categoryID, thumbnailID, age, deletedAt)
	s.Require().NoError(err)
	return id
}

func (s *ProductSuggestionsSuite) TestGetSuggestionsMatchingRankingAndThumbnail() {
	ctx := context.Background()

	prefix := s.seedProduct("mouse-pro", "Mouse Pro", "published", nil, &s.otherCategoryID, &s.objectID, time.Hour, false)
	title := s.seedProduct("wireless-mouse", "Wireless Mouse", "published", nil, &s.otherCategoryID, nil, 0, false)
	slug := s.seedProduct("mouse-cable", "Generic Thing", "published", nil, &s.otherCategoryID, nil, 0, false)
	brand := s.seedProduct("branded-thing", "Branded Thing", "published", &s.brandID, &s.otherCategoryID, nil, 0, false)
	cat := s.seedProduct("blue-pad", "Blue Pad", "published", nil, &s.categoryID, nil, 0, false)
	tagged := s.seedProduct("tagged-thing", "Tagged Thing", "published", nil, &s.otherCategoryID, nil, 0, false)
	draft := s.seedProduct("mouse-draft", "Mouse Draft", "draft", nil, &s.otherCategoryID, nil, 0, false)
	deleted := s.seedProduct("mouse-deleted", "Mouse Deleted", "published", nil, &s.otherCategoryID, nil, 0, true)

	_, err := s.db.ExecContext(ctx, `INSERT INTO product_tag_assignments (product_id, tag_id) VALUES ($1, $2)`, tagged, s.tagID)
	s.Require().NoError(err)

	sku := s.seedProduct("sku-thing", "SKU Thing", "published", nil, &s.otherCategoryID, nil, 0, false)
	_, err = s.db.ExecContext(ctx, `INSERT INTO product_variants (id, sku, product_id, title) VALUES ($1, 'MOUSE-SKU-1', $2, 'Default')`, uuid.New(), sku)
	s.Require().NoError(err)

	deletedSku := s.seedProduct("deleted-sku", "Deleted SKU Thing", "published", nil, &s.otherCategoryID, nil, 0, false)
	_, err = s.db.ExecContext(ctx, `INSERT INTO product_variants (id, sku, product_id, title, deleted_at) VALUES ($1, 'MOUSE-DELETED', $2, 'Default', now())`, uuid.New(), deletedSku)
	s.Require().NoError(err)

	userID := uuid.New()
	_, err = s.db.ExecContext(ctx, `INSERT INTO users (id, phone, full_name, role, phone_verified, status, country_code) VALUES ($1, $2, 'Seller', 'seller', true, 'active', '+20')`, userID, "seller_"+userID.String()+"@example.com")
	s.Require().NoError(err)
	sellerID := uuid.New()
	_, err = s.db.ExecContext(ctx, `INSERT INTO sellers (id, user_id, store_name, slug, status) VALUES ($1, $2, $3, $4, 'suspended')`, sellerID, userID, "Suspended Store", "suspended-store")
	s.Require().NoError(err)
	suspended := s.seedProduct("mouse-suspended", "Mouse Suspended", "published", nil, &s.otherCategoryID, nil, 0, false)
	_, err = s.db.ExecContext(ctx, `UPDATE products SET seller_id = $1 WHERE id = $2`, sellerID, suspended)
	s.Require().NoError(err)

	var results []*ProductSuggestionResponse
	err = s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		var getErr error
		results, getErr = s.repo.GetSuggestions(ctx, "mouse", 10)
		return getErr
	})
	s.Require().NoError(err)

	ids := make([]string, 0, len(results))
	for _, r := range results {
		ids = append(ids, r.ID)
	}
	s.Require().Equal(prefix.String(), ids[0], "exact title prefix ranks first")
	s.Require().Len(ids, 7)
	s.Require().ElementsMatch(
		[]string{prefix.String(), title.String(), slug.String(), brand.String(), cat.String(), tagged.String(), sku.String()},
		ids,
		"title, slug, brand, category, tag and SKU matches are included",
	)
	s.Require().NotContains(ids, draft.String(), "draft products are hidden")
	s.Require().NotContains(ids, deleted.String(), "soft-deleted products are hidden")
	s.Require().NotContains(ids, deletedSku.String(), "soft-deleted variants cannot match by SKU")
	s.Require().NotContains(ids, suspended.String(), "suspended sellers' products are hidden from storefront suggestions")

	for _, r := range results {
		if r.ID == prefix.String() {
			s.Require().Equal("catalog", r.thumbnailBucket)
			s.Require().Equal("img/a.jpg", r.thumbnailKey)
		}
	}

	resolveSuggestionThumbnails(results, storage.NoopProvider{})
	for _, r := range results {
		if r.ID == prefix.String() {
			s.Require().Equal("http://localhost/catalog/img/a.jpg", r.Thumbnail)
		}
	}
}

func (s *ProductSuggestionsSuite) TestGetSuggestionsRespectsLimit() {
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		s.seedProduct("limited-"+string(rune('a'+i)), "Limited "+string(rune('a'+i)), "published", nil, &s.categoryID, nil, 0, false)
	}

	var results []*ProductSuggestionResponse
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		var getErr error
		results, getErr = s.repo.GetSuggestions(ctx, "Limited", 2)
		return getErr
	})
	s.Require().NoError(err)
	s.Require().Len(results, 2)
}

func TestProductSuggestionsSuite(t *testing.T) {
	suite.Run(t, new(ProductSuggestionsSuite))
}
