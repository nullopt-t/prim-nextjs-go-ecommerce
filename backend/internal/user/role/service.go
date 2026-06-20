package role

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

var (
	ErrRoleNotFound = errors.New("role not found")
)

type RoleService struct {
	roleRepo   *RoleRepository
	dbExecuter database.Runner
}

func NewRoleService(
	dbExecutor database.Runner,
	roleRepo *RoleRepository,
) *RoleService {
	return &RoleService{
		roleRepo:   roleRepo,
		dbExecuter: dbExecutor,
	}
}

func (s *RoleService) GetRoleByID(
	ctx context.Context,
	roleID int,
) (*model.Role, error) {
	var role *model.Role
	err := s.dbExecuter.WithDB(ctx,
		func(pool database.QueryExecutor) error {
			r, err := s.roleRepo.GetByID(ctx, pool, roleID)
			if err != nil {
				return err
			}
			role = r
			return nil
		},
	)
	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(
			mappedErr,
			database.ErrNotFound,
		):
			return nil, nil
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to get a role",
				err,
			)
		}
	}
	return role, nil
}

func (s *RoleService) GetRoleByCode(
	ctx context.Context,
	code model.RoleCode,
) (*model.Role, error) {
	var role *model.Role
	err := s.dbExecuter.WithDB(ctx,
		func(pool database.QueryExecutor) error {
			r, err := s.roleRepo.GetByCode(ctx, pool, code)
			if err != nil {
				return err
			}
			role = r
			return nil
		},
	)
	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(
			mappedErr,
			database.ErrNotFound,
		):
			return nil, nil
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to get a role",
				err,
			)
		}
	}
	return role, nil
}

func (s *RoleService) GetAll(
	ctx context.Context,
	query api.PageQuery,
) ([]*model.Role, *api.Page, error) {
	var roles []*model.Role
	var page *api.Page
	err := s.dbExecuter.WithDB(ctx,
		func(pool database.QueryExecutor) error {
			r, p, err := s.roleRepo.List(
				ctx,
				pool,
				query,
			)
			if err != nil {
				return err
			}
			roles = r
			page = p
			return nil
		},
	)
	return roles, page, err
}

func (s *RoleService) Assign(
	ctx context.Context,
	userID uuid.UUID,
	roleID int,
) error {
	ur := &model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return s.dbExecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return s.roleRepo.Assign(
				ctx,
				db,
				ur,
			)
		},
	)
}

func (s *RoleService) Revoke(
	ctx context.Context,
	userID uuid.UUID,
	roleID int,
) error {
	ur := &model.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return s.dbExecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return s.roleRepo.Revoke(
				ctx,
				db,
				ur,
			)
		},
	)
}

func (s *RoleService) HasRole(
	ctx context.Context,
	userID uuid.UUID,
	code model.RoleCode,
) error {

	exists := false
	err := s.dbExecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			e, err := s.roleRepo.HasRole(
				ctx,
				db,
				userID,
				code,
			)
			exists = e
			return err
		},
	)

	if err != nil {
		return err
	}

	if !exists {
		return ErrRoleNotFound
	}

	return nil
}
