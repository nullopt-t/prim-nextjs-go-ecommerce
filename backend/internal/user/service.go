package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
			return nil, nil
		default:
			return nil, security.NewSecureError(
				http.StatusInternalServerError,
				security.CodeInternal,
				"failed to get a user",
				err,
			)
		}
	}
	return user, nil
}

func (s *UserService) CreateUser(
	ctx context.Context,
	email string,
	password string,
) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to hash password",
			err,
		)
	}
	dummyUser := model.NewUser(
		email,
		hashedPassword,
	)
	err = s.dbExecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			err = s.repo.Create(
				ctx,
				db,
				dummyUser,
			)
			if err != nil {
				mappedErr := database.MapError(err)
				switch {
				case errors.Is(
					mappedErr,
					database.ErrConflict,
				):
					return security.NewSecureError(
						http.StatusConflict,
						security.CodeConflict,
						"user already exists",
						err,
					)
				default:
					return security.NewSecureError(
						http.StatusInternalServerError,
						security.CodeInternal,
						"failed to create a new user",
						err,
					)
				}
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return dummyUser, nil
}

func (s *UserService) Login(
	ctx context.Context,
	email string,
	password string,
) (*model.User, error) {
	var user *model.User
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			user, err = s.get(
				ctx,
				db,
				Filter{
					Email: &email,
				},
			)
			if err != nil {
				return err
			}
			if user == nil {
				return security.NewSecureError(
					http.StatusUnauthorized,
					security.CodeUnauthorized,
					"invalid credentials",
					nil,
				)
			}
			if err := bcrypt.CompareHashAndPassword(
				[]byte(user.PasswordHash),
				[]byte(password),
			); err != nil {
				return security.NewSecureError(
					http.StatusUnauthorized,
					security.CodeUnauthorized,
					"invalid credentials",
					nil,
				)
			}
			return nil
		})
	return user, err
}

func (s *UserService) MarkEmailVerified(
	ctx context.Context,
	email string,
) error {
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			user, err := s.repo.Get(ctx, db, Filter{
				Email: &email,
			})
			if err != nil {
				return err
			}

			if user.EmailVerifiedAt != nil {
				s.logger.Info(
					"email is already verified",
					log.Meta{
						"userID": user.ID,
						"email":  user.Email,
					})
				return nil
			}

			err = s.repo.VerifyIdentifier(
				ctx,
				db,
				IdentifierTypeEmail,
				email,
			)
			if err != nil {
				return err
			}
			return nil
		})
	return err
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

func (s *UserService) GetUserByEmail(
	ctx context.Context,
	email string,
) (*model.User, error) {
	var user *model.User
	err := s.dbExecuter.WithDB(ctx,
		func(db database.QueryExecutor) error {
			var err error
			user, err = s.get(
				ctx,
				db,
				Filter{
					Email: &email,
				},
			)
			if err != nil {
				return err
			}
			return nil
		})
	return user, err
}

func (s *UserService) GetAllUsers(
	ctx context.Context,
	q api.PageQuery,
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
				return security.NewSecureError(
					http.StatusInternalServerError,
					security.CodeInternal,
					"failed to fetch users",
					err,
				)
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

func (s *UserService) UpdateUserPassword(
	ctx context.Context,
	user *model.User,
	newPassword string,
) error {
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to process the credentials",
			nil,
		)
	}

	err = s.dbExecuter.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return s.repo.UpdatePasswordHash(
				ctx,
				db,
				Filter{
					ID: &user.ID,
				},
				hashedPassword,
			)
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to update password",
			err,
		)
	}
	return nil
}
