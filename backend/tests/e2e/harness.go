// Package e2e provides an in-memory testing harness and opaque-box test suites
// for validating domain service decoupling, single-connection execution, error mapping,
// direct storage resolution, and multi-domain workflows offline.
package e2e

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// =============================================================================
// In-Memory Database & Runner Test Doubles
// =============================================================================

// MockQueryExecutor implements database.QueryExecutor for in-memory testing.
type MockQueryExecutor struct {
	mu          sync.Mutex
	queries     []string
	execErr     error
	queryErr    error
	queryRowErr error
	rowScanFn   func(dest ...any) error
}

func NewMockQueryExecutor() *MockQueryExecutor {
	return &MockQueryExecutor{
		queries: make([]string, 0),
	}
}

func (m *MockQueryExecutor) ExecContext(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	m.mu.Lock()
	m.queries = append(m.queries, query)
	err := m.execErr
	m.mu.Unlock()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return pgconn.NewCommandTag("EXEC 1"), nil
}

func (m *MockQueryExecutor) QueryContext(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	m.mu.Lock()
	m.queries = append(m.queries, query)
	err := m.queryErr
	m.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return &mockRows{}, nil
}

func (m *MockQueryExecutor) QueryRowContext(ctx context.Context, query string, args ...any) pgx.Row {
	m.mu.Lock()
	m.queries = append(m.queries, query)
	err := m.queryRowErr
	scanFn := m.rowScanFn
	m.mu.Unlock()
	return &mockRow{err: err, scanFn: scanFn}
}

func (m *MockQueryExecutor) Queries() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]string, len(m.queries))
	copy(copied, m.queries)
	return copied
}

func (m *MockQueryExecutor) SetExecError(err error) {
	m.mu.Lock()
	m.execErr = err
	m.mu.Unlock()
}

func (m *MockQueryExecutor) SetQueryRowError(err error) {
	m.mu.Lock()
	m.queryRowErr = err
	m.mu.Unlock()
}

func (m *MockQueryExecutor) SetRowScanFn(fn func(dest ...any) error) {
	m.mu.Lock()
	m.rowScanFn = fn
	m.mu.Unlock()
}

type mockRow struct {
	err    error
	scanFn func(dest ...any) error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if r.scanFn != nil {
		return r.scanFn(dest...)
	}
	return nil
}

type mockRows struct {
	closed bool
}

func (r *mockRows) Close()                                       { r.closed = true }
func (r *mockRows) Err() error                                   { return nil }
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.NewCommandTag("SELECT 0") }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) Next() bool                                   { return false }
func (r *mockRows) Scan(dest ...any) error                       { return nil }
func (r *mockRows) Values() ([]any, error)                       { return nil, nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }

// FakePgxTx implements pgx.Tx to allow instantiating real *database.Tx in tests.
type FakePgxTx struct {
	committed  bool
	rolledBack bool
	execErr    error
	commitErr  error
}

func (f *FakePgxTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return f, nil
}

func (f *FakePgxTx) Commit(ctx context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *FakePgxTx) Rollback(ctx context.Context) error {
	f.rolledBack = true
	return nil
}

