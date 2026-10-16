package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/crypto"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/utils"
	"github.com/redis/go-redis/v9"
)

const (
	// MaxResendTimes specifies the maximum number of times an OTP code may be resent for a single challenge flow.
	MaxResendTimes = 3
	// MaxVerifyAttempts specifies the maximum number of incorrect OTP verification attempts allowed before locking.
	MaxVerifyAttempts = 5
)

// CreateChallengeRequest defines parameters for programmatic challenge creation.
type CreateChallengeRequest struct {
	Identifier string
}

// ChallengeService manages the lifecycle of passwordless OTP authentication challenges in Redis:
// generation, cryptographic Argon2/Bcrypt/HMAC hashing of one-time secrets, rate-limiting,
// attempt tracking, resend policies, atomic verification transitions, and expiration cleanup.
type ChallengeService struct {
	redisClient  *redis.Client
	notifier     Notifier
	logger       log.Logger
	challengeTTL time.Duration
	isProduction bool
	otpPepper    string
}

// NewChallengeService constructs a new ChallengeService instance.
func NewChallengeService(
	rdc *redis.Client,
	notifier Notifier,
	logger log.Logger,
	challengeTTL time.Duration,
	isProduction bool,
	otpPepper string,
) *ChallengeService {
	return &ChallengeService{
		redisClient:  rdc,
		notifier:     notifier,
		logger:       logger,
		challengeTTL: challengeTTL,
		isProduction: isProduction,
		otpPepper:    otpPepper,
	}
}

// key generates the Redis hash key for the specified user identifier.
func key(identifier string) string {
	return fmt.Sprintf("challenge:%s", identifier)
}

// Create generates and stores a new OTP challenge in Redis.
// It verifies existing challenge status, resets resend counts, securely hashes the OTP secret,
// persists all fields in a Redis hash with an expiration TTL, and delivers the OTP code via Notifier.
func (cs *ChallengeService) Create(
	ctx context.Context,
	identifier string,
	channel string,
) (*model.Challenge, error) {
	// Generate cryptographically secure 6-digit random OTP
	otp, err := GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("create challenge: %w", err)
	}

	// Hash the raw OTP before storing in Redis so plaintext credentials never persist in cache
	otpHash, err := crypto.Hash(otp, cs.otpPepper)
	if err != nil {
		return nil, fmt.Errorf("create challenge: %w", err)
	}

	k := key(identifier)

	// Check whether an active challenge already exists for this identifier
	existing, err := cs.Get(ctx, identifier)

	var challenge *model.Challenge
	isNew := false

	switch {
	case err != nil:
		// Only a missing or expired challenge justifies creating a replacement.
		// Infrastructure errors (e.g. Redis failure) are surfaced instead of
		// silently resetting the resend quota.
		if ae, ok := err.(*apierr.APIError); ok && ae.Status == http.StatusUnauthorized {
			challenge = model.NewChallenge(identifier, model.ChallengeChannel(channel), otpHash, cs.challengeTTL)
			isNew = true
		} else {
			return nil, fmt.Errorf("create challenge: fetch existing: %w", err)
		}

	case existing.Status == model.ChallengeStatusPending:
		// Challenge is pending: verify resend limit cap before issuing a replacement OTP
		if existing.ResendCount >= MaxResendTimes {
			return nil, apierr.New(
				http.StatusTooManyRequests,
				"Too many resend attempts",
			).WithCode(apierr.CodeRateLimitExceeded)
		}
		existing.OtpHash = otpHash
		existing.ResendCount++
		// Refresh the expiration so a resend cannot expire the challenge mid-flow
		existing.ExpiresAt = time.Now().Add(cs.challengeTTL)
		challenge = existing

	case existing.Status == model.ChallengeStatusVerified || existing.Status == model.ChallengeStatusExpired:
		// Previous challenge finished or expired: initialize a fresh challenge
		challenge = model.NewChallenge(identifier, model.ChallengeChannel(channel), otpHash, cs.challengeTTL)
		isNew = true

	default:
		return nil, fmt.Errorf("unknown challenge status: %q", existing.Status)
	}

	// Persist challenge metadata and hashed OTP in Redis via pipeline with TTL
	// Always refresh TTL so a resend cannot expire mid-flow
	pipe := cs.redisClient.Pipeline()
	pipe.HSet(ctx, k, map[string]any{
		"id":           challenge.ID,
		"identifier":   challenge.Identifier,
		"channel":      string(challenge.Channel),
		"otp_hash":     challenge.OtpHash,
		"status":       string(challenge.Status),
		"attempts":     challenge.Attempts,
		"resend_count": challenge.ResendCount,
		"expires_at":   challenge.ExpiresAt.Format(time.RFC3339),
		"created_at":   challenge.CreatedAt.Format(time.RFC3339),
	})
	pipe.Expire(ctx, k, cs.challengeTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("create challenge: persist: %w", err)
	}

	// Deliver the raw OTP code through the external communication channel (email/SMS)
	// If notification fails, roll back state and delete the challenge so a failed
	// delivery never leaves a hash that cannot be matched to a delivered OTP.
	if err := cs.notifier.NotifyOTP(ctx, string(challenge.Channel), challenge.Identifier, otp); err != nil {
		cs.logger.Warn("failed to send otp code", log.Meta{"Error": err})

		if !isNew {
			// Roll back the resend-count increment so the failed send doesn't penalize the user
			if _, rbErr := cs.redisClient.HIncrBy(ctx, k, "resend_count", -1).Result(); rbErr != nil {
				cs.logger.Warn("failed to roll back resend_count", log.Meta{"Error": rbErr})
			}
		}

		// Delete the challenge so the flow can be restarted cleanly instead of
		// being stuck with a stored hash for an OTP the user never received.
		if delErr := cs.redisClient.Del(ctx, k).Err(); delErr != nil {
			cs.logger.Warn("failed to delete challenge after notify failure", log.Meta{"Error": delErr})
		}

		return nil, fmt.Errorf("create challenge: notify: %w", err)
	}

	return challenge, nil
}

