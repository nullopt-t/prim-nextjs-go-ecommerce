package user

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// CreateAddressInput represents inbound payload parameters required to create a new address record.
type CreateAddressInput struct {
	Label      *string `json:"label"`
	FullName   string  `json:"fullName"`
	Phone      *string `json:"phone"`
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2"`
	City       string  `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postalCode"`
	Country    string  `json:"country"`
	IsDefault  bool    `json:"isDefault"`
}

// UpdateAddressInput represents inbound payload parameters required to update an existing address record.
type UpdateAddressInput struct {
	Label      *string `json:"label"`
	FullName   string  `json:"fullName"`
	Phone      *string `json:"phone"`
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2"`
	City       string  `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postalCode"`
	Country    string  `json:"country"`
	IsDefault  bool    `json:"isDefault"`
}

// AddressService orchestrates business logic for customer addresses:
// input validation, database transactions, default address state machine transitions,
// automatic default address promotion upon deletion, and tenant-scoped security.
type AddressService struct {
	dbRunner database.Runner
	repo     *AddressRepository
	logger   log.Logger
}

// NewAddressService constructs a new AddressService instance.
func NewAddressService(
	dbRunner database.Runner,
	repo *AddressRepository,
	logger log.Logger,
) *AddressService {
	return &AddressService{
		dbRunner: dbRunner,
		repo:     repo,
		logger:   logger,
	}
}

// ListAddresses returns all non-deleted saved addresses for the authenticated user, ordered default-first.
func (s *AddressService) ListAddresses(ctx context.Context, userID uuid.UUID) ([]model.UserAddress, error) {
	var addresses []model.UserAddress
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		addresses, err = s.repo.ListByUser(ctx, db, userID)
		return err
	})
	if err != nil {
		s.logger.Error("failed to list user addresses", log.Meta{"userId": userID, "error": err.Error()})
		return nil, apierr.InternalServerError("Failed to list addresses").Wrap(err)
	}
	return addresses, nil
}

// GetAddress retrieves a single address by its UUID, verifying it belongs to the authenticated user.
func (s *AddressService) GetAddress(ctx context.Context, id, userID uuid.UUID) (*model.UserAddress, error) {
	var address *model.UserAddress
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		address, err = s.repo.GetByIDAndUser(ctx, db, id, userID)
		return err
	})
	if err != nil {
		if errors.Is(err, ErrAddressNotFound) {
			return nil, apierr.NotFoundError("Address not found")
		}
		s.logger.Error("failed to get user address", log.Meta{"id": id, "userId": userID, "error": err.Error()})
		return nil, apierr.InternalServerError("Failed to retrieve address").Wrap(err)
	}
	return address, nil
}

// CreateAddress validates address input fields and persists a new address inside a transaction.
// If this is the user's first address, it is automatically promoted to the default address.
// If the user marked it as default, existing default addresses are atomically cleared.
func (s *AddressService) CreateAddress(ctx context.Context, userID uuid.UUID, in *CreateAddressInput) (*model.UserAddress, error) {
	// Validate required address fields are non-empty after trimming whitespace
	if strings.TrimSpace(in.FullName) == "" {
		return nil, apierr.BadRequestError("Full name is required").WithCode(apierr.CodeInvalidInput)
	}
	if strings.TrimSpace(in.Line1) == "" {
		return nil, apierr.BadRequestError("Address line 1 is required").WithCode(apierr.CodeInvalidInput)
	}
	if strings.TrimSpace(in.City) == "" {
		return nil, apierr.BadRequestError("City is required").WithCode(apierr.CodeInvalidInput)
	}
	if strings.TrimSpace(in.Country) == "" {
		return nil, apierr.BadRequestError("Country is required").WithCode(apierr.CodeInvalidInput)
	}

	address := &model.UserAddress{
		ID:         uuid.New(),
		UserID:     userID,
		Label:      in.Label,
		FullName:   strings.TrimSpace(in.FullName),
		Phone:      in.Phone,
		Line1:      strings.TrimSpace(in.Line1),
		Line2:      in.Line2,
		City:       strings.TrimSpace(in.City),
		State:      in.State,
		PostalCode: in.PostalCode,
		Country:    strings.TrimSpace(in.Country),
		IsDefault:  in.IsDefault,
	}

	// Execute transactional insertion and default flag reconciliation
	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		// Check existing active address count
		count, err := s.repo.CountByUser(ctx, tx, userID)
		if err != nil {
			return err
		}

		// First address created is unconditionally set as the default
		if count == 0 {
			address.IsDefault = true
		} else if address.IsDefault {
			// Atomically clear any existing default address
			if err := s.repo.ClearDefault(ctx, tx, userID); err != nil {
				return err
			}
		}

		return s.repo.Create(ctx, tx, address)
	})

	if err != nil {
		s.logger.Error("failed to create user address", log.Meta{"userId": userID, "error": err.Error()})
		return nil, apierr.InternalServerError("Failed to create address").Wrap(err)
	}

	return address, nil
}