func (f *FakePgxTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (f *FakePgxTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (f *FakePgxTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (f *FakePgxTx) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	if f.execErr != nil {
		return pgconn.CommandTag{}, f.execErr
	}
	return pgconn.NewCommandTag("EXEC 1"), nil
}

func (f *FakePgxTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return &mockRows{}, nil
}

func (f *FakePgxTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return &mockRow{}
}

func (f *FakePgxTx) Conn() *pgx.Conn {
	return nil
}

func (f *FakePgxTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

// NewTestTx constructs a real *database.Tx wrapping a FakePgxTx.
func NewTestTx() (*database.Tx, *FakePgxTx) {
	fake := &FakePgxTx{}
	return database.NewTx(fake), fake
}

// =============================================================================
// MockRunner: Connection Invariant & Pool Starvation Testing
// =============================================================================

// MockRunner implements database.Runner and tracks connection acquisitions,
// enforcing that concurrent operations do not exceed MaxConns (default 1).
type MockRunner struct {
	mu            sync.Mutex
	maxConns      int32
	activeConns   int32
	withDBCount   int32
	withTxCount   int32
	commitCount   int32
	rollbackCount int32
	txErr         error
	dbErr         error
	commitErr     error
	executor      *MockQueryExecutor
	hopDetected   bool
}

func NewMockRunner(maxConns ...int32) *MockRunner {
	limit := int32(1)
	if len(maxConns) > 0 && maxConns[0] > 0 {
		limit = maxConns[0]
	}
	return &MockRunner{
		maxConns: limit,
		executor: NewMockQueryExecutor(),
	}
}

func (r *MockRunner) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	atomic.AddInt32(&r.withTxCount, 1)

	if _, ok := database.TxFromContext(ctx); ok {
		return fn(ctx)
	}

	// Single connection check: simulate connection acquisition.
	current := atomic.AddInt32(&r.activeConns, 1)
	defer atomic.AddInt32(&r.activeConns, -1)

	if current > r.maxConns {
		r.mu.Lock()
		r.hopDetected = true
		r.mu.Unlock()
		return fmt.Errorf("%w: pool connection limit exceeded (active=%d, max=%d)", database.ErrConnectionFailure, current, r.maxConns)
	}

	if r.txErr != nil {
		atomic.AddInt32(&r.rollbackCount, 1)
		return r.txErr
	}

	fakeTx := &FakePgxTx{commitErr: r.commitErr}
	tx := database.NewTx(fakeTx)
	txCtx := database.WithTx(ctx, tx)

	var panicOccurred bool
	defer func() {
		if p := recover(); p != nil {
			panicOccurred = true
			atomic.AddInt32(&r.rollbackCount, 1)
			_ = fakeTx.Rollback(context.Background())
			panic(p)
		}
	}()

	err := fn(txCtx)
	if err != nil {
		atomic.AddInt32(&r.rollbackCount, 1)
		_ = fakeTx.Rollback(context.Background())
		return err
	}

	if err := fakeTx.Commit(ctx); err != nil {
		atomic.AddInt32(&r.rollbackCount, 1)
		return database.MapError(err)
	}

	atomic.AddInt32(&r.commitCount, 1)
	_ = panicOccurred
	return nil
}

func (r *MockRunner) WithExecutor(ctx context.Context, fn func(ctx context.Context) error) error {
	atomic.AddInt32(&r.withDBCount, 1)

	if _, ok := database.TxFromContext(ctx); ok {
		return fn(ctx)
	}
	if _, err := database.Executor(ctx); err == nil {
		return fn(ctx)
	}

	current := atomic.AddInt32(&r.activeConns, 1)
	defer atomic.AddInt32(&r.activeConns, -1)

	if current > r.maxConns {
		r.mu.Lock()
		r.hopDetected = true
		r.mu.Unlock()
		return fmt.Errorf("%w: pool connection limit exceeded (active=%d, max=%d)", database.ErrConnectionFailure, current, r.maxConns)
	}

	if r.dbErr != nil {
		return r.dbErr
	}

	ctx = database.WithExecutor(ctx, r.executor)
	return fn(ctx)
}

func (r *MockRunner) SetTxError(err error) {
	r.mu.Lock()
	r.txErr = err
	r.mu.Unlock()
}

func (r *MockRunner) SetDBError(err error) {
	r.mu.Lock()
	r.dbErr = err
	r.mu.Unlock()
}

func (r *MockRunner) SetCommitError(err error) {
	r.mu.Lock()
	r.commitErr = err
	r.mu.Unlock()
}

func (r *MockRunner) Stats() (withDB, withTx, commits, rollbacks int) {
	return int(atomic.LoadInt32(&r.withDBCount)),
		int(atomic.LoadInt32(&r.withTxCount)),
		int(atomic.LoadInt32(&r.commitCount)),
		int(atomic.LoadInt32(&r.rollbackCount))
}

func (r *MockRunner) ConnectionHopDetected() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hopDetected
}

// =============================================================================
// MockStorageProvider
// =============================================================================

// MockStorageProvider satisfies storage.StorageProvider directly in memory.
type MockStorageProvider struct {
	mu            sync.RWMutex
	objects       map[string][]byte
	uploadCalls   int
	deleteCalls   int
	urlCalls      int
	uploadErr     error
	deleteErr     error
	publicBaseURL string
}

func NewMockStorageProvider(baseURL ...string) *MockStorageProvider {
	base := "https://cdn.example.com"
	if len(baseURL) > 0 && baseURL[0] != "" {
		base = baseURL[0]
	}
	return &MockStorageProvider{
		objects:       make(map[string][]byte),
		publicBaseURL: base,
	}
}

func (s *MockStorageProvider) Upload(ctx context.Context, bucket, key string, file io.Reader, size int64, contentType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploadCalls++
	if s.uploadErr != nil {
		return s.uploadErr
	}
	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	data, _ := io.ReadAll(file)
	s.objects[fullKey] = data
	return nil
}

func (s *MockStorageProvider) Delete(ctx context.Context, bucket, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteCalls++
	if s.deleteErr != nil {
		return s.deleteErr
	}
	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	delete(s.objects, fullKey)
	return nil
}

func (s *MockStorageProvider) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	return fmt.Sprintf("%s/%s/%s?expires=%d", s.publicBaseURL, bucket, key, int(expires.Seconds())), nil
}

