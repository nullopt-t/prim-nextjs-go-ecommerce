package user_test

// Database-backed tests for the customer address and notification services.
// Both are user-scoped and rely on SQL tenancy filters plus a per-user advisory
// lock, so they are exercised against real PostgreSQL rather than fakes.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type UserDomainSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *database.DB
	runner      database.Runner
	addresses   *user.AddressService
	notifs      *user.NotificationService
}

func (s *UserDomainSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_user_domain_test"),
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
		DBUser: "testuser", DBPassword: "testpass", DBName: "prim_user_domain_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	schema, err := os.ReadFile("../../migrations/000001_init_sechema.up.sql")
	require.NoError(s.T(), err)
	_, err = db.Exec(ctx, string(schema))
	require.NoError(s.T(), err)

	logger := log.NewNoopLogger()
	s.runner = database.NewTxRunner(db)
	s.addresses = user.NewAddressService(s.runner, user.NewAddressRepository(), logger)
	s.notifs = user.NewNotificationService(s.runner, user.NewNotificationRepository(), logger)
}

func (s *UserDomainSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *UserDomainSuite) seedUser() uuid.UUID {
	id := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO users (id, email, full_name, role, is_email_verified, status)
		VALUES ($1, $2, $3, 'customer', true, 'active')
	`, id, "user_domain_"+id.String()+"@example.com", "User Domain")
	s.Require().NoError(err)
	return id
}

func (s *UserDomainSuite) newAddressInput(name string) *user.CreateAddressInput {
	return &user.CreateAddressInput{
		FullName: name,
		Line1:    " 12 Test Street ",
		City:     " Cairo ",
		Country:  " EG ",
	}
}

func statusOf(t *testing.T, err error) int {
	t.Helper()
	var apiErr *apierr.APIError
	require.ErrorAs(t, err, &apiErr)
	return apiErr.Status
}

func (s *UserDomainSuite) TestAddress_FirstAddressBecomesDefaultAndValuesAreTrimmed() {
	userID := s.seedUser()

	created, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput(" Alice "))
	s.Require().NoError(err)
	s.True(created.IsDefault)
	s.Equal("Alice", created.FullName)
	s.Equal("12 Test Street", created.Line1)
	s.Equal("Cairo", created.City)
	s.Equal("EG", created.Country)

	second, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("Bob"))
	s.Require().NoError(err)
	s.False(second.IsDefault, "only the first address is auto-promoted")
}

func (s *UserDomainSuite) TestAddress_OwnershipIsolation() {
	owner := s.seedUser()
	intruder := s.seedUser()

	addr, err := s.addresses.CreateAddress(context.Background(), owner, s.newAddressInput("Owner"))
	s.Require().NoError(err)

	_, err = s.addresses.GetAddress(context.Background(), addr.ID, intruder)
	s.Equal(404, statusOf(s.T(), err))

	in := &user.UpdateAddressInput{FullName: "Hacked", Line1: "x", City: "y", Country: "z"}
	_, err = s.addresses.UpdateAddress(context.Background(), addr.ID, intruder, in)
	s.Equal(404, statusOf(s.T(), err))

	_, err = s.addresses.SetDefaultAddress(context.Background(), addr.ID, intruder)
	s.Equal(404, statusOf(s.T(), err))

	err = s.addresses.DeleteAddress(context.Background(), addr.ID, intruder)
	s.Equal(404, statusOf(s.T(), err))

	// The owner's row is untouched.
	reloaded, err := s.addresses.GetAddress(context.Background(), addr.ID, owner)
	s.Require().NoError(err)
	s.Equal("Owner", reloaded.FullName)
	s.True(reloaded.IsDefault)
}

func (s *UserDomainSuite) TestAddress_SetDefaultClearsPrevious() {
	userID := s.seedUser()

	first, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("First"))
	s.Require().NoError(err)
	second, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("Second"))
	s.Require().NoError(err)
	s.False(second.IsDefault)

	_, err = s.addresses.SetDefaultAddress(context.Background(), second.ID, userID)
	s.Require().NoError(err)

	list, err := s.addresses.ListAddresses(context.Background(), userID)
	s.Require().NoError(err)
	s.Require().Len(list, 2)
	s.True(list[0].IsDefault)
	s.Equal(second.ID, list[0].ID)
	for _, addr := range list[1:] {
		s.False(addr.IsDefault)
	}

	// Idempotent when already default.
	again, err := s.addresses.SetDefaultAddress(context.Background(), second.ID, userID)
	s.Require().NoError(err)
	s.True(again.IsDefault)
	_ = first
}

func (s *UserDomainSuite) TestAddress_UnsettingDefaultPromotesAnother() {
	userID := s.seedUser()

	first, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("First"))
	s.Require().NoError(err)
	second, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("Second"))
	s.Require().NoError(err)

	_, err = s.addresses.UpdateAddress(context.Background(), first.ID, userID, &user.UpdateAddressInput{
		FullName:  first.FullName,
		Line1:     first.Line1,
		City:      first.City,
		Country:   first.Country,
		IsDefault: false,
	})
	s.Require().NoError(err)

	reloadedFirst, err := s.addresses.GetAddress(context.Background(), first.ID, userID)
	s.Require().NoError(err)
	s.False(reloadedFirst.IsDefault)

	reloadedSecond, err := s.addresses.GetAddress(context.Background(), second.ID, userID)
	s.Require().NoError(err)
	s.True(reloadedSecond.IsDefault, "demoting the default must promote another address")
}

func (s *UserDomainSuite) TestAddress_CannotUnsetOnlyDefault() {
	userID := s.seedUser()
	only, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("Only"))
	s.Require().NoError(err)

	_, err = s.addresses.UpdateAddress(context.Background(), only.ID, userID, &user.UpdateAddressInput{
		FullName:  only.FullName,
		Line1:     only.Line1,
		City:      only.City,
		Country:   only.Country,
		IsDefault: false,
	})
	s.Equal(400, statusOf(s.T(), err))
}

func (s *UserDomainSuite) TestAddress_DeletingDefaultPromotesAnotherAndLastLeavesNone() {
	userID := s.seedUser()

	first, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("First"))
	s.Require().NoError(err)
	second, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("Second"))
	s.Require().NoError(err)

	s.Require().NoError(s.addresses.DeleteAddress(context.Background(), first.ID, userID))
	reloadedSecond, err := s.addresses.GetAddress(context.Background(), second.ID, userID)
	s.Require().NoError(err)
	s.True(reloadedSecond.IsDefault)

	s.Require().NoError(s.addresses.DeleteAddress(context.Background(), second.ID, userID))
	list, err := s.addresses.ListAddresses(context.Background(), userID)
	s.Require().NoError(err)
	s.Empty(list)
}

func (s *UserDomainSuite) TestAddress_ValidationRejectsWhitespaceFields() {
	userID := s.seedUser()

	cases := []struct {
		name  string
		input *user.CreateAddressInput
	}{
		{
			name: "blank full name",
			input: &user.CreateAddressInput{
				FullName: "  ", Line1: "line", City: "city", Country: "EG",
			},
		},
		{
			name: "blank line1",
			input: &user.CreateAddressInput{
				FullName: "Alice", Line1: " ", City: "city", Country: "EG",
			},
		},
		{
			name: "blank city",
			input: &user.CreateAddressInput{
				FullName: "Alice", Line1: "line", City: "\t", Country: "EG",
			},
		},
		{
			name: "blank country",
			input: &user.CreateAddressInput{
				FullName: "Alice", Line1: "line", City: "city", Country: "",
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, err := s.addresses.CreateAddress(context.Background(), userID, tc.input)
			s.Equal(400, statusOf(s.T(), err))
		})
	}
}

func (s *UserDomainSuite) TestAddress_ConcurrentFirstCreationsProduceOneDefault() {
	userID := s.seedUser()

	start := make(chan struct{})
	results := make(chan *model.UserAddress, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			addr, err := s.addresses.CreateAddress(context.Background(), userID, s.newAddressInput("Concurrent"))
			if err != nil {
				results <- nil
				return
			}
			results <- addr
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	created := 0
	for addr := range results {
		if addr != nil {
			created++
		}
	}
	// The advisory lock may serialize the two creations; both may succeed, but
	// at most one may be marked default.
	var defaults int
	s.Require().NoError(s.db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM addresses WHERE user_id = $1 AND deleted_at IS NULL AND is_default = true`,
		userID).Scan(&defaults))
	s.Equal(1, defaults)
	s.Equal(2, created)
}

