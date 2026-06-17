package otp

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/crypto"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"context"
	"fmt"
	"net/http"
	"time"
)

const OTPTTL = 15 * time.Minute

type OTPPurpose string

const (
	OTPTypeRegister OTPPurpose = "register"
	OTPTypeEmail    OTPPurpose = "email"
	OTPTypeLogin    OTPPurpose = "login"
	OTPTypeReset    OTPPurpose = "reset"
)

type Notifier interface {
	NotifyOTP(
		ctx context.Context,
		to,
		otp string,
	) error
}

type Sender interface {
	SendSMS(
		ctx context.Context,
		to,
		otp string,
	) error
}

type Generator interface {
	GenerateOTP(
		purpose string,
	) (string, error)
}

type Store interface {
	Save(
		ctx context.Context,
		key string,
		otp OTP,
		ttl time.Duration,
	) error

	Get(
		ctx context.Context,
		key string,
	) (*OTP, error)

	Increment(
		ctx context.Context,
		key string,
	) error

	Delete(
		ctx context.Context,
		key string,
	) error
}

type Service struct {
	store    Store
	gen      Generator
	logger   log.Logger
	notifier Notifier
}

func NewService(
	store Store,
	gen Generator,
	logger log.Logger,
	notifier Notifier,
) *Service {
	return &Service{
		store:    store,
		gen:      gen,
		logger:   logger,
		notifier: notifier,
	}
}

func invalidOTP() error {
	return security.NewSecureError(
		http.StatusUnauthorized,
		"EXPIRED_OR_INVALID_OTP",
		"expired or invalid OTP",
		nil,
	)
}

func (s *Service) genKey(purpose string, suffix string) string {
	return fmt.Sprintf("%s:%s",
		purpose,
		suffix,
	)
}

func (s *Service) GenerateOTP(
	ctx context.Context,
	purpose string,
	suffix string,
) (string, error) {
	otp, err := s.gen.GenerateOTP(purpose)
	if err != nil {
		return "", err
	}

	hashedOTP, err := crypto.Hash(otp)
	if err != nil {
		return "", err
	}

	key := s.genKey(purpose, suffix)
	o := OTP{
		CodeHash: hashedOTP,
		Attempts: 0,
	}

	if err := s.store.Save(
		ctx,
		key,
		o,
		OTPTTL,
	); err != nil {
		return "", err
	}
	return otp, nil
}

func (s *Service) SendOTP(
	ctx context.Context,
	email string,
	purpose string,
) error {
	otp, err := s.GenerateOTP(
		ctx,
		purpose,
		email,
	)
	if err != nil {
		return err
	}

	switch purpose {
	case "register":
		err = s.notifier.NotifyOTP(
			ctx,
			email,
			otp,
		)
	default:
		return fmt.Errorf("unsupported purpose: %s", purpose)
	}

	return err
}

func (s *Service) incrementAttempts(ctx context.Context, key string) error {
	err := s.store.Increment(
		ctx,
		key,
	)
	return err
}

func (s *Service) VerifyOTP(
	ctx context.Context,
	email string,
	purpose string,
	otp string,
) error {
	var key string
	key = fmt.Sprintf("%s:%s",
		purpose,
		email,
	)

	s.logger.Debug(
		"fetching otp hash....",
		log.Meta{"key": key},
	)

	uo, err := s.store.Get(
		ctx,
		key,
	)

	if err != nil {
		s.logger.Error(
			"failed to fetch otp hash",
			log.Meta{"key": key},
		)
		return security.NewSecureError(
			http.StatusUnauthorized,
			"EXPIRED_OR_INVALID_OTP",
			"expired or invalid OTP",
			err,
		)
	}
	s.logger.Debug(
		"otp hash fetched successfully",
		log.Meta{"key": key},
	)

	s.logger.Debug(
		"comparing otp hash",
		log.Meta{
			"key": key,
		},
	)

	hashedOTP, err := crypto.Hash(otp)
	if err != nil {
		return err
	}

	if uo.CodeHash != hashedOTP {
		err = s.incrementAttempts(ctx, key)
		if err != nil {
			s.logger.Error(
				"failed to increment attempts",
				log.Meta{"key": key},
			)
		}
		s.logger.Error(
			"invalid otp",
			log.Meta{"key": key},
		)
		return invalidOTP()
	}

	s.logger.Debug(
		"deleting otp after successful verification",
		log.Meta{"key": key},
	)
	err = s.store.Delete(ctx, key)
	if err != nil {
		s.logger.Warn(
			"failed to delete otp",
			log.Meta{
				"key": key,
			},
		)
	}
	s.logger.Info(
		"otp deleted successfully",
		log.Meta{"key": key},
	)
	return nil
}

func (s *Service) ResendOTP(
	ctx context.Context,
	email string,
	purpose string,
) error {
	otp, err := s.gen.GenerateOTP(purpose)
	if err != nil {
		return err
	}

	hashedOTP, err := crypto.Hash(otp)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s",
		purpose,
		email,
	)
	err = s.store.Save(
		ctx,
		key,
		OTP{
			CodeHash: hashedOTP,
		},
		OTPTTL,
	)
	if err != nil {
		return err
	}

	s.logger.Info(
		"otp saved successfully",
		log.Meta{"key": key},
	)

	s.logger.Info(
		"Sending OTP",
		log.Meta{
			"purpose": purpose,
			"email":   email,
			"otp":     otp,
		},
	)
	return err
}
