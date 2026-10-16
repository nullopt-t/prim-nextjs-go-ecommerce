package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// SessionManager defines the session revocation contract required by the user domain
// to forcefully terminate all active Redis sessions for a user upon security events or administrative action.
type SessionManager interface {
	DeleteAllSessions(ctx context.Context, userID uuid.UUID) error
}

// RoleRank defines numeric privilege levels for role hierarchy enforcement.
var RoleRank = map[model.UserRole]int{
	model.CustomerRole: 1,
	model.VendorRole:   2,
	model.SupportRole:  3,
	model.AdminRole:    4,
	model.SuperRole:    5,
}

func GetRoleRank(role model.UserRole) int {
	if rank, ok := RoleRank[role]; ok {
		return rank
	}
	return 1
}

// UserService orchestrates business logic, state machines, and lifecycle rules for user accounts:
// account creation, retrieval by UUID/identifier, paginated queries, administrative role mutations,
// account suspension/status transitions, and session invalidation.
type UserService struct {
	dbExecuter database.Runner
	repo       *UserRepository
	sessions   SessionManager
	logger     log.Logger
}

// NewService constructs a new UserService domain component.
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

// SetSessionManager assigns the SessionManager dependency (typically session_service from auth package),
// resolving circular package dependency through interface abstraction.
func (s *UserService) SetSessionManager(sm SessionManager) {
	s.sessions = sm
}

// CreateUser provisions a new user record in PostgreSQL for the specified email identifier.
// Enforces unique constraint handling by mapping database conflict errors to HTTP 409 Conflict.
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

// GetUserByID retrieves a single active user by their UUID.
// Returns a 404 Not Found error if the user does not exist or has been soft-deleted.
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

// GetUserByIdentifier retrieves a single active user by their email address.
// Returns a 404 Not Found error if no matching record exists.
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

// GetAllUsers queries a paginated list of non-deleted users matching optional role and status filters.
func (s *UserService) GetAllUsers(
	ctx context.Context,
	q pagination.ListQuery,
	role *string,
	status *string,
) ([]model.User, *pagination.Page, error) {
	var users []model.User
	var page *pagination.Page

	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		// PageSize/Offset are already clamped and overflow-safe from ListQuery.Process
		limit := q.PageSize
		offset := q.Offset

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
		page = pagination.NewPage(q.Page, limit, totalCount)
		return nil
	})

	if err != nil {
		s.logger.Error("failed to list users", log.Meta{"error": err.Error()})
		return nil, nil, apierr.InternalServerError("Failed to list users").Wrap(err)
	}

	return users, page, nil
}

// DeleteUserByID soft-deletes a user account and revokes all active sessions.
func (s *UserService) DeleteUserByID(
	ctx context.Context,
	userID uuid.UUID,
	callerID uuid.UUID,
) error {
	if userID == callerID {
		return apierr.BadRequestError("Users cannot delete their own account").
			WithCode("CANNOT_DELETE_SELF")
	}

	callerUser, err := s.GetUserByID(ctx, callerID)
	if err != nil {
		return apierr.UnauthorizedError("Caller not found")
	}

	targetUser, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Hierarchy check: cannot delete a user with equal or higher rank
	if GetRoleRank(callerUser.Role) <= GetRoleRank(targetUser.Role) {
		return apierr.ForbiddenError("You cannot delete a user with equal or higher privileges").
			WithCode("CANNOT_MODIFY_EQUAL_OR_HIGHER_ROLE")
	}

	err = s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.SoftDelete(ctx, db, userID)
	})
	if err != nil {
		mappedErr := database.MapError(err)
		if errors.Is(mappedErr, database.ErrNotFound) {
			return apierr.NotFoundError("User not found")
		}
		s.logger.Error("failed to delete user", log.Meta{"userId": userID, "error": err.Error()})
		return apierr.InternalServerError("Failed to delete user").Wrap(err)
	}

	// Invalidate all active sessions for the deleted user
	if s.sessions != nil {
		if err := s.sessions.DeleteAllSessions(ctx, userID); err != nil {
			s.logger.Warn("failed to revoke user sessions upon deletion", log.Meta{"userId": userID, "error": err.Error()})
		}
	}

	s.logger.Info("user successfully deleted", log.Meta{"userId": userID})
	return nil
}

