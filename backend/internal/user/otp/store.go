package otp

import (
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type OTP struct {
	CodeHash string `json:"code_hash"`
	Attempts int    `json:"attempts"`
}

type OTPStore struct {
	redisClient *redis.Client
	logger      log.Logger
}

func NewOTPStore(
	redisClient *redis.Client,
	logger log.Logger,
) *OTPStore {
	return &OTPStore{
		redisClient: redisClient,
		logger:      logger,
	}
}

func (r *OTPStore) Save(
	ctx context.Context,
	key string,
	otp OTP,
	expiration time.Duration,
) error {
	r.logger.Debug(
		"saving otp",
		log.Meta{
			"key":        key,
			"expiration": expiration,
		},
	)

	mo, err := json.Marshal(otp)
	if err != nil {
		return fmt.Errorf("marshaling otp: %w", err)
	}

	err = r.redisClient.Set(
		ctx,
		key,
		string(mo),
		expiration,
	).Err()
	if err != nil {
		return fmt.Errorf("saving otp: %w", err)
	}

	return nil
}

func (r *OTPStore) Get(
	ctx context.Context,
	key string,
) (*OTP, error) {
	r.logger.Debug(
		"getting otp",
		log.Meta{
			"key": key,
		},
	)

	mo, err := r.redisClient.Get(
		ctx,
		key,
	).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("get otp: key not found")
		}
		return nil, fmt.Errorf("get otp: %w", err)
	}

	var uo OTP
	err = json.Unmarshal([]byte(mo), &uo)
	if err != nil {
		return nil, fmt.Errorf("get otp: %w", err)
	}

	return &uo, nil
}

func (r *OTPStore) Increment(
	ctx context.Context,
	key string,
) error {
	r.logger.Debug(
		"incrementing otp attempts",
		log.Meta{
			"key": key,
		},
	)

	mo, err := r.redisClient.Get(
		ctx,
		key,
	).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("increment otp: %w", err)
	}

	var uo OTP
	err = json.Unmarshal([]byte(mo), &uo)
	if err != nil {
		return fmt.Errorf("increment otp: %w", err)
	}

	uo.Attempts++

	marshaledOTP, err := json.Marshal(uo)
	if err != nil {
		return fmt.Errorf("increment otp: %w", err)
	}

	err = r.redisClient.Set(
		ctx,
		key,
		string(marshaledOTP),
		0,
	).Err()
	if err != nil {
		return fmt.Errorf("increment otp: %w", err)
	}

	return nil
}

func (r *OTPStore) Delete(
	ctx context.Context,
	key string,
) error {
	r.logger.Debug(
		"deleting otp",
		log.Meta{
			"key": key,
		},
	)
	return r.redisClient.Del(
		ctx,
		key,
	).Err()
}