func (s *MockStorageProvider) GetPublicURL(bucket, key string) string {
	s.mu.Lock()
	s.urlCalls++
	s.mu.Unlock()
	return fmt.Sprintf("%s/%s/%s", s.publicBaseURL, bucket, key)
}

func (s *MockStorageProvider) Stats() (uploads, deletes, urlLookups int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.uploadCalls, s.deleteCalls, s.urlCalls
}

func (s *MockStorageProvider) SetUploadError(err error) {
	s.mu.Lock()
	s.uploadErr = err
	s.mu.Unlock()
}

func (s *MockStorageProvider) SetDeleteError(err error) {
	s.mu.Lock()
	s.deleteErr = err
	s.mu.Unlock()
}

func (s *MockStorageProvider) HasObject(bucket, key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.objects[fmt.Sprintf("%s/%s", bucket, key)]
	return exists
}

// =============================================================================
// Canonical Domain Error Handling Reference Mapper
// =============================================================================

// CanonicalHandleError provides the authoritative reference implementation of
// Rule 5: Canonical Domain Error Mapping via handleError.
func CanonicalHandleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return err // Rule 5: Already mapped - do not double wrap
	}

	mappedErr := database.MapError(err)

	if errors.Is(mappedErr, database.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
		return apierr.NotFoundError(fallbackMsg + " not found").Wrap(err)
	}

	if errors.Is(mappedErr, database.ErrConflict) {
		return apierr.ConflictError(fallbackMsg + " already exists or conflict occurred").Wrap(err)
	}

	if errors.Is(mappedErr, database.ErrForeignKeyViolation) {
		return apierr.BadRequestError("Referenced entity does not exist for " + fallbackMsg).Wrap(err)
	}

	if errors.Is(mappedErr, database.ErrNotNullViolation) {
		return apierr.BadRequestError("Missing required field for " + fallbackMsg).Wrap(err)
	}

	if errors.Is(mappedErr, database.ErrCheckViolation) {
		return apierr.BadRequestError("Constraint check violation for " + fallbackMsg).Wrap(err)
	}

	if errors.Is(mappedErr, database.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
		return apierr.New(408, "Request timed out during "+fallbackMsg).
			WithCode("TIMEOUT").
			Wrap(err)
	}

	if errors.Is(mappedErr, database.ErrCanceled) || errors.Is(err, context.Canceled) {
		return apierr.New(499, "Request canceled during "+fallbackMsg).
			WithCode(apierr.CodeClientClosedRequest).
			Wrap(err)
	}

	// Rule 5: Internal errors expose no raw SQL, table names, or driver details.
	return apierr.InternalServerError("Failed to " + fallbackMsg).
		WithCode(apierr.CodeInternalError).
		Wrap(err)
}