// UpdateUserRole updates the administrative role for a user (Superadmin only).
// Prevents superadmins from modifying their own role to avoid self-lockout or privilege loss.
func (s *UserService) UpdateUserRole(
	ctx context.Context,
	targetUserID uuid.UUID,
	newRole model.UserRole,
	callerID uuid.UUID,
) error {
	// Prevent self-demotion or self-modification
	if targetUserID == callerID {
		return apierr.BadRequestError("Users cannot alter their own role").
			WithCode("CANNOT_MODIFY_OWN_ROLE")
	}

	callerUser, err := s.GetUserByID(ctx, callerID)
	if err != nil {
		return apierr.UnauthorizedError("Caller not found")
	}

	targetUser, err := s.GetUserByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	// Role hierarchy: higher rank means greater privileges
	roleRanks := map[model.UserRole]int{
		model.CustomerRole: 1,
		model.VendorRole:   2,
		model.SupportRole:  3,
		model.AdminRole:    4,
		model.SuperRole:    5,
	}

	callerRank, callerValid := roleRanks[callerUser.Role]
	if !callerValid {
		return apierr.ForbiddenError("Caller does not have a valid administrative role").
			WithCode("INSUFFICIENT_PERMISSIONS")
	}

	targetRank, targetValid := roleRanks[targetUser.Role]
	if !targetValid {
		return apierr.BadRequestError("Target user has an unrecognized role").
			WithCode("INVALID_USER_ROLE")
	}

	newRank, newValid := roleRanks[newRole]
	if !newValid {
		return apierr.BadRequestError(fmt.Sprintf("Invalid target role: %s", newRole)).
			WithCode("INVALID_USER_ROLE")
	}

	// 1. Caller cannot modify users with an equal or higher role
	if callerRank <= targetRank {
		return apierr.ForbiddenError("You cannot modify the role of a user with equal or higher privileges").
			WithCode("CANNOT_MODIFY_EQUAL_OR_HIGHER_ROLE")
	}

	// 2. Caller cannot promote anyone to their own role or any higher role
	if newRank >= callerRank {
		return apierr.ForbiddenError("You cannot assign a role equal to or higher than your own").
			WithCode("CANNOT_PROMOTE_TO_OR_ABOVE_OWN_ROLE")
	}

	// 3. Enforce single superadmin policy: multiple superadmins are not permitted
	if newRole == model.SuperRole {
		return apierr.BadRequestError("Only one Superadministrator is permitted in the system").
			WithCode("MULTIPLE_SUPERADMINS_FORBIDDEN")
	}

	// Update role in database
	err = s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.UpdateRole(ctx, db, targetUserID, newRole)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		if errors.Is(mappedErr, database.ErrNotFound) {
			return apierr.NotFoundError("User not found")
		}
		if errors.Is(mappedErr, database.ErrConflict) {
			return apierr.ConflictError("Only one Superadministrator is permitted in the system").
				WithCode("MULTIPLE_SUPERADMINS_FORBIDDEN")
		}
		s.logger.Error("failed to update user role", log.Meta{"userId": targetUserID, "role": newRole, "error": err.Error()})
		return apierr.InternalServerError("Failed to update user role").Wrap(err)
	}

	// Revoke existing active sessions so permissions/roles update immediately
	if err := s.RevokeAllUserSessions(ctx, targetUserID, callerID); err != nil {
		s.logger.Warn("failed to revoke sessions on role update", log.Meta{"userId": targetUserID, "error": err.Error()})
	}

	s.logger.Info("user role updated", log.Meta{"userId": targetUserID, "newRole": newRole, "updatedBy": callerID})
	return nil
}

// UpdateAccountStatus updates account standing (active, suspended, deleted) (Superadmin only).
// Enforces self-suspension guardrails and sanitizes the suspended_until window.
func (s *UserService) UpdateAccountStatus(
	ctx context.Context,
	targetUserID uuid.UUID,
	newStatus model.AccountStatus,
	suspendedUntil *time.Time,
	callerID uuid.UUID,
) error {
	// Prevent modifying own status
	if targetUserID == callerID {
		return apierr.BadRequestError("Users cannot alter their own account status").
			WithCode("CANNOT_SUSPEND_SELF")
	}

	callerUser, err := s.GetUserByID(ctx, callerID)
	if err != nil {
		return apierr.UnauthorizedError("Caller not found")
	}

	targetUser, err := s.GetUserByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	// Hierarchy check: cannot suspend/alter status of users with equal or higher rank
	if GetRoleRank(callerUser.Role) <= GetRoleRank(targetUser.Role) {
		return apierr.ForbiddenError("You cannot alter the status of a user with equal or higher privileges").
			WithCode("CANNOT_MODIFY_EQUAL_OR_HIGHER_ROLE")
	}

	// Validate status value
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

	// Clear suspendedUntil if transitioning away from suspended status
	if newStatus != model.StatusSuspended {
		suspendedUntil = nil
	}

	// Update status in PostgreSQL
	err = s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
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

	// Invalidate active sessions immediately on suspension or deletion
	if newStatus == model.StatusSuspended || newStatus == model.StatusDeleted {
		if err := s.RevokeAllUserSessions(ctx, targetUserID, callerID); err != nil {
			s.logger.Warn("failed to revoke sessions on account status update", log.Meta{"userId": targetUserID, "error": err.Error()})
		}
	}

	s.logger.Info("user status updated", log.Meta{"userId": targetUserID, "newStatus": newStatus, "updatedBy": callerID})
	return nil
}

// RevokeAllUserSessions invalidates all active login sessions in Redis for the target user.
func (s *UserService) RevokeAllUserSessions(
	ctx context.Context,
	targetUserID uuid.UUID,
	callerID uuid.UUID,
) error {
	if targetUserID == callerID {
		return apierr.BadRequestError("Users cannot revoke their own sessions via this administrative endpoint").
			WithCode("CANNOT_REVOKE_OWN_SESSIONS")
	}

	callerUser, err := s.GetUserByID(ctx, callerID)
	if err != nil {
		return apierr.UnauthorizedError("Caller not found")
	}

	targetUser, err := s.GetUserByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	// Hierarchy check: cannot revoke sessions of users with equal or higher rank
	if GetRoleRank(callerUser.Role) <= GetRoleRank(targetUser.Role) {
		return apierr.ForbiddenError("You cannot revoke sessions of a user with equal or higher privileges").
			WithCode("CANNOT_MODIFY_EQUAL_OR_HIGHER_ROLE")
	}

	if s.sessions != nil {
		if err := s.sessions.DeleteAllSessions(ctx, targetUserID); err != nil {
			s.logger.Error("failed to revoke user sessions", log.Meta{"userId": targetUserID, "error": err.Error()})
			return apierr.InternalServerError("Failed to revoke user sessions").Wrap(err)
		}
	}
	s.logger.Info("all sessions revoked for user", log.Meta{"userId": targetUserID, "revokedBy": callerID})
	return nil
}


