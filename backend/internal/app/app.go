package app

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/auth"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/brand"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/cart"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/category"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/product"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/review"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/tag"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/variant"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/checkout"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/http/swagger"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/notifier"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/object"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/order"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/job"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/user"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type App struct {
	// http server
	server *http.Server

	// logger
	logger *log.ConsoleLogger

	// cache
	redisClient *redis.Client

	// database
	db *database.DB

	// minio
	minioClient *minio.Client

	// storage provider
	storageProvider storage.StorageProvider
}

func (app *App) setupRoutes(config *config.Config, router *gin.Engine) {
	// setup middlewares
	router.Use(middleware.ErrorHandler(app.logger))
	router.Use(middleware.CORS(config.AllowedOrigins...))

	v1 := router.Group("/api/v1")
	swagger.SetUpDocs(v1)

	txRunner := database.NewTxRunner(app.db)

	jobQueue := job.NewJobQueue(
		app.redisClient,
		job.EmailQueue,
	)

	notifier := notifier.NewEmailNotifier(
		jobQueue,
		app.logger,
	)

	rateLimiter := security.NewRateLimiter(
		app.redisClient,
	)

	userRepo := user.NewPostgresRepository()
	userService := user.NewService(
		txRunner,
		userRepo,
		app.logger,
	)

	jwtService := jwt.NewJWTManager(
		config.KeysCfg,
	)

	challengeService := auth.NewChallengeService(
		app.redisClient,
		auth.NewOTPGenerator(),
		notifier,
		app.logger,
	)

	authService := auth.NewAuthService(
		app.logger,
		challengeService,
		userService,
		jwtService,
		app.redisClient,
		notifier,
		config.KeysCfg,
	)

	objectRepository := object.NewRepository()
	objectService := object.NewService(
		txRunner,
		objectRepository,
		app.storageProvider,
	)

	brandRepo := brand.NewRepository()
	brandService := brand.NewService(
		txRunner,
		brandRepo,
		objectService,
	)
	brandHandler := brand.NewHandler(
		brandService,
	)
	brandRouter := brand.NewRouter(
		brandHandler,
		config.KeysCfg,
	)
	brandRouter.MapRoutes(v1)

	tagRepo := tag.NewRepository()
	tagService := tag.NewService(
		txRunner,
		tagRepo,
	)
	tagHandler := tag.NewHandler(
		tagService,
	)
	tagRouter := tag.NewRouter(
		tagHandler,
		config.KeysCfg,
	)
	tagRouter.MapRoutes(v1)

	categoryRepo := category.NewRepository()
	categoryService := category.NewService(
		txRunner,
		categoryRepo,
	)
	categoryHandler := category.NewHandler(
		categoryService,
	)
	categoryRouter := category.NewRouter(
		categoryHandler,
		config.KeysCfg,
	)
	categoryRouter.MapRoutes(v1)

	variantRepository := variant.NewRepository()
	variantService := variant.NewService(
		app.logger,
		txRunner,
		objectService,
		variantRepository,
	)

	variantHandler := variant.NewHandler(variantService)
	variantRouter := variant.NewRouter(variantHandler, config.KeysCfg)
	variantRouter.MapRoutes(v1)

	// review
	reviewRepo := review.NewReviewRepository()
	reviewService := review.NewService(txRunner, reviewRepo)
	reviewHandler := review.NewHandler(reviewService)
	reviewRouter := review.NewRouter(reviewHandler, config.KeysCfg)
	reviewRouter.MapRoutes(v1)

	// product
	productRepo := product.NewProductRepository()
	productService := product.NewService(
		txRunner,
		app.logger,
		productRepo,
		objectService,
		brandService,
		categoryService,
		tagService,
		variantService,
		reviewService,
	)
	productHandler := product.NewHandler(productService)
	productRouter := product.NewRouter(productHandler, config.KeysCfg)
	productRouter.MapRoutes(v1)

	// cart
	cartRepo := cart.NewRepository()
	cartService := cart.NewService(txRunner, cartRepo, variantService, productService)
	cartHandler := cart.NewHandler(cartService)
	cartRouter := cart.NewRouter(cartHandler, config.KeysCfg)
	cartRouter.MapRoutes(v1)

	authHandler := auth.NewAuthHandler(
		authService,
		rateLimiter,
		app.logger,
		config.IsProduction,
		cartService,
	)

	authRouter := auth.NewRouter(
		authHandler,
		config.KeysCfg,
	)
	authRouter.MapRoutes(v1)

	userHandler := user.NewHandler(
		userService,
		rateLimiter,
		config.KeysCfg,
		app.logger,
	)

	userRouter := user.NewRouter(
		userHandler,
		config,
	)

	userRouter.MapRoutes(v1)

	// order
	orderRepo := order.NewRepository()
	orderService := order.NewService(txRunner, orderRepo, app.logger)
	orderHandler := order.NewHandler(orderService)
	orderRouter := order.NewRouter(orderHandler, config.KeysCfg)
	orderRouter.MapRoutes(v1)

	// checkout
	checkoutService := checkout.NewService(cartService, orderService)
	checkoutHandler := checkout.NewHandler(checkoutService)
	checkoutRouter := checkout.NewRouter(checkoutHandler, config.KeysCfg)
	checkoutRouter.MapRoutes(v1)
}

