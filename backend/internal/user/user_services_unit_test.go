package user_test

// In-memory repository doubles for the address and notification services. They
// encode the same tenancy and default-address rules as the SQL repositories so
// the service-layer state machine can be tested without a database.

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type fakeAddressRepo struct {
	mu    sync.Mutex
	items map[uuid.UUID]*model.UserAddress
	seq   int
}

func newFakeAddressRepo() *fakeAddressRepo {
	return &fakeAddressRepo{items: make(map[uuid.UUID]*model.UserAddress)}
}

func (f *fakeAddressRepo) stamp(a *model.UserAddress) {
	f.seq++
	ts := time.Unix(int64(f.seq), 0).UTC()
	a.CreatedAt = ts
	a.UpdatedAt = ts
}

func (f *fakeAddressRepo) active(userID uuid.UUID) []*model.UserAddress {
	var list []*model.UserAddress
	for _, a := range f.items {
		if a.UserID == userID && a.DeletedAt == nil {
			list = append(list, a)
		}
	}
	return list
}

func (f *fakeAddressRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]model.UserAddress, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	list := f.active(userID)
	sort.Slice(list, func(i, j int) bool {
		if list[i].IsDefault != list[j].IsDefault {
			return list[i].IsDefault
		}
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})
	out := make([]model.UserAddress, 0, len(list))
	for _, a := range list {
		out = append(out, *a)
	}
	return out, nil
}

func (f *fakeAddressRepo) GetByIDAndUser(_ context.Context, id, userID uuid.UUID) (*model.UserAddress, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.items[id]
	if !ok || a.UserID != userID || a.DeletedAt != nil {
		return nil, database.ErrNotFound
	}
	copy := *a
	return &copy, nil
}

func (f *fakeAddressRepo) LockUserAddresses(context.Context, uuid.UUID) error { return nil }

func (f *fakeAddressRepo) CountByUser(_ context.Context, userID uuid.UUID) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.active(userID)), nil
}

func (f *fakeAddressRepo) ClearDefault(_ context.Context, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.active(userID) {
		a.IsDefault = false
	}
	return nil
}

func (f *fakeAddressRepo) SetDefault(_ context.Context, id, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.items[id]
	if !ok || a.UserID != userID || a.DeletedAt != nil {
		return user.ErrAddressNotFound
	}
	a.IsDefault = true
	return nil
}

func (f *fakeAddressRepo) Create(_ context.Context, a *model.UserAddress) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stamp(a)
	copy := *a
	f.items[a.ID] = &copy
	return nil
}

func (f *fakeAddressRepo) Update(_ context.Context, a *model.UserAddress) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.items[a.ID]
	if !ok || existing.UserID != a.UserID || existing.DeletedAt != nil {
		return user.ErrAddressNotFound
	}
	a.UpdatedAt = existing.UpdatedAt.Add(time.Second)
	copy := *a
	f.items[a.ID] = &copy
	return nil
}

func (f *fakeAddressRepo) SoftDelete(_ context.Context, id, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.items[id]
	if !ok || a.UserID != userID || a.DeletedAt != nil {
		return user.ErrAddressNotFound
	}
	now := time.Now()
	a.DeletedAt = &now
	a.IsDefault = false
	return nil
}

func (f *fakeAddressRepo) FindMostRecentActive(_ context.Context, userID uuid.UUID) (*model.UserAddress, error) {
	return f.mostRecent(userID, uuid.Nil)
}

func (f *fakeAddressRepo) FindMostRecentActiveExcluding(_ context.Context, userID, excludeID uuid.UUID) (*model.UserAddress, error) {
	return f.mostRecent(userID, excludeID)
}

func (f *fakeAddressRepo) mostRecent(userID, excludeID uuid.UUID) (*model.UserAddress, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var best *model.UserAddress
	for _, a := range f.active(userID) {
		if a.ID == excludeID {
			continue
		}
		if best == nil || a.UpdatedAt.After(best.UpdatedAt) {
			best = a
		}
	}
	if best == nil {
		return nil, user.ErrAddressNotFound
	}
	copy := *best
	return &copy, nil
}