// =============================================================================
// In-Memory Repository Test Doubles for Cross-Domain Interactions
// =============================================================================

// MockInventoryRepository implements the consumer InventoryRepository interfaces
// for order and checkout domains.
type MockInventoryRepository struct {
	mu           sync.Mutex
	stocks       map[uuid.UUID]*model.InventoryStock
	reservations map[uuid.UUID]model.InventoryReservation
	ledgers      []model.InventoryLedger
	lockedIDs    map[uuid.UUID]bool
	releaseErr   error
	stockErr     error
}

func NewMockInventoryRepository() *MockInventoryRepository {
	return &MockInventoryRepository{
		stocks:       make(map[uuid.UUID]*model.InventoryStock),
		reservations: make(map[uuid.UUID]model.InventoryReservation),
		ledgers:      make([]model.InventoryLedger, 0),
		lockedIDs:    make(map[uuid.UUID]bool),
	}
}

func (m *MockInventoryRepository) SetStock(variantID uuid.UUID, onHand, reserved int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stocks[variantID] = &model.InventoryStock{
		VariantID:         variantID,
		OnHandQuantity:    onHand,
		ReservedQuantity:  reserved,
		AvailableQuantity: onHand - reserved,
	}
}

func (m *MockInventoryRepository) LockVariantForUpdate(ctx context.Context, variantID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lockedIDs[variantID] = true
	return nil
}

func (m *MockInventoryRepository) GetStock(ctx context.Context, variantID uuid.UUID) (*model.InventoryStock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stockErr != nil {
		return nil, m.stockErr
	}
	s, ok := m.stocks[variantID]
	if !ok {
		return nil, database.ErrNotFound
	}
	return s, nil
}

func (m *MockInventoryRepository) GetStockForVariants(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID]*model.InventoryStock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make(map[uuid.UUID]*model.InventoryStock)
	for _, id := range variantIDs {
		if s, ok := m.stocks[id]; ok {
			res[id] = s
		}
	}
	return res, nil
}

func (m *MockInventoryRepository) CreateReservation(ctx context.Context, res *model.InventoryReservation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reservations[res.ID] = *res
	return nil
}

func (m *MockInventoryRepository) GetUnreleasedReservationsForOrder(ctx context.Context, orderID uuid.UUID) ([]model.InventoryReservation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []model.InventoryReservation
	for _, r := range m.reservations {
		if r.OrderID != nil && *r.OrderID == orderID && r.ReleasedAt == nil {
			list = append(list, r)
		}
	}
	return list, nil
}

func (m *MockInventoryRepository) ReleaseReservation(ctx context.Context, reservationID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.reservations[reservationID]; ok {
		now := time.Now()
		r.ReleasedAt = &now
		m.reservations[reservationID] = r
	}
	return nil
}

func (m *MockInventoryRepository) ReleaseReservationsForOrder(ctx context.Context, orderID uuid.UUID) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.releaseErr != nil {
		return 0, m.releaseErr
	}
	var count int64
	now := time.Now()
	for k, r := range m.reservations {
		if r.OrderID != nil && *r.OrderID == orderID && r.ReleasedAt == nil {
			r.ReleasedAt = &now
			m.reservations[k] = r
			count++
		}
	}
	return count, nil
}

func (m *MockInventoryRepository) ExpiredReservationOrderIDs(ctx context.Context) ([]uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	seen := make(map[uuid.UUID]bool)
	var expired []uuid.UUID
	for _, r := range m.reservations {
		if r.ExpiresAt.Before(now) && r.ReleasedAt == nil && r.OrderID != nil && !seen[*r.OrderID] {
			seen[*r.OrderID] = true
			expired = append(expired, *r.OrderID)
		}
	}
	return expired, nil
}

func (m *MockInventoryRepository) CreateLedger(ctx context.Context, ledger *model.InventoryLedger) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ledgers = append(m.ledgers, *ledger)
	return nil
}

