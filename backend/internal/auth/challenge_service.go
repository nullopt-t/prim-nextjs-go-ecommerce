package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/crypto"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
	"github.com/redis/go-redis/v9"
)

const (
	MaxResendTimes    = 3
	MaxVerifyAttempts = 5
)

type CreateChallengeRequest struct {
	Identifier string
}

type ChallengeService struct {
	redisClient  *redis.Client
	notifier     Notifier
	logger       log.Logger
	challengeTTL time.Duration
}

func NewChallengeService(
	rdc *redis.Client,
	notifier Notifier,
	logger log.Logger,
	challengeTTL time.Duration,
) *ChallengeService {
	return &ChallengeService{
		redisClient:  rdc,
		notifier:     notifier,
		logger:       logger,
		challengeTTL: challengeTTL,
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
	otp, err := GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("create challenge: %w", err)
	}
	otpHash, err := crypto.Hash(otp)
	if err != nil {
		return nil, fmt.Errorf("create challenge: %w", err)
	}

	k := key(identifier)

	existing, err := cs.Get(ctx, identifier)

	var challenge *model.Challenge
	isNew := false

	switch {
	case err != nil:
		// No existing challenge.
		challenge = model.NewChallenge(identifier, channel, otpHash, cs.challengeTTL)
		isNew = true

	case existing.Status == "pending":
		if existing.ResendCount >= MaxResendTimes {
			return nil, apierr.New(
				http.StatusTooManyRequests,
				"Too many resend attempts",
			).WithCode(apierr.CodeRateLimitExceeded)
		}
		existing.OtpHash = otpHash
		existing.ResendCount++
		challenge = existing

	case existing.Status == "verified" || existing.Status == "expired":
		challenge = model.NewChallenge(identifier, channel, otpHash, cs.challengeTTL)
		isNew = true

	default:
		return nil, fmt.Errorf("unknown challenge status: %q", existing.Status)
	}

	// Persist in all cases — new challenge, resend, or reset after
	// verified/expired. Always refresh TTL so a resend can't expire
	// mid-flow.
	pipe := cs.redisClient.Pipeline()
	pipe.HSet(ctx, k, map[string]any{
		"id":           challenge.ID,
		"identifier":   challenge.Identifier,
		"channel":      challenge.Channel,
		"otp_hash":     challenge.OtpHash,
		"status":       challenge.Status,
		"attempts":     challenge.Attempts,
		"resend_count": challenge.ResendCount,
		"expires_at":   challenge.ExpiresAt,
		"created_at":   challenge.CreatedAt,
	})
	pipe.Expire(ctx, k, cs.challengeTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("create challenge: persist: %w", err)
	}

	// Send the OTP. A failed send is NOT a successful challenge creation —
	// don't burn the user's resend budget on a code they never got, and
	// don't tell the caller everything is fine.
	if err := cs.notifier.NotifyOTP(ctx, challenge.Channel, challenge.Identifier, otp); err != nil {
		cs.logger.Warn("failed to send otp code", log.Meta{"Error": err})

		if !isNew {
			// Roll back the resend-count increment so the failed send
			// doesn't cost the user an attempt.
			if _, rbErr := cs.redisClient.HIncrBy(ctx, k, "resend_count", -1).Result(); rbErr != nil {
				cs.logger.Warn("failed to roll back resend_count", log.Meta{"Error": rbErr})
			}
		}
		return nil, fmt.Errorf("create challenge: notify: %w", err)
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
		return nil, apierr.New(
			http.StatusUnauthorized,
			"Challenge not found or expired",
		).WithCode(apierr.CodeUnauthorized)
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
			"expires_at": time.Now().Add(cs.challengeTTL),
		},
	)
	pipe.Expire(
		ctx,
		key(identifier),
		cs.challengeTTL,
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
		return apierr.New(
			http.StatusTooManyRequests,
			"Too many resend attempts",
		).WithCode(apierr.CodeRateLimitExceeded)
	}

	if challenge.Status != "pending" {
		return apierr.New(
			http.StatusGone,
			"Challenge expired or invalid",
		).WithCode(apierr.CodeExpired)
	}

	if time.Now().After(challenge.ExpiresAt) {
		return apierr.New(
			http.StatusGone,
			"Challenge expired",
		).WithCode(apierr.CodeExpired)
	}

	newOtp, err := GenerateOTP()
	if err != nil {
		return fmt.Errorf("resend challenge:%w", err)
	}

	if err := cs.updateOTP(ctx, challenge.Identifier, newOtp); err != nil {
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
		return false, apierr.New(
			http.StatusConflict,
			"Challenge already verified",
		).WithCode(apierr.CodeResourceConflict)
	}

	if challenge.Status == "expired" {
		return false, apierr.New(
			http.StatusGone,
			"Challenge expired",
		).WithCode(apierr.CodeExpired)
	}

	if time.Now().After(challenge.ExpiresAt) {
		return false, apierr.New(
			http.StatusGone,
			"Challenge expired",
		).WithCode(apierr.CodeExpired)
	}

	if challenge.Attempts >= MaxVerifyAttempts {
		return false, apierr.New(
			http.StatusTooManyRequests,
			"Too many verification attempts",
		).WithCode(apierr.CodeRateLimitExceeded)
	}

	ok, err := crypto.Equal(challenge.OtpHash, otp)
	if err != nil {
		return false, fmt.Errorf("challenge verify:%w", err)
	}

	if !ok {
		// Increment attempts asynchronously
		cs.redisClient.HIncrBy(ctx, key(challenge.Identifier), "attempts", 1)
		return false, nil
	}

	return true, nil
}

const markVerifiedScript = `
local status = redis.call("HGET", KEYS[1], "status")
if status == "verified" then
    return -1
end
if status ~= "pending" then
    return -2
end
redis.call("HSET", KEYS[1], "status", "verified")
return 1
`

func (cs *ChallengeService) MarkVerified(
	ctx context.Context,
	challenge *model.Challenge,
) error {
	result, err := cs.redisClient.Eval(
		ctx,
		markVerifiedScript,
		[]string{key(challenge.Identifier)},
	).Result()
	
	if err != nil {
		return fmt.Errorf("mark challenge verified: eval script: %w", err)
	}
	
	resNum, _ := result.(int64)
	if resNum == -1 {
		return apierr.New(http.StatusConflict, "Challenge already verified").WithCode(apierr.CodeResourceConflict)
	} else if resNum == -2 {
		return apierr.New(http.StatusGone, "Challenge expired or invalid").WithCode(apierr.CodeExpired)
	}
	
	return nil
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
			"mark challenge expired :%w",
			err,
		)
	}
	return err
}
