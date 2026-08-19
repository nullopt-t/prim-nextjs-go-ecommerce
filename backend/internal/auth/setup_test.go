package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/auth"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/user"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

type mockNotifier struct {
	LastOTP string
}

func (m *mockNotifier) NotifyOTP(ctx context.Context, channel, identifier, otp string) error {
	m.LastOTP = otp
	return nil
}

type AuthTestSuite struct {
	suite.Suite
	pgContainer    *postgres.PostgresContainer
	redisContainer *redisContainer.RedisContainer
	db             *database.DB
	redisClient    *redis.Client
	router         *gin.Engine
	notifier       *mockNotifier
}

func (s *AuthTestSuite) SetupSuite() {
	ctx := context.Background()

	// 1. Start Postgres Container
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(s.T(), err)
	s.pgContainer = pgContainer

	// We have to connect manually since pkg/database uses individual fields, not DSN string.
	// But actually, database.ConnectDB takes the config fields. Let's parse the connection string to fields.
	// For simplicity, let's just use pgxpool directly here to run migrations and then pass to DB.
	// We can't easily override host/port in database.ConnectDB without getting the mapped port.
	// Better: get host and port
	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)

	db, err := database.ConnectDB(ctx, config.Config{
		DBCfg: config.DatabaseConfig{
			DBHost:     host,
			DBPort:     port.Port(),
			DBUser:     "testuser",
			DBPassword: "testpass",
			DBName:     "testdb",
		},
	})
	require.NoError(s.T(), err)
	s.db = db

	// Run migration
	schemaBytes, err := os.ReadFile("../../migrations/000001_init_sechema.up.sql")
	require.NoError(s.T(), err)
	_, err = db.Exec(ctx, string(schemaBytes))
	require.NoError(s.T(), err)

	// 3. Start Redis Container
	rContainer, err := redisContainer.Run(ctx,
		"redis:7-alpine",
	)
	require.NoError(s.T(), err)
	s.redisContainer = rContainer

	redisHost, err := rContainer.Host(ctx)
	require.NoError(s.T(), err)
	redisPort, err := rContainer.MappedPort(ctx, "6379")
	require.NoError(s.T(), err)

	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort.Port(),
	})
	s.redisClient = rdb

	// 4. Setup Services
	logger := log.NewConsoleLogger()
	s.notifier = &mockNotifier{}

	challengeTTL := 5 * time.Minute
	sessionTTL := 30 * 24 * time.Hour

	challengeService := auth.NewChallengeService(s.redisClient, s.notifier, logger, challengeTTL)
	sessionService := auth.NewSessionService(s.redisClient, logger, sessionTTL)

	txRunner := database.NewTxRunner(s.db)
	userRepo := user.NewPostgresRepository()
	userService := user.NewService(txRunner, userRepo, logger)

	secrets := config.Secrets{
		JwtAccessTokenSecretKey:  "access",
		JwtRefreshTokenSecretKey: "refresh",
	}
	jwtService := jwt.NewJWTManager(secrets)

	authService := auth.NewAuthService(
		logger,
		challengeService,
		userService,
		sessionService,
		jwtService,
		s.redisClient,
		s.notifier,
		secrets,
	)

	// 5. Setup Gin Router
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.ErrorHandler(logger))

	// Missing cartService for authHandler. We can pass nil since cart merging isn't used in these flows immediately,
	// or create a dummy interface. Wait, let's pass nil and hope it doesn't panic on normal flows.
	// Wait, VerifyChallenge might use CartMerger. Let's look at auth_handler.go.
	authHandler := auth.NewAuthHandler(authService, sessionService, logger, false, nil, challengeTTL)
	rateLimiter := security.NewRateLimiter(s.redisClient)
	authRouter := auth.NewRouter(
		authHandler,
		secrets,
		rateLimiter,
		logger,
		s.redisClient,
		config.RateLimitConfig{
			AuthStartRequests:  5,
			AuthStartWindow:    time.Minute,
			AuthResendRequests: 3,
			AuthResendWindow:   time.Minute,
			AuthVerifyRequests: 10,
			AuthVerifyWindow:   time.Minute,
		},
	)

	v1 := s.router.Group("/api/v1")
	authRouter.MapRoutes(v1)
}

func (s *AuthTestSuite) TearDownSuite() {
	ctx := context.Background()
	if s.pgContainer != nil {
		s.pgContainer.Terminate(ctx)
	}
	if s.redisContainer != nil {
		s.redisContainer.Terminate(ctx)
	}
}

func (s *AuthTestSuite) SetupTest() {
	// Clear Redis before each test
	s.redisClient.FlushAll(context.Background())
	s.notifier.LastOTP = ""
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
