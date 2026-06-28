package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/crypto"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
	"github.com/redis/go-redis/v9"
)

const (
	ChallengeTTL   = 5 * time.Minute
	MaxResendTimes = 3
)

type CreateChallengeRequest struct {
	Identifier string
}

type ChallengeService struct {
	redisClient *redis.Client
	otpGen      *OTPGenerator
	notifier    Notifier
	logger      log.Logger
}

func NewChallengeService(
	rdc *redis.Client,
	otpGen *OTPGenerator,
	notifier Notifier,
	logger log.Logger,
) *ChallengeService {
	return &ChallengeService{
		redisClient: rdc,
		notifier:    notifier,
		otpGen:      otpGen,
		logger:      logger,
	}
}

func key(identifier string) string {
	return fmt.Sprintf("challenge:%s", identifier)
}

func (cs *ChallengeService) Create(
	ctx context.Context,
	identifier string,
	channel string,
) (*model.Challenge, error) {
	// generate a new otp code
	otp, err := cs.otpGen.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("create challenge:%w", err)
	}

	// hash the new otp code
	otpHash, err := crypto.Hash(otp)
	if err != nil {
		return nil, fmt.Errorf("create challenge:%w", err)
	}

	// at first check if there is an existed old challenge
	challenge, err := cs.Get(
		ctx,
		identifier,
	)
	if err != nil {
		// validate the identifier

		// create a new challenge instance
		challenge = model.NewChallenge(
			identifier,
			channel,
			otpHash,
			ChallengeTTL,
		)

		key := key(challenge.Identifier)

		// start a new redis pipeline
		pipe := cs.redisClient.Pipeline()
		pipe.HSet(
			ctx,
			key,
			map[string]any{
				"id":           challenge.ID,
				"identifier":   challenge.Identifier,
				"channel":      challenge.Channel,
				"otp_hash":     challenge.OtpHash,
				"status":       challenge.Status,
				"attempts":     challenge.Attempts,
				"resend_count": challenge.ResendCount,
				"expires_at":   challenge.ExpiresAt,
				"created_at":   challenge.CreatedAt,
			},
		)

		pipe.Expire(
			ctx,
			key,
			ChallengeTTL,
		)

		// execute the pipeline and make sure it's successfully done
		_, err = pipe.Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("create challenge:%w", err)
		}

	} else {
		if challenge.ResendCount >= MaxResendTimes {
			return nil, security.NewSecureError(
				http.StatusTooManyRequests,
				security.CodeRateLimit,
				"rate limit exceeded",
				nil,
			)
		}
		err = cs.updateOTP(
			ctx,
			challenge.Identifier,
			otpHash,
		)
		if err != nil {
			return nil, err
		}

		_, err := cs.redisClient.HIncrBy(
			ctx,
			key(challenge.Identifier),
			"resend_count",
			1,
		).Result()
		if err != nil {
			return nil, fmt.Errorf("create challenge: %w", err)
		}
	}

	// send the otp code to the user identifier
	err = cs.notifier.NotifyOTP(
		ctx,
		challenge.Channel,
		challenge.Identifier,
		otp,
	)
	if err != nil {
		cs.logger.Warn(
			"failed to send otp code",
			log.Meta{
				"Error": err,
			},
		)
	}
	return challenge, nil
}

func (cs *ChallengeService) Get(
	ctx context.Context,
	identifier string,
) (*model.Challenge, error) {
	val, err := cs.redisClient.HGetAll(
		ctx,
		key(identifier),
	).Result()
	if err != nil {
		return nil, fmt.Errorf("get challenge:%w", err)
	}

	if len(val) == 0 {
		return nil, security.NewSecureError(
			http.StatusUnauthorized,
			security.CodeInvalidOrExpired,
			"invalid or expired challenge, please request a new challenge",
			nil,
		)
	}

	expiresAt, err := time.Parse(time.RFC3339, val["expires_at"])
	if err != nil {
		return nil, fmt.Errorf("get challenge:%w", err)
	}

	challenge := &model.Challenge{
		ID:          val["id"],
		Identifier:  val["identifier"],
		Channel:     val["channel"],
		OtpHash:     val["otp_hash"],
		ResendCount: utils.StringToInt(val["resend_count"], 0),
		Attempts:    utils.StringToInt(val["attempts"], 0),
		Status:      val["status"],
		ExpiresAt:   expiresAt,
	}

	return challenge, nil
}

