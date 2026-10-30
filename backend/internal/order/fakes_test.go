package order_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type fakeUserReader struct {
	users map[uuid.UUID]*model.User
	err   error
}

func (f *fakeUserReader) GetByID(_ context.Context, _ database.QueryExecutor, id uuid.UUID) (*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.users != nil {
		if u, ok := f.users[id]; ok {
			return u, nil
		}
	}
	return nil, nil
}

type fakeSettingsProvider struct {
	config *setting.Config
}

func (f *fakeSettingsProvider) GetConfig(context.Context) (setting.Config, error) {
	if f.config != nil {
		return *f.config, nil
	}
	return setting.DefaultConfig(), nil
}

type fakeOrderNotifier struct {
	shipped   []notifier.OrderShippedPayload
	cancelled []notifier.OrderCancelledPayload
}

func (f *fakeOrderNotifier) NotifyOrderShipped(_ context.Context, in notifier.OrderShippedPayload) error {
	f.shipped = append(f.shipped, in)
	return nil
}

func (f *fakeOrderNotifier) NotifyOrderCancelled(_ context.Context, in notifier.OrderCancelledPayload) error {
	f.cancelled = append(f.cancelled, in)
	return nil
}