func (app *App) Shutdown() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if app.server != nil {
		err := app.server.Shutdown(ctx)
		if err != nil {
			app.logger.Info("the server is down")
		}
	}

	if app.db != nil {
		app.db.Close()
	}

	app.logger.Debug("Graceful Shutdown")
}

func (app *App) Run() error {
	app.logger = log.NewConsoleLogger()
	err := godotenv.Load()
	if err != nil {
		app.logger.Warn(
			"failed to load .env file",
			log.Meta{
				"Error": err.Error(),
			},
		)
	}

	cfg := config.Load()

	app.db, err = database.ConnectDB(context.Background(), cfg)
	if err != nil {
		app.logger.Error(
			"database connection issue",
			log.Meta{
				"Error": err.Error(),
			},
		)
		return err
	}

	if err := app.db.Ping(context.Background()); err != nil {
		app.logger.Warn(
			"database is not live",
			log.Meta{
				"Error": err.Error(),
			},
		)
	}

	app.minioClient, err = minio.New(cfg.MinioCfg.Endpoint, &minio.Options{
		Creds:      credentials.NewStaticV4(cfg.MinioCfg.AccessKey, cfg.MinioCfg.SecretKey, ""),
		Secure:     false,
		EnableRDMA: true,
	})
	if err != nil {
		app.logger.Error(
			"minio connection issue",
			log.Meta{
				"Error": err,
			},
		)
		return err
	}

	app.storageProvider, err = storage.NewMinioStorageProvider(
		cfg.MinioCfg.Endpoint,
		cfg.MinioCfg.AccessKey,
		cfg.MinioCfg.SecretKey,
		cfg.MinioCfg.PublicURL,
	)
	if err != nil {
		app.logger.Error(
			"storage provider init issue",
			log.Meta{
				"Error": err,
			},
		)
		return err
	}

	exists, err := app.minioClient.BucketExists(
		context.Background(),
		"product-media",
	)
	if err != nil {
		app.logger.Error(
			"minio bucket issue",
			log.Meta{
				"Error": err,
			},
		)
		return err
	}
	if !exists {
		err := app.minioClient.MakeBucket(
			context.Background(),
			"product-media",
			minio.MakeBucketOptions{},
		)
		if err != nil {
			app.logger.Error(
				"minio bucket issue",
				log.Meta{
					"Error": err,
				},
			)
			return err
		}
		app.logger.Info(
			"minio bucket created",
			log.Meta{
				"Bucket": "product-media",
			},
		)
	}

	app.redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.RedisCfg.Host,
			cfg.RedisCfg.Port,
		),
	})

	if err := app.redisClient.Ping(context.Background()); err != nil {
		app.logger.Warn(
			"redis connection issue",
			log.Meta{
				"Error": err,
			},
		)
	}

	router := gin.Default()
	app.setupRoutes(cfg, router)

	app.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.SvPort),
		Handler: router,
	}

	app.logger.Info(
		"Server started",
		log.Meta{
			"URL":  fmt.Sprintf("http://localhost:%s", cfg.SvPort),
			"Port": cfg.SvPort,
		},
	)
	return app.server.ListenAndServe()
}