type fakeNotificationRepo struct {
	mu    sync.Mutex
	items map[int64]*model.UserNotification
	seq   int64
}

func newFakeNotificationRepo() *fakeNotificationRepo {
	return &fakeNotificationRepo{items: make(map[int64]*model.UserNotification)}
}

func (f *fakeNotificationRepo) Create(_ context.Context, n *model.UserNotification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seq++
	n.ID = f.seq
	n.CreatedAt = time.Unix(f.seq, 0).UTC()
	copy := *n
	f.items[n.ID] = &copy
	return nil
}

func (f *fakeNotificationRepo) ListByUser(_ context.Context, userID uuid.UUID, unreadOnly bool, limit int, cursor *pagination.Keyset) ([]*model.UserNotification, *pagination.CursorPage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var all []*model.UserNotification
	for _, n := range f.items {
		if n.UserID != userID {
			continue
		}
		if unreadOnly && n.ReadAt != nil {
			continue
		}
		if cursor != nil {
			if !n.CreatedAt.Before(cursor.CreatedAt) && !(n.CreatedAt.Equal(cursor.CreatedAt) && n.ID < cursor.IDInt) {
				continue
			}
		}
		copy := *n
		all = append(all, &copy)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].CreatedAt.Equal(all[j].CreatedAt) {
			return all[i].ID > all[j].ID
		}
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	items, page := pagination.NewCursorPageFrom(all, limit, func(n *model.UserNotification) pagination.Keyset {
		return pagination.Keyset{
			CreatedAt: n.CreatedAt,
			IDInt:     n.ID,
		}
	})
	return items, page, nil
}

func (f *fakeNotificationRepo) CountUnreadByUser(_ context.Context, userID uuid.UUID) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, n := range f.items {
		if n.UserID == userID && n.ReadAt == nil {
			count++
		}
	}
	return count, nil
}

func (f *fakeNotificationRepo) MarkAsRead(_ context.Context, id int64, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, ok := f.items[id]
	if !ok || n.UserID != userID {
		return pgx.ErrNoRows
	}
	if n.ReadAt == nil {
		now := time.Now()
		n.ReadAt = &now
	}
	return nil
}

func (f *fakeNotificationRepo) MarkAllAsRead(_ context.Context, userID uuid.UUID) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var count int64
	now := time.Now()
	for _, n := range f.items {
		if n.UserID == userID && n.ReadAt == nil {
			n.ReadAt = &now
			count++
		}
	}
	return count, nil
}

func (f *fakeNotificationRepo) Delete(_ context.Context, id int64, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	n, ok := f.items[id]
	if !ok || n.UserID != userID {
		return pgx.ErrNoRows
	}
	delete(f.items, id)
	return nil
}

func newAddressService() (*user.AddressService, *fakeAddressRepo) {
	repo := newFakeAddressRepo()
	return user.NewAddressService(fakeRunner{}, repo, log.NewNoopLogger()), repo
}

func addressInput(name string) *user.CreateAddressInput {
	return &user.CreateAddressInput{
		FullName: "  " + name + " ",
		Line1:    " 12 Test Street ",
		City:     " Cairo ",
		Country:  " EG ",
	}
}

func requireStatus(t *testing.T, err error, status int) {
	t.Helper()
	var apiErr *apierr.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, status, apiErr.Status)
}

func TestAddressService_Unit_FirstAddressDefaultAndTrim(t *testing.T) {
	svc, _ := newAddressService()
	ctx := context.Background()
	userID := uuid.New()

	first, err := svc.CreateAddress(ctx, userID, addressInput("Alice"))
	require.NoError(t, err)
	assert.True(t, first.IsDefault)
	assert.Equal(t, "Alice", first.FullName)
	assert.Equal(t, "12 Test Street", first.Line1)
	assert.Equal(t, "Cairo", first.City)
	assert.Equal(t, "EG", first.Country)

	second, err := svc.CreateAddress(ctx, userID, addressInput("Bob"))
	require.NoError(t, err)
	assert.False(t, second.IsDefault)
}