// Get retrieves the active challenge hash from Redis, reconstructing the domain Challenge struct.
// Returns an Unauthorized API error if the challenge does not exist or has already expired.
func (cs *ChallengeService) Get(
	ctx context.Context,
	identifier string,
) (*model.Challenge, error) {
	// Retrieve all hash fields for the challenge identifier
	val, err := cs.redisClient.HGetAll(
		ctx,
		key(identifier),
	).Result()
	if err != nil {
		return nil, fmt.Errorf("get challenge:%w", err)
	}

	// If the hash is empty, the challenge expired or was never created
	if len(val) == 0 {
		return nil, apierr.New(
			http.StatusUnauthorized,
			"Challenge not found or expired",
		).WithCode(apierr.CodeUnauthorized)
	}

	// Parse RFC3339 expiration timestamp
	expiresAt, err := time.Parse(time.RFC3339, val["expires_at"])
	if err != nil {
		return nil, fmt.Errorf("get challenge:%w", err)
	}

	// Construct challenge entity
	challenge := &model.Challenge{
		ID:          val["id"],
		Identifier:  val["identifier"],
		Channel:     model.ChallengeChannel(val["channel"]),
		OtpHash:     val["otp_hash"],
		ResendCount: utils.StringToInt(val["resend_count"], 0),
		Attempts:    utils.StringToInt(val["attempts"], 0),
		Status:      model.ChallengeStatus(val["status"]),
		ExpiresAt:   expiresAt,
	}

	// Restore the original creation timestamp when present, so resends never
	// zero out created_at and the challenge's age stays accurate.
	if createdAtStr := val["created_at"]; createdAtStr != "" {
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			challenge.CreatedAt = createdAt
		}
	}

	return challenge, nil
}

