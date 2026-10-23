package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// SessionManager defines the session contracts required by the user domain:
// reading active sessions for oversight, targeted revocation, and forced
// termination of all sessions for a user on security events.
type SessionManager interface {
	DeleteAllSessions(ctx context.Context, userID uuid.UUID) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]model.UserSession, error)
	GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error)
	DeleteSessionByID(ctx context.Context, userID uuid.UUID, sessionID string) error
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

// handleError maps persistence and database errors to domain APIErrors.
func (s *UserService) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError("User not found").Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("User already exists").Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced entity does not exist").Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).Wrap(err).WithStack()
	}
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
			if errors.Is(mappedErr, database.ErrConflict) {
				return apierr.New(http.StatusConflict, "user already exists")
			}
			return s.handleError(err, "failed to create a new user")
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
		var repoErr error
		user, repoErr = s.getUserByID(ctx, db, userID)
		return repoErr
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserExists reports whether a non-deleted user with the given id exists. It is
// used by the authentication layer to detect sessions whose principal was
// removed out-of-band (for example a database reseed), so those requests fail
// with 401 instead of hitting a foreign-key error on the first write.
func (s *UserService) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, nil
	}

	var exists bool
	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		exists, repoErr = s.repo.ExistsByID(ctx, db, userID)
		return repoErr
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *UserService) getUserByID(
	ctx context.Context,
	exec database.QueryExecutor,
	userID uuid.UUID,
) (*model.User, error) {
	u, err := s.repo.GetByID(ctx, exec, userID)
	if err != nil {
		return nil, s.handleError(err, "Failed to fetch user")
	}
	return u, nil
}

// GetUserByIdentifier retrieves a single active user by their email address.
// Returns a 404 Not Found error if no matching record exists.
func (s *UserService) GetUserByIdentifier(
	ctx context.Context,
	identifier string,
) (*model.User, error) {
	var user *model.User
	err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		user, repoErr = s.repo.GetByIdentifier(ctx, db, identifier)
		return repoErr
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to fetch user")
	}
	return user, nil
}

// GetAllUsers queries a keyset-paginated list of non-deleted users matching optional role and status filters.
func (s *UserService) GetAllUsers(
	ctx context.Context,
	q pagination.ListQuery,
	role *string,
	status *string,
) ([]model.User, *pagination.CursorPage, error) {
	var users []model.User
	var page *pagination.CursorPage

	cursor, err := q.Keyset()
	if err != nil {
		return nil, nil, apierr.BadRequestError("Invalid pagination cursor").Wrap(err)
	}

	err = s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		// PageSize is already clamped and overflow-safe from ListQuery.Process
		limit := q.PageSize

		var searchPtr *string
		if q.Search != "" {
			searchPtr = &q.Search
		}

		filter := UserListFilter{
			Role:   role,
			Status: status,
			Search: searchPtr,
			Limit:  limit,
			Cursor: cursor,
		}

		uList, err := s.repo.ListUsers(ctx, db, filter)
		if err != nil {
			return err
		}

		users, page = pagination.NewCursorPageFrom(uList, limit, func(u model.User) pagination.Keyset {
			return pagination.Keyset{CreatedAt: u.CreatedAt, ID: u.ID}
		})
		return nil
	})

	if err != nil {
		s.logger.Error("failed to list users", log.Meta{"error": err.Error()})
		return nil, nil, s.handleError(err, "Failed to list users")
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

	// Revoke all active sessions BEFORE the soft delete, and fail closed. The
	// auth middleware trusts the session snapshot, so a still-valid session
	// would let a deleted account keep operating until its TTL expired.
	if s.sessions != nil {
		if err := s.sessions.DeleteAllSessions(ctx, userID); err != nil {
			s.logger.Error("failed to revoke user sessions upon deletion", log.Meta{"userId": userID, "error": err.Error()})
			return apierr.InternalServerError("Failed to revoke user sessions").Wrap(err)
		}
	}

	err = s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.SoftDelete(ctx, db, userID)
	})
	if err != nil {
		s.logger.Error("failed to delete user", log.Meta{"userId": userID, "error": err.Error()})
		return s.handleError(err, "Failed to delete user")
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

	// Revoke existing sessions BEFORE committing the role change, and fail
	// closed. Authenticate() trusts the role carried in the session snapshot
	// without a per-request DB re-check, so an un-revoked session would retain
	// its previous privileges until it expired. Revoking first guarantees we
	// never commit a privilege change we failed to enforce; the worst case is an
	// unnecessary logout if the subsequent update fails.
	if err := s.RevokeAllUserSessions(ctx, targetUserID, callerID); err != nil {
		s.logger.Error("failed to revoke sessions on role update", log.Meta{"userId": targetUserID, "error": err.Error()})
		return err
	}

	// Update role in database
	err = s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.UpdateRole(ctx, db, targetUserID, newRole)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		if errors.Is(mappedErr, database.ErrConflict) {
			return apierr.ConflictError("Only one Superadministrator is permitted in the system").
				WithCode("MULTIPLE_SUPERADMINS_FORBIDDEN")
		}
		s.logger.Error("failed to update user role", log.Meta{"userId": targetUserID, "role": newRole, "error": err.Error()})
		return s.handleError(err, "Failed to update user role")
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

	// Revoke active sessions BEFORE committing a suspension/deletion, and fail
	// closed: the session snapshot carries the account's role/status, so an
	// un-revoked suspended account would keep operating until its session TTL.
	if newStatus == model.StatusSuspended || newStatus == model.StatusDeleted {
		if err := s.RevokeAllUserSessions(ctx, targetUserID, callerID); err != nil {
			s.logger.Error("failed to revoke sessions on account status update", log.Meta{"userId": targetUserID, "error": err.Error()})
			return err
		}
	}

	// Update status in PostgreSQL. A deleted status is a real soft delete:
	// status and deleted_at are set atomically so the account disappears from
	// listings.
	err = s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		if newStatus == model.StatusDeleted {
			return s.repo.SoftDelete(ctx, db, targetUserID)
		}
		return s.repo.UpdateStatus(ctx, db, targetUserID, newStatus, suspendedUntil)
	})

	if err != nil {
		s.logger.Error("failed to update user status", log.Meta{"userId": targetUserID, "status": newStatus, "error": err.Error()})
		return s.handleError(err, "Failed to update user status")
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

// staffRoles lists the roles considered "staff" for session oversight. Customer
// sessions are deliberately excluded: they are numerous and out of scope for
// device management.
var staffRoles = []model.UserRole{
	model.SuperRole,
	model.AdminRole,
	model.SupportRole,
	model.VendorRole,
}

// StaffSession is the administrative view of an active staff login device.
type StaffSession struct {
	ID         string    `json:"id"`
	UserID     uuid.UUID `json:"userId"`
	Email      string    `json:"email"`
	FullName   *string   `json:"fullName,omitempty"`
	Role       string    `json:"role"`
	UserAgent  string    `json:"userAgent"`
	IPAddress  string    `json:"ipAddress"`
	CreatedAt  time.Time `json:"createdAt"`
	LastActive time.Time `json:"lastActive"`
	// Current marks the session making the oversight request.
	Current bool `json:"current"`
}

// ListStaffSessions aggregates every active Redis session belonging to a staff
// account, optionally narrowed to a single staff role. Results are sorted by
// most recent activity so suspicious or forgotten devices surface first.
func (s *UserService) ListStaffSessions(
	ctx context.Context,
	roleFilter *string,
	currentSessionID string,
) ([]StaffSession, error) {
	roles := staffRoles
	if roleFilter != nil && *roleFilter != "" {
		role := model.UserRole(strings.ToLower(strings.TrimSpace(*roleFilter)))
		if GetRoleRank(role) < GetRoleRank(model.VendorRole) {
			return nil, apierr.BadRequestError("role must be a staff role (vendor, support, admin or superadmin)").
				WithCode(apierr.CodeInvalidInput)
		}
		roles = []model.UserRole{role}
	}

	if s.sessions == nil {
		return nil, apierr.InternalServerError("Session store unavailable")
	}

	var users []model.User
	if err := s.dbExecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		users, repoErr = s.repo.ListByRoles(ctx, db, roles)
		return repoErr
	}); err != nil {
		return nil, s.handleError(err, "Failed to list staff accounts")
	}

	sessions := make([]StaffSession, 0)
	for _, u := range users {
		userSessions, err := s.sessions.GetUserSessions(ctx, u.ID)
		if err != nil {
			s.logger.Warn("failed to list sessions for staff user", log.Meta{
				"userId": u.ID.String(),
				"error":  err.Error(),
			})
			continue
		}
		for _, sess := range userSessions {
			sessions = append(sessions, StaffSession{
				ID:         sess.ID,
				UserID:     sess.UserID,
				Email:      u.Identifier,
				FullName:   u.FullName,
				Role:       string(u.Role),
				UserAgent:  sess.UserAgent,
				IPAddress:  sess.IPAddress,
				CreatedAt:  sess.CreatedAt,
				LastActive: sess.LastActive,
				Current:    sess.ID == currentSessionID,
			})
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActive.After(sessions[j].LastActive)
	})

	return sessions, nil
}

