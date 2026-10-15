package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type SessionManager interface {
	DeleteAllSessions(ctx context.Context, userID uuid.UUID) error
}

type UserService struct {
	dbExecuter database.Runner
	repo       *UserRepository
	sessions   SessionManager
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

func (s *UserService) SetSessionManager(sm SessionManager) {
	s.sessions = sm
}

func (s *UserService) CreateUser(
	ctx context.Context,
	identifier string,
) (*model.User, error) {
	u := &model.User{
		ID:         uuid.New(),
		Identifier: identifier,
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
				return apierr.NotFoundError("User not found")
			}
			return apierr.InternalServerError("Failed to fetch user").Wrap(err)
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
				return apierr.NotFoundError("User not found")
			}
			return apierr.InternalServerError("Failed to fetch user").Wrap(err)
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
	role *string,
	status *string,
) ([]model.User, *pagination.Page, error) {
	var users []model.User
	var page *pagination.Page

	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		limit := q.PageSize
		if limit <= 0 {
			limit = 10
		}
		pageNumber := q.Page
		if pageNumber <= 0 {
			pageNumber = 1
		}
		offset := (pageNumber - 1) * limit

		var searchPtr *string
		if q.Search != "" {
			searchPtr = &q.Search
		}

		filter := UserListFilter{
			Role:   role,
			Status: status,
			Search: searchPtr,
			Limit:  limit,
			Offset: offset,
		}

		totalCount, err := s.repo.CountUsers(ctx, db, filter)
		if err != nil {
			return err
		}

		uList, err := s.repo.ListUsers(ctx, db, filter)
		if err != nil {
			return err
		}

		users = uList
		page = pagination.NewPage(pageNumber, limit, totalCount)
		return nil
	})

	if err != nil {
		s.logger.Error("failed to list users", log.Meta{"error": err.Error()})
		return nil, nil, apierr.InternalServerError("Failed to list users").Wrap(err)
	}

	return users, page, nil
}

func (s *UserService) DeleteUserByID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return nil
}

func (s *UserService) UpdateUserRole(
	ctx context.Context,
	targetUserID uuid.UUID,
	newRole model.UserRole,
	callerID uuid.UUID,
) error {
	if targetUserID == callerID {
		return apierr.BadRequestError("Superadmins cannot alter their own role").
			WithCode("CANNOT_MODIFY_OWN_ROLE")
	}

	validRoles := map[model.UserRole]bool{
		model.CustomerRole: true,
		model.VendorRole:   true,
		model.SupportRole:  true,
		model.AdminRole:    true,
		model.SuperRole:    true,
	}

	if !validRoles[newRole] {
		return apierr.BadRequestError(fmt.Sprintf("Invalid role: %s", newRole)).
			WithCode("INVALID_USER_ROLE")
	}

	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.UpdateRole(ctx, db, targetUserID, newRole)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		if errors.Is(mappedErr, database.ErrNotFound) {
			return apierr.NotFoundError("User not found")
		}
		s.logger.Error("failed to update user role", log.Meta{"userId": targetUserID, "role": newRole, "error": err.Error()})
		return apierr.InternalServerError("Failed to update user role").Wrap(err)
	}

	s.logger.Info("user role updated", log.Meta{"userId": targetUserID, "newRole": newRole, "updatedBy": callerID})
	return nil
}

func (s *UserService) UpdateAccountStatus(
	ctx context.Context,
	targetUserID uuid.UUID,
	newStatus model.AccountStatus,
	suspendedUntil *time.Time,
	callerID uuid.UUID,
) error {
	if targetUserID == callerID {
		return apierr.BadRequestError("Superadmins cannot suspend their own account").
			WithCode("CANNOT_SUSPEND_SELF")
	}

	validStatuses := map[model.AccountStatus]bool{
		model.StatusActive:    true,
		model.StatusInactive:  true,
		model.StatusSuspended: true,
		model.StatusDeleted:   true,
	}

	if !validStatuses[newStatus] {
		return apierr.BadRequestError(fmt.Sprintf("Invalid status: %s", newStatus)).
			WithCode("INVALID_ACCOUNT_STATUS")
	}

	if newStatus != model.StatusSuspended {
		suspendedUntil = nil
	}

	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.UpdateStatus(ctx, db, targetUserID, newStatus, suspendedUntil)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		if errors.Is(mappedErr, database.ErrNotFound) {
			return apierr.NotFoundError("User not found")
		}
		s.logger.Error("failed to update user status", log.Meta{"userId": targetUserID, "status": newStatus, "error": err.Error()})
		return apierr.InternalServerError("Failed to update user status").Wrap(err)
	}

	s.logger.Info("user status updated", log.Meta{"userId": targetUserID, "newStatus": newStatus, "updatedBy": callerID})
	return nil
}

func (s *UserService) RevokeAllUserSessions(
	ctx context.Context,
	targetUserID uuid.UUID,
	callerID uuid.UUID,
) error {
	if s.sessions != nil {
		if err := s.sessions.DeleteAllSessions(ctx, targetUserID); err != nil {
			s.logger.Error("failed to revoke user sessions", log.Meta{"userId": targetUserID, "error": err.Error()})
			return apierr.InternalServerError("Failed to revoke user sessions").Wrap(err)
		}
	}
	s.logger.Info("all sessions revoked for user", log.Meta{"userId": targetUserID, "revokedBy": callerID})
	return nil
}