// UpdateAddress updates attributes of an existing address within an atomic database transaction.
func (s *AddressService) UpdateAddress(ctx context.Context, id, userID uuid.UUID, in *UpdateAddressInput) (*model.UserAddress, error) {
	// Validate required fields
	if strings.TrimSpace(in.FullName) == "" {
		return nil, apierr.BadRequestError("Full name is required").WithCode(apierr.CodeInvalidInput)
	}
	if strings.TrimSpace(in.Line1) == "" {
		return nil, apierr.BadRequestError("Address line 1 is required").WithCode(apierr.CodeInvalidInput)
	}
	if strings.TrimSpace(in.City) == "" {
		return nil, apierr.BadRequestError("City is required").WithCode(apierr.CodeInvalidInput)
	}
	if strings.TrimSpace(in.Country) == "" {
		return nil, apierr.BadRequestError("Country is required").WithCode(apierr.CodeInvalidInput)
	}

	var updated *model.UserAddress
	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		current, err := s.repo.GetByIDAndUser(ctx, tx, id, userID)
		if err != nil {
			return err
		}

		current.Label = in.Label
		current.FullName = strings.TrimSpace(in.FullName)
		current.Phone = in.Phone
		current.Line1 = strings.TrimSpace(in.Line1)
		current.Line2 = in.Line2
		current.City = strings.TrimSpace(in.City)
		current.State = in.State
		current.PostalCode = in.PostalCode
		current.Country = strings.TrimSpace(in.Country)

		// Transitioning address to default: clear existing default and set this address
		if in.IsDefault && !current.IsDefault {
			if err := s.repo.ClearDefault(ctx, tx, userID); err != nil {
				return err
			}
			current.IsDefault = true
		} else if !in.IsDefault && current.IsDefault {
			// Prevent unsetting isDefault if it leaves zero default addresses
			count, err := s.repo.CountByUser(ctx, tx, userID)
			if err != nil {
				return err
			}
			if count <= 1 {
				return apierr.BadRequestError("Cannot unset default on the only address").WithCode(apierr.CodeInvalidInput)
			}
			// When un-defaulting this address, find another active address and promote it
			current.IsDefault = false
			if err := s.repo.Update(ctx, tx, current); err != nil {
				return err
			}
			next, err := s.repo.FindMostRecentActive(ctx, tx, userID)
			if err == nil && next != nil && next.ID != current.ID {
				if err := s.repo.SetDefault(ctx, tx, next.ID, userID); err != nil {
					return err
				}
			}
		}

		if err := s.repo.Update(ctx, tx, current); err != nil {
			return err
		}
		updated = current
		return nil
	})

	if err != nil {
		if errors.Is(err, ErrAddressNotFound) {
			return nil, apierr.NotFoundError("Address not found")
		}
		s.logger.Error("failed to update user address", log.Meta{"id": id, "userId": userID, "error": err.Error()})
		return nil, apierr.InternalServerError("Failed to update address").Wrap(err)
	}

	return updated, nil
}

// SetDefaultAddress marks the specified address as the default address within an atomic transaction.
func (s *AddressService) SetDefaultAddress(ctx context.Context, id, userID uuid.UUID) (*model.UserAddress, error) {
	var address *model.UserAddress
	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		curr, err := s.repo.GetByIDAndUser(ctx, tx, id, userID)
		if err != nil {
			return err
		}
		// If already the default address, no mutation is required
		if curr.IsDefault {
			address = curr
			return nil
		}

		// Atomically unset previous default and set new default
		if err := s.repo.ClearDefault(ctx, tx, userID); err != nil {
			return err
		}
		if err := s.repo.SetDefault(ctx, tx, id, userID); err != nil {
			return err
		}
		curr.IsDefault = true
		address = curr
		return nil
	})

	if err != nil {
		if errors.Is(err, ErrAddressNotFound) {
			return nil, apierr.NotFoundError("Address not found")
		}
		s.logger.Error("failed to set default address", log.Meta{"id": id, "userId": userID, "error": err.Error()})
		return nil, apierr.InternalServerError("Failed to set default address").Wrap(err)
	}

	return address, nil
}

// DeleteAddress soft-deletes the target address.
// If the deleted address was designated as the default address, it queries the most recently updated
// active address and automatically promotes it as the new default address.
func (s *AddressService) DeleteAddress(ctx context.Context, id, userID uuid.UUID) error {
	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		current, err := s.repo.GetByIDAndUser(ctx, tx, id, userID)
		if err != nil {
			return err
		}

		wasDefault := current.IsDefault

		// Soft-delete the address
		if err := s.repo.SoftDelete(ctx, tx, id, userID); err != nil {
			return err
		}

		// If the deleted address was default, promote the next most recent active address
		if wasDefault {
			next, err := s.repo.FindMostRecentActive(ctx, tx, userID)
			if err == nil && next != nil {
				if err := s.repo.SetDefault(ctx, tx, next.ID, userID); err != nil {
					return err
				}
			}
			// If no remaining address found (ErrAddressNotFound), that's completely normal for the user's last address
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrAddressNotFound) {
			return apierr.NotFoundError("Address not found")
		}
		s.logger.Error("failed to delete user address", log.Meta{"id": id, "userId": userID, "error": err.Error()})
		return apierr.InternalServerError("Failed to delete address").Wrap(err)
	}

	return nil
}