func (m *MockInventoryRepository) Ledgers() []model.InventoryLedger {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]model.InventoryLedger, len(m.ledgers))
	copy(copied, m.ledgers)
	return copied
}

// MockCouponRepository satisfies checkout consumer coupon repository.
type MockCouponRepository struct {
	mu          sync.Mutex
	coupons     map[uuid.UUID]*model.Coupon
	codeIndex   map[string]uuid.UUID
	redemptions []model.CouponRedemption
}

func NewMockCouponRepository() *MockCouponRepository {
	return &MockCouponRepository{
		coupons:     make(map[uuid.UUID]*model.Coupon),
		codeIndex:   make(map[string]uuid.UUID),
		redemptions: make([]model.CouponRedemption, 0),
	}
}

func (m *MockCouponRepository) AddCoupon(c *model.Coupon) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.coupons[c.ID] = c
	m.codeIndex[c.Code] = c.ID
}

func (m *MockCouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Coupon, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.coupons[id]
	if !ok {
		return nil, database.ErrNotFound
	}
	return c, nil
}

func (m *MockCouponRepository) GetByCode(ctx context.Context, code string) (*model.Coupon, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.codeIndex[code]
	if !ok {
		return nil, database.ErrNotFound
	}
	return m.coupons[id], nil
}

func (m *MockCouponRepository) GetRedemptionCount(ctx context.Context, couponID uuid.UUID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, r := range m.redemptions {
		if r.CouponID == couponID {
			count++
		}
	}
	return count, nil
}

func (m *MockCouponRepository) GetUserRedemptionCount(ctx context.Context, couponID uuid.UUID, userID uuid.UUID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, r := range m.redemptions {
		if r.CouponID == couponID && r.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockCouponRepository) RecordRedemption(ctx context.Context, red *model.CouponRedemption) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.redemptions = append(m.redemptions, *red)
	return nil
}

// MockUserRepository & AddressRepository test doubles.
type MockUserRepository struct {
	mu    sync.Mutex
	users map[uuid.UUID]*model.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{users: make(map[uuid.UUID]*model.User)}
}

func (m *MockUserRepository) AddUser(u *model.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.ID] = u
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, database.ErrNotFound
	}
	return u, nil
}

type MockAddressRepository struct {
	mu        sync.Mutex
	addresses map[uuid.UUID]*model.UserAddress
}

func NewMockAddressRepository() *MockAddressRepository {
	return &MockAddressRepository{addresses: make(map[uuid.UUID]*model.UserAddress)}
}

func (m *MockAddressRepository) AddAddress(a *model.UserAddress) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addresses[a.ID] = a
}

func (m *MockAddressRepository) GetByIDAndUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.UserAddress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.addresses[id]
	if !ok || a.UserID != userID {
		return nil, database.ErrNotFound
	}
	return a, nil
}

// MockCartRepository test double.
type MockCartRepository struct {
	mu        sync.Mutex
	carts     map[uuid.UUID]*model.Cart
	userCarts map[uuid.UUID]uuid.UUID
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{
		carts:     make(map[uuid.UUID]*model.Cart),
		userCarts: make(map[uuid.UUID]uuid.UUID),
	}
}

func (m *MockCartRepository) SetCart(c *model.Cart) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.carts[c.ID] = c
	if c.UserID != nil {
		m.userCarts[*c.UserID] = c.ID
	}
}

func (m *MockCartRepository) GetCartByUserID(ctx context.Context, userID uuid.UUID) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cartID, ok := m.userCarts[userID]
	if !ok {
		return nil, database.ErrNotFound
	}
	return m.carts[cartID], nil
}

func (m *MockCartRepository) GetCartBySessionID(ctx context.Context, sessionID string) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.carts {
		if c.SessionID != nil && *c.SessionID == sessionID {
			return c, nil
		}
	}
	return nil, database.ErrNotFound
}

func (m *MockCartRepository) ClearCart(ctx context.Context, cartID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.carts[cartID]; ok {
		c.Items = nil
		m.carts[cartID] = c
	}
	return nil
}