func (s *UserDomainSuite) createNotification(userID uuid.UUID, title string, read bool) int64 {
	n := &model.UserNotification{
		UserID:   userID,
		Title:    json.RawMessage(`{"en":"` + title + `"}`),
		Message:  json.RawMessage(`{"en":"body"}`),
		Category: "system",
	}
	if read {
		now := time.Now()
		n.ReadAt = &now
	}
	s.Require().NoError(s.notifs.CreateNotification(context.Background(), n))
	return n.ID
}

func (s *UserDomainSuite) TestNotification_CreateListCountReadAndDelete() {
	userID := s.seedUser()

	first := s.createNotification(userID, "first", false)
	second := s.createNotification(userID, "second", false)
	s.createNotification(userID, "already-read", true)

	count, err := s.notifs.GetUnreadCount(context.Background(), userID)
	s.Require().NoError(err)
	s.Equal(2, count)

	page, err := s.notifs.GetMyNotifications(context.Background(), userID, false, &pagination.ListQuery{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.Len(page.Items, 3)
	s.Equal(3, page.Page.TotalItems)

	unread, err := s.notifs.GetMyNotifications(context.Background(), userID, true, &pagination.ListQuery{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.Len(unread.Items, 2)
	for _, item := range unread.Items {
		s.Nil(item.ReadAt)
	}

	s.Require().NoError(s.notifs.MarkAsRead(context.Background(), first, userID))
	count, err = s.notifs.GetUnreadCount(context.Background(), userID)
	s.Require().NoError(err)
	s.Equal(1, count)

	// Idempotent re-mark.
	s.Require().NoError(s.notifs.MarkAsRead(context.Background(), first, userID))

	marked, err := s.notifs.MarkAllAsRead(context.Background(), userID)
	s.Require().NoError(err)
	s.Equal(int64(1), marked)

	count, err = s.notifs.GetUnreadCount(context.Background(), userID)
	s.Require().NoError(err)
	s.Equal(0, count)

	s.Require().NoError(s.notifs.DeleteNotification(context.Background(), second, userID))
	rest, err := s.notifs.GetMyNotifications(context.Background(), userID, false, &pagination.ListQuery{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.Len(rest.Items, 2)
}

func (s *UserDomainSuite) TestNotification_CrossUserMutationsAreNotFound() {
	owner := s.seedUser()
	intruder := s.seedUser()
	notificationID := s.createNotification(owner, "private", false)

	err := s.notifs.MarkAsRead(context.Background(), notificationID, intruder)
	s.Equal(404, statusOf(s.T(), err))

	err = s.notifs.DeleteNotification(context.Background(), notificationID, intruder)
	s.Equal(404, statusOf(s.T(), err))

	list, err := s.notifs.GetMyNotifications(context.Background(), intruder, false, &pagination.ListQuery{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.Empty(list.Items)

	count, err := s.notifs.GetUnreadCount(context.Background(), intruder)
	s.Require().NoError(err)
	s.Equal(0, count)

	// The owner can still read it.
	s.Require().NoError(s.notifs.MarkAsRead(context.Background(), notificationID, owner))
}

func TestUserDomainSuite(t *testing.T) {
	// Integration coverage needs a Docker daemon; the in-memory unit suite
	// covers the same service rules when it is unavailable.
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon unavailable; skipping PostgreSQL integration suite")
	}
	suite.Run(t, new(UserDomainSuite))
}