// RevokeStaffSession terminates a single active session by ID on behalf of an
// administrator. A caller may revoke their own other devices, and may revoke a
// lower-ranked user's session; equal or higher rank is refused to prevent one
// admin from locking out a peer or superadmin.
func (s *UserService) RevokeStaffSession(
	ctx context.Context,
	sessionID string,
	callerID uuid.UUID,
) error {
	if strings.TrimSpace(sessionID) == "" {
		return apierr.BadRequestError("Session ID is required").WithCode(apierr.CodeInvalidInput)
	}
	if s.sessions == nil {
		return apierr.InternalServerError("Session store unavailable")
	}

	sess, err := s.sessions.GetSessionByID(ctx, sessionID)
	if err != nil {
		var ae *apierr.APIError
		if errors.As(err, &ae) && ae.Status == http.StatusUnauthorized {
			return apierr.NotFoundError("Session not found").WithCode(apierr.CodeNotFound)
		}
		return apierr.InternalServerError("Failed to load session").Wrap(err)
	}

	caller, err := s.GetUserByID(ctx, callerID)
	if err != nil {
		return apierr.UnauthorizedError("Caller not found")
	}

	if sess.UserID != callerID {
		target, err := s.GetUserByID(ctx, sess.UserID)
		if err != nil {
			return err
		}
		if GetRoleRank(caller.Role) <= GetRoleRank(target.Role) {
			return apierr.ForbiddenError("You cannot revoke sessions of a user with equal or higher privileges").
				WithCode("CANNOT_MODIFY_EQUAL_OR_HIGHER_ROLE")
		}
	}

	if err := s.sessions.DeleteSessionByID(ctx, sess.UserID, sessionID); err != nil {
		var ae *apierr.APIError
		if errors.As(err, &ae) {
			return err
		}
		return apierr.InternalServerError("Failed to revoke session").Wrap(err)
	}

	s.logger.Info("staff session revoked", log.Meta{
		"sessionId": sessionID,
		"userId":    sess.UserID.String(),
		"revokedBy": callerID.String(),
	})
	return nil
}
