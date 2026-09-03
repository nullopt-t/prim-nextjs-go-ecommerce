package app

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/auth"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/brand"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/cart"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/category"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/inventory"
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

	// app config
	config config.Config
}

func (app *App) setupRoutes(router *gin.Engine) {
	// setup middlewares
	router.Use(middleware.ErrorHandler(app.logger))
	router.Use(middleware.CORS(app.config.AllowedOrigins...))

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
		app.config.KeysCfg,
	)

	challengeService := auth.NewChallengeService(
		app.redisClient,
		notifier,
		app.logger,
		app.config.AuthCfg.ChallengeTTL,
		app.config.IsProduction,
	)

	sessionService := auth.NewSessionService(
		app.redisClient,
		app.logger,
		app.config.AuthCfg.SessionTTL,
	)

	authService := auth.NewAuthService(
		app.logger,
		challengeService,
		userService,
		sessionService,
		jwtService,
		app.redisClient,
		notifier,
		app.config.KeysCfg,
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
		app.config.KeysCfg,
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
		app.config.KeysCfg,
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
		app.config.KeysCfg,
	)
	categoryRouter.MapRoutes(v1)

	// inventory
	inventoryRepo := inventory.NewRepository()
	inventoryService := inventory.NewService(
		app.logger,
		txRunner,
		inventoryRepo,
	)

	variantRepository := variant.NewRepository()
	variantService := variant.NewService(
		app.logger,
		txRunner,
		objectService,
		variantRepository,
		inventoryService,
	)

	variantHandler := variant.NewHandler(variantService)
	variantRouter := variant.NewRouter(variantHandler, app.config.KeysCfg)
	variantRouter.MapRoutes(v1)

	// review
	reviewRepo := review.NewReviewRepository()
	reviewService := review.NewService(txRunner, reviewRepo)
	reviewHandler := review.NewHandler(reviewService)
	reviewRouter := review.NewRouter(reviewHandler, app.config.KeysCfg)
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
		inventoryService,
		reviewService,
	)
	productHandler := product.NewHandler(productService)
	productRouter := product.NewRouter(productHandler, app.config.KeysCfg)
	productRouter.MapRoutes(v1)

	// cart
	cartRepo := cart.NewRepository()
	cartService := cart.NewService(txRunner, cartRepo, variantService, productService, inventoryService)
	cartHandler := cart.NewHandler(cartService)
	cartRouter := cart.NewRouter(cartHandler, app.config.KeysCfg)
	cartRouter.MapRoutes(v1)

	authHandler := auth.NewAuthHandler(
		authService,
		sessionService,
		app.logger,
		app.config.IsProduction,
		cartService,
		app.config.AuthCfg.ChallengeTTL,
	)

	authRouter := auth.NewRouter(
		authHandler,
		app.config.KeysCfg,
		rateLimiter,
		app.logger,
		app.redisClient,
		app.config.RateLimitCfg,
	)
	authRouter.MapRoutes(v1)

	userHandler := user.NewHandler(
		userService,
		rateLimiter,
		app.config.KeysCfg,
		app.logger,
	)

	userRouter := user.NewRouter(
		userHandler,
		app.config,
	)

	userRouter.MapRoutes(v1)

	// order
	orderRepo := order.NewRepository()
	orderService := order.NewService(txRunner, orderRepo, app.logger)
	orderHandler := order.NewHandler(orderService)
	orderRouter := order.NewRouter(orderHandler, app.config.KeysCfg)
	orderRouter.MapRoutes(v1)

	// checkout
	checkoutService := checkout.NewService(cartService, orderService)
	checkoutHandler := checkout.NewHandler(checkoutService)
	checkoutRouter := checkout.NewRouter(checkoutHandler, app.config.KeysCfg)
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

	app.config = config.Load()
	app.db, err = database.ConnectDB(context.Background(), app.config.DBCfg)
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

	storageProv, err := storage.NewMinioStorageProvider(
		app.config.MinioCfg.Endpoint,
		app.config.MinioCfg.AccessKey,
		app.config.MinioCfg.SecretKey,
		app.config.MinioCfg.PublicURL,
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
	app.storageProvider = storageProv
	app.minioClient = storageProv.Client

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
			app.config.RedisCfg.Host,
			app.config.RedisCfg.Port,
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
	app.setupRoutes(router)

	app.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", app.config.SvPort),
		Handler: router,
	}

	app.logger.Info(
		"Server started",
		log.Meta{
			"URL":  fmt.Sprintf("http://localhost:%s", app.config.SvPort),
			"Port": app.config.SvPort,
		},
	)
	return app.server.ListenAndServe()
}