// updateOTP updates the hashed OTP secret and resets the expiration TTL in Redis for an existing challenge.
func (cs *ChallengeService) updateOTP(
	ctx context.Context,
	identifier string,
	otp string,
) error {
	// Securely hash the new OTP code
	otpHash, err := crypto.Hash(otp, cs.otpPepper)
	if err != nil {
		return fmt.Errorf("update challenge otp:%w", err)
	}

	newExpiresAt := time.Now().Add(cs.challengeTTL).Format(time.RFC3339)

	// Pipeline hash field update and TTL refresh in Redis
	pipe := cs.redisClient.Pipeline()
	pipe.HSet(
		ctx,
		key(identifier),
		map[string]any{
			"otp_hash":   otpHash,
			"expires_at": newExpiresAt,
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

// Resend generates a new OTP, updates the stored hash, increments the resend counter,
// and dispatches the new OTP to the user.
func (cs *ChallengeService) Resend(
	ctx context.Context,
	challenge *model.Challenge,
) error {
	// Disallow resend if challenge is no longer pending (e.g. already verified or expired)
	if challenge.Status != model.ChallengeStatusPending {
		return apierr.New(
			http.StatusGone,
			"Challenge expired or invalid",
		).WithCode(apierr.CodeExpired)
	}

	// Disallow resend if expiration timestamp has elapsed
	if time.Now().After(challenge.ExpiresAt) {
		return apierr.New(
			http.StatusGone,
			"Challenge expired",
		).WithCode(apierr.CodeExpired)
	}

	challengeKey := key(challenge.Identifier)

	// Atomically increment resend counter in Redis to prevent concurrency race conditions
	newCount, err := cs.redisClient.HIncrBy(ctx, challengeKey, "resend_count", 1).Result()
	if err != nil {
		return fmt.Errorf("resend challenge: %w", err)
	}

	if newCount > MaxResendTimes {
		// Roll back the counter since cap was exceeded
		_ = cs.redisClient.HIncrBy(ctx, challengeKey, "resend_count", -1).Err()
		return apierr.New(
			http.StatusTooManyRequests,
			"Too many resend attempts",
		).WithCode(apierr.CodeRateLimitExceeded)
	}

	// Generate new random OTP
	newOtp, err := GenerateOTP()
	if err != nil {
		_ = cs.redisClient.HIncrBy(ctx, challengeKey, "resend_count", -1).Err()
		return fmt.Errorf("resend challenge: %w", err)
	}

	// Update OTP hash and refresh TTL in Redis
	if err := cs.updateOTP(ctx, challenge.Identifier, newOtp); err != nil {
		_ = cs.redisClient.HIncrBy(ctx, challengeKey, "resend_count", -1).Err()
		return err
	}

	// Send the new OTP code to the user's communication channel
	err = cs.notifier.NotifyOTP(
		ctx,
		string(challenge.Channel),
		challenge.Identifier,
		newOtp,
	)
	if err != nil {
		// Roll back resend_count via HIncrBy -1 on delivery error and return error
		_ = cs.redisClient.HIncrBy(ctx, challengeKey, "resend_count", -1).Err()
		cs.logger.Warn(
			"failed to send otp code",
			log.Meta{
				"Error": err,
			},
		)
		return fmt.Errorf("failed to send otp: %w", err)
	}
	return nil
}

// Verify validates a submitted OTP against the cryptographically hashed secret in Redis.
// It enforces challenge state checks, expiration checks, and attempt throttling.
func (cs *ChallengeService) Verify(
	ctx context.Context,
	challenge *model.Challenge,
	otp string,
) (bool, error) {
	// Reject if challenge has already been verified (prevents OTP reuse / replay attacks)
	if challenge.Status == model.ChallengeStatusVerified {
		return false, apierr.New(
			http.StatusConflict,
			"Challenge already verified",
		).WithCode(apierr.CodeResourceConflict)
	}

	// Reject if challenge status is expired
	if challenge.Status == model.ChallengeStatusExpired {
		return false, apierr.New(
			http.StatusGone,
			"Challenge expired",
		).WithCode(apierr.CodeExpired)
	}

	// Reject if wall-clock expiration timestamp has elapsed
	if time.Now().After(challenge.ExpiresAt) {
		return false, apierr.New(
			http.StatusGone,
			"Challenge expired",
		).WithCode(apierr.CodeExpired)
	}

	challengeKey := key(challenge.Identifier)

	// Atomically increment attempts via Redis HIncrBy BEFORE checking whether
	// the attempt count exceeds MaxVerifyAttempts, eliminating TOCTOU brute-force races.
	attempts, err := cs.redisClient.HIncrBy(ctx, challengeKey, "attempts", 1).Result()
	if err != nil {
		return false, fmt.Errorf("challenge verify track attempt: %w", err)
	}

	if attempts > MaxVerifyAttempts {
		return false, apierr.New(
			http.StatusTooManyRequests,
			"Too many verification attempts",
		).WithCode(apierr.CodeRateLimitExceeded)
	}

	// Compare submitted OTP against stored hash using constant-time cryptographic verification
	ok, err := crypto.Equal(challenge.OtpHash, otp, cs.otpPepper)
	if err != nil {
		return false, fmt.Errorf("challenge verify: %w", err)
	}

	if !ok {
		return false, nil
	}

	return true, nil
}

// markVerifiedScript is a Redis Lua script providing an atomic test-and-set transition:
// it verifies the challenge status is currently 'pending', and transitions it to 'verified'.
// Returns 1 on success, -1 if already verified (conflict), -2 if expired or not pending.
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

// MarkVerified executes an atomic Lua script in Redis to transition the challenge state
// to 'verified', guaranteeing that a verified OTP cannot be re-used in concurrent or replay requests.
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
	switch resNum {
	case -1:
		return apierr.New(http.StatusConflict, "Challenge already verified").WithCode(apierr.CodeResourceConflict)
	case -2:
		return apierr.New(http.StatusGone, "Challenge expired or invalid").WithCode(apierr.CodeExpired)
	}

	return nil
}

// Expire transitions the challenge status in Redis to 'expired' immediately.
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