// MockVariantRepository test double.
type MockVariantRepository struct {
	mu       sync.Mutex
	variants map[uuid.UUID]*model.ProductVariant
	byProd   map[uuid.UUID][]uuid.UUID
}

func NewMockVariantRepository() *MockVariantRepository {
	return &MockVariantRepository{
		variants: make(map[uuid.UUID]*model.ProductVariant),
		byProd:   make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *MockVariantRepository) AddVariant(v *model.ProductVariant) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.variants[v.ID] = v
	m.byProd[v.ProductID] = append(m.byProd[v.ProductID], v.ID)
}

func (m *MockVariantRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ProductVariant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.variants[id]
	if !ok {
		return nil, database.ErrNotFound
	}
	return v, nil
}

func (m *MockVariantRepository) CountByProductID(ctx context.Context, productID uuid.UUID) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.byProd[productID]), nil
}

// MockProductRepository test double.
type MockProductRepository struct {
	mu       sync.Mutex
	products map[uuid.UUID]*model.Product
}

func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{products: make(map[uuid.UUID]*model.Product)}
}

func (m *MockProductRepository) AddProduct(p *model.Product) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.products[p.ID] = p
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.products[id]
	if !ok {
		return nil, database.ErrNotFound
	}
	return p, nil
}

func (m *MockProductRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PublicationStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.products[id]
	if !ok {
		return database.ErrNotFound
	}
	p.Status = status
	return nil
}

// MockObjectRepository test double.
type MockObjectRepository struct {
	mu      sync.Mutex
	objects map[uuid.UUID]*model.Object
	byKey   map[string]uuid.UUID
}

func NewMockObjectRepository() *MockObjectRepository {
	return &MockObjectRepository{
		objects: make(map[uuid.UUID]*model.Object),
		byKey:   make(map[string]uuid.UUID),
	}
}

func (m *MockObjectRepository) Create(ctx context.Context, obj *model.Object) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[obj.ID] = obj
	key := fmt.Sprintf("%s/%s", obj.Bucket, obj.Key)
	m.byKey[key] = obj.ID
	return nil
}

func (m *MockObjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Object, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	obj, ok := m.objects[id]
	if !ok {
		return nil, database.ErrNotFound
	}
	return obj, nil
}

func (m *MockObjectRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Object, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []*model.Object
	for _, id := range ids {
		if obj, ok := m.objects[id]; ok {
			list = append(list, obj)
		}
	}
	return list, nil
}

func (m *MockObjectRepository) MarkDeletingByKey(ctx context.Context, bucket, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	id, ok := m.byKey[fullKey]
	if !ok {
		return database.ErrNotFound
	}
	m.objects[id].Status = model.ObjectStatusDeleting
	return nil
}

func (m *MockObjectRepository) MarkDeletedByKey(ctx context.Context, bucket, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	fullKey := fmt.Sprintf("%s/%s", bucket, key)
	id, ok := m.byKey[fullKey]
	if !ok {
		return database.ErrNotFound
	}
	now := time.Now()
	m.objects[id].Status = model.ObjectStatusDeleted
	m.objects[id].DeletedAt = &now
	return nil
}

// MockOrderRepository test double.
type MockOrderRepository struct {
	mu     sync.Mutex
	orders map[uuid.UUID]*model.Order
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{orders: make(map[uuid.UUID]*model.Order)}
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, o *model.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[o.ID] = o
	return nil
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, database.ErrNotFound
	}
	return o, nil
}

func (m *MockOrderRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return database.ErrNotFound
	}
	o.Status = status
	return nil
}

// MockAuditRepository test double.
type MockAuditRepository struct {
	mu      sync.Mutex
	entries []*model.AuditLog
}

func NewMockAuditRepository() *MockAuditRepository {
	return &MockAuditRepository{entries: make([]*model.AuditLog, 0)}
}

func (m *MockAuditRepository) Insert(ctx context.Context, entry *model.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockAuditRepository) Entries() []*model.AuditLog {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]*model.AuditLog, len(m.entries))
	copy(copied, m.entries)
	return copied
}
