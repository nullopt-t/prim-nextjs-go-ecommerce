package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

const ResetTokenTTL = 15 * time.Minute

type UserService struct {
	dbExecuter database.Runner
	repo       *UserRepository
	logger     log.Logger
}

func NewService(
	dbExecuter database.Runner,
	repo *UserRepository,
	logger log.Logger,
) *UserService {
	return &UserService{
		dbExecuter: dbExecuter,
		repo:       repo,
		logger:     logger,
	}
}

func (s *UserService) get(
	ctx context.Context,
	db database.QueryExecutor,
	filter Filter,
) (*model.User, error) {
	user, err := s.repo.Get(
		ctx,
		db,
		filter,
	)
	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(
			mappedErr,
			database.ErrNotFound,
		):
			return nil, apierr.New(
				http.StatusNotFound,
				"user not found",
			)
		default:
			return nil, apierr.New(
				http.StatusInternalServerError,
				"failed to get a user",
			).Wrap(err)
		}
	}
	return user, nil
}

func (s *UserService) CreateUser(
	ctx context.Context,
	identifier string,
) (*model.User, error) {

	now := time.Now()
	u := &model.User{
		Identifier: identifier,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := s.dbExecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			id, err := s.repo.Create(
				ctx,
				db,
				u,
			)
			if err != nil {
				mappedErr := database.MapError(err)
				switch {
				case errors.Is(
					mappedErr,
					database.ErrConflict,
				):
					return apierr.New(
						http.StatusConflict,
						"user already exists",
					)
				default:
					return apierr.New(
						http.StatusInternalServerError,
						"failed to create a new user",
					).Wrap(err)
				}
			}
			u.ID = id
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	s.logger.Info("created user", log.Meta{
		"user": u,
	})
	return u, nil
}

func (s *UserService) GetUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (*model.User, error) {
	var user *model.User
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			user, err = s.get(
				ctx,
				db,
				Filter{
					ID: &userID,
				},
			)
			if err != nil {
				return err
			}
			return nil
		})
	return user, err
}

func (s *UserService) GetUserByIdentifier(
	ctx context.Context,
	identifier string,
) (*model.User, error) {
	var user *model.User
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			u, err := s.get(
				ctx,
				db,
				Filter{
					Identifier: &identifier,
				},
			)
			if err != nil {
				return err
			}
			user = u
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return user, err
}

func (s *UserService) GetAllUsers(
	ctx context.Context,
	q api.ListQuery,
) ([]model.User, api.Page, error) {
	var users []model.User
	var page api.Page
	err := s.dbExecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			var err error
			users, page, err = s.repo.GetAll(
				ctx,
				db,
				q,
			)
			if err != nil {
				return apierr.New(
					http.StatusInternalServerError,
					"failed to fetch users",
				).Wrap(err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, page, err
	}
	return users, page, nil
}

func (s *UserService) DeleteUserByID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	err := s.dbExecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return s.repo.Delete(
				ctx,
				db,
				Filter{
					ID: &userID,
				},
			)
		},
	)
	if err != nil {
		return err
	}
	return nil
}
