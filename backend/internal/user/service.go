package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

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

func (s *UserService) CreateUser(
	ctx context.Context,
	identifier string,
) (*model.User, error) {
	now := time.Now()
	u := &model.User{
		ID:         uuid.New(),
		Identifier: identifier,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		id, err := s.repo.Create(ctx, db, u)
		if err != nil {
			mappedErr := database.MapError(err)
			switch {
			case errors.Is(mappedErr, database.ErrConflict):
				return apierr.New(http.StatusConflict, "user already exists")
			default:
				return apierr.New(http.StatusInternalServerError, "failed to create a new user").Wrap(err)
			}
		}
		u.ID = id
		return nil
	})

	if err != nil {
		return nil, err
	}
	s.logger.Info("created user", log.Meta{"user": u})
	return u, nil
}

func (s *UserService) GetUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (*model.User, error) {
	var user *model.User
	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		u, err := s.repo.GetByID(ctx, db, userID)
		if err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrNotFound) {
				return apierr.ErrNotFound("User not found")
			}
			return apierr.ErrInternalError("Failed to fetch user").Wrap(err)
		}
		user = u
		return nil
	})
	return user, err
}

func (s *UserService) GetUserByIdentifier(
	ctx context.Context,
	identifier string,
) (*model.User, error) {
	var user *model.User
	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		u, err := s.repo.GetByIdentifier(ctx, db, identifier)
		if err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrNotFound) {
				return apierr.ErrNotFound("User not found")
			}
			return apierr.ErrInternalError("Failed to fetch user").Wrap(err)
		}
		user = u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetAllUsers(
	ctx context.Context,
	q pagination.ListQuery,
) ([]model.User, pagination.Page, error) {
	var users []model.User
	var page pagination.Page
	return users, page, nil
}

func (s *UserService) DeleteUserByID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return nil
}