func TestAddressService_Unit_OwnershipIsolation(t *testing.T) {
	svc, _ := newAddressService()
	ctx := context.Background()
	owner, intruder := uuid.New(), uuid.New()

	addr, err := svc.CreateAddress(ctx, owner, addressInput("Owner"))
	require.NoError(t, err)

	_, err = svc.GetAddress(ctx, addr.ID, intruder)
	requireStatus(t, err, 404)

	_, err = svc.UpdateAddress(ctx, addr.ID, intruder, &user.UpdateAddressInput{
		FullName: "Hacked", Line1: "x", City: "y", Country: "z",
	})
	requireStatus(t, err, 404)

	_, err = svc.SetDefaultAddress(ctx, addr.ID, intruder)
	requireStatus(t, err, 404)

	err = svc.DeleteAddress(ctx, addr.ID, intruder)
	requireStatus(t, err, 404)

	reloaded, err := svc.GetAddress(ctx, addr.ID, owner)
	require.NoError(t, err)
	assert.Equal(t, "Owner", reloaded.FullName)
	assert.True(t, reloaded.IsDefault)
}

func TestAddressService_Unit_DefaultTransitions(t *testing.T) {
	svc, _ := newAddressService()
	ctx := context.Background()
	userID := uuid.New()

	first, err := svc.CreateAddress(ctx, userID, addressInput("First"))
	require.NoError(t, err)
	second, err := svc.CreateAddress(ctx, userID, addressInput("Second"))
	require.NoError(t, err)

	_, err = svc.SetDefaultAddress(ctx, second.ID, userID)
	require.NoError(t, err)

	list, err := svc.ListAddresses(ctx, userID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.True(t, list[0].IsDefault)
	assert.Equal(t, second.ID, list[0].ID)
	assert.False(t, list[1].IsDefault)

	// Demoting the default promotes the other address.
	_, err = svc.UpdateAddress(ctx, second.ID, userID, &user.UpdateAddressInput{
		FullName: second.FullName, Line1: second.Line1, City: second.City, Country: second.Country,
		IsDefault: false,
	})
	require.NoError(t, err)

	reloadedFirst, err := svc.GetAddress(ctx, first.ID, userID)
	require.NoError(t, err)
	assert.True(t, reloadedFirst.IsDefault)
	reloadedSecond, err := svc.GetAddress(ctx, second.ID, userID)
	require.NoError(t, err)
	assert.False(t, reloadedSecond.IsDefault)

	// Deleting the default promotes the remaining address; deleting the last
	// leaves the user with no default and no addresses.
	require.NoError(t, svc.DeleteAddress(ctx, first.ID, userID))
	reloadedSecond, err = svc.GetAddress(ctx, second.ID, userID)
	require.NoError(t, err)
	assert.True(t, reloadedSecond.IsDefault)

	require.NoError(t, svc.DeleteAddress(ctx, second.ID, userID))
	list, err = svc.ListAddresses(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestAddressService_Unit_CannotUnsetOnlyDefault(t *testing.T) {
	svc, _ := newAddressService()
	ctx := context.Background()
	userID := uuid.New()

	only, err := svc.CreateAddress(ctx, userID, addressInput("Only"))
	require.NoError(t, err)

	_, err = svc.UpdateAddress(ctx, only.ID, userID, &user.UpdateAddressInput{
		FullName: only.FullName, Line1: only.Line1, City: only.City, Country: only.Country,
		IsDefault: false,
	})
	requireStatus(t, err, 400)
}

func TestAddressService_Unit_Validation(t *testing.T) {
	svc, _ := newAddressService()
	ctx := context.Background()
	userID := uuid.New()

	cases := []*user.CreateAddressInput{
		{FullName: "  ", Line1: "line", City: "city", Country: "EG"},
		{FullName: "Alice", Line1: " ", City: "city", Country: "EG"},
		{FullName: "Alice", Line1: "line", City: "\t", Country: "EG"},
		{FullName: "Alice", Line1: "line", City: "city", Country: ""},
	}
	for i, in := range cases {
		_, err := svc.CreateAddress(ctx, userID, in)
		requireStatus(t, err, 400)
		_ = i
	}
}

func TestAddressService_Unit_InvalidPhoneRejected(t *testing.T) {
	svc, _ := newAddressService()
	ctx := context.Background()
	userID := uuid.New()

	badPhones := []string{"abc", "+441234567890", "011234", "012345678", "+2012345678901"}
	for _, phone := range badPhones {
		in := addressInput("Alice")
		in.Phone = &phone
		_, err := svc.CreateAddress(ctx, userID, in)
		requireStatus(t, err, 400)
	}

	goodPhones := []string{"01012345678", "+201012345678", "0020 111 234 5678", "01123456789"}
	for _, phone := range goodPhones {
		in := addressInput("Alice")
		in.Phone = &phone
		_, err := svc.CreateAddress(ctx, userID, in)
		require.NoError(t, err)
	}
}

func TestAddressService_Unit_InvalidPostalCodeRejected(t *testing.T) {
	svc, _ := newAddressService()
	ctx := context.Background()
	userID := uuid.New()

	for _, postal := range []string{"123", "123456", "12A45"} {
		in := addressInput("Alice")
		in.PostalCode = &postal
		_, err := svc.CreateAddress(ctx, userID, in)
		requireStatus(t, err, 400)
	}

	in := addressInput("Alice")
	valid := "11222"
	in.PostalCode = &valid
	_, err := svc.CreateAddress(ctx, userID, in)
	require.NoError(t, err)
}

func TestAddressService_Unit_UpdateRejectsInvalidPhone(t *testing.T) {
	svc, _ := newAddressService()
	ctx := context.Background()
	userID := uuid.New()

	created, err := svc.CreateAddress(ctx, userID, addressInput("Alice"))
	require.NoError(t, err)

	bad := "not-a-phone"
	_, err = svc.UpdateAddress(ctx, created.ID, userID, &user.UpdateAddressInput{
		FullName: "Alice", Line1: "line", City: "Cairo", Country: "EG", Phone: &bad,
		IsDefault: true,
	})
	requireStatus(t, err, 400)
}

func TestNotificationService_Unit_LifecycleAndIsolation(t *testing.T) {
	repo := newFakeNotificationRepo()
	svc := user.NewNotificationService(fakeRunner{}, repo, log.NewNoopLogger())
	ctx := context.Background()
	owner, intruder := uuid.New(), uuid.New()

	create := func(read bool) int64 {
		n := &model.UserNotification{
			UserID:   owner,
			Title:    []byte(`{"en":"title"}`),
			Message:  []byte(`{"en":"body"}`),
			Category: "system",
		}
		if read {
			now := time.Now()
			n.ReadAt = &now
		}
		require.NoError(t, svc.CreateNotification(ctx, n))
		return n.ID
	}

	first := create(false)
	second := create(false)
	create(true)

	count, err := svc.GetUnreadCount(ctx, owner)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	all, page, err := svc.GetMyNotifications(ctx, owner, false, 10, nil)
	require.NoError(t, err)
	assert.Len(t, all, 3)
	assert.False(t, page.HasMore)

	unread, unreadPage, err := svc.GetMyNotifications(ctx, owner, true, 10, nil)
	require.NoError(t, err)
	assert.Len(t, unread, 2)
	assert.False(t, unreadPage.HasMore)

	require.NoError(t, svc.MarkAsRead(ctx, first, owner))
	count, err = svc.GetUnreadCount(ctx, owner)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	marked, err := svc.MarkAllAsRead(ctx, owner)
	require.NoError(t, err)
	assert.Equal(t, int64(1), marked)

	count, err = svc.GetUnreadCount(ctx, owner)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	require.NoError(t, svc.DeleteNotification(ctx, second, owner))

	// Cross-user mutations are not found; reads never leak rows.
	err = svc.MarkAsRead(ctx, first, intruder)
	requireStatus(t, err, 404)
	err = svc.DeleteNotification(ctx, first, intruder)
	requireStatus(t, err, 404)

	other, otherPage, err := svc.GetMyNotifications(ctx, intruder, false, 10, nil)
	require.NoError(t, err)
	assert.Empty(t, other)
	assert.False(t, otherPage.HasMore)

	otherCount, err := svc.GetUnreadCount(ctx, intruder)
	require.NoError(t, err)
	assert.Equal(t, 0, otherCount)
}