func (cs *ChallengeService) updateOTP(
	ctx context.Context,
	identifier string,
	otp string,
) error {
	otpHash, err := crypto.Hash(otp)
	if err != nil {
		return fmt.Errorf("update challenge otp:%w", err)
	}

	pipe := cs.redisClient.Pipeline()
	pipe.HSet(
		ctx,
		key(identifier),
		map[string]any{
			"otp_hash":   otpHash,
			"expires_at": time.Now().Add(ChallengeTTL),
		},
	)
	pipe.Expire(
		ctx,
		key(identifier),
		ChallengeTTL,
	)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("update challenge otp: %w", err)
	}
	return nil
}

func (cs *ChallengeService) Resend(
	ctx context.Context,
	challenge *model.Challenge,
) error {
	if challenge.ResendCount >= MaxResendTimes {
		return security.NewSecureError(
			http.StatusTooManyRequests,
			security.CodeRateLimit,
			"rate limit exceeded",
			nil,
		)
	}

	if challenge.Status != "pending" {
		return security.NewSecureError(
			http.StatusGone,
			security.CodeExpired,
			"challenge expired",
			nil,
		)
	}

	if time.Now().After(challenge.ExpiresAt) {
		return security.NewSecureError(
			http.StatusGone,
			security.CodeExpired,
			"challenge expired",
			nil,
		)
	}

	newOtp, err := cs.otpGen.GenerateOTP()
	if err != nil {
		return fmt.Errorf("resend challenge:%w", err)
	}

	if err := cs.updateOTP(ctx, challenge.ID, newOtp); err != nil {
		return err
	}

	_, err = cs.redisClient.HIncrBy(
		ctx,
		key(challenge.Identifier),
		"resend_count",
		1,
	).Result()
	if err != nil {
		return fmt.Errorf("resend challenge:%w", err)
	}

	// send the otp code to the user identifier
	err = cs.notifier.NotifyOTP(
		ctx,
		challenge.Channel,
		challenge.Identifier,
		newOtp,
	)
	if err != nil {
		cs.logger.Warn(
			"failed to send otp code",
			log.Meta{
				"Error": err,
			},
		)
	}
	return nil
}

func (cs *ChallengeService) Verify(
	ctx context.Context,
	challenge *model.Challenge,
	otp string,
) (bool, error) {
	if challenge.Status == "verified" {
		return false, security.NewSecureError(
			http.StatusConflict,
			security.CodeConflict,
			"already verified",
			nil,
		)
	}

	if challenge.Status == "expired" {
		return false, security.NewSecureError(
			http.StatusGone,
			security.CodeExpired,
			"challenge expired",
			nil,
		)
	}

	if time.Now().After(challenge.ExpiresAt) {
		return false, security.NewSecureError(
			http.StatusGone,
			security.CodeExpired,
			"challenge expired",
			nil,
		)
	}

	ok, err := crypto.Equal(challenge.OtpHash, otp)
	if err != nil {
		return false, fmt.Errorf("challenge verify:%w", err)
	}

	if !ok {
		return false, nil
	}

	return true, nil
}

func (cs *ChallengeService) MarkVerified(
	ctx context.Context,
	challenge *model.Challenge,
) error {
	err := cs.redisClient.HSet(
		ctx,
		key(challenge.Identifier),
		"status",
		"verified",
	).Err()
	if err != nil {
		return fmt.Errorf(
			"mark challenge verified :%w",
			err,
		)
	}
	return err
}

func (cs *ChallengeService) Expire(
	ctx context.Context,
	challenge *model.Challenge,
) error {
	err := cs.redisClient.HSet(
		ctx,
		key(challenge.Identifier),
		"status",
		"expired",
	).Err()
	if err != nil {
		return fmt.Errorf(
			"mark challenge verified :%w",
			err,
		)
	}
	return err
}
