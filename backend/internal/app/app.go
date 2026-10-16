package app

import (
	"github.com/nullopt-t/prim-backend/internal/auth"
	"github.com/nullopt-t/prim-backend/internal/catalog/brand"
	"github.com/nullopt-t/prim-backend/internal/catalog/cart"
	"github.com/nullopt-t/prim-backend/internal/catalog/category"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/catalog/review"
	"github.com/nullopt-t/prim-backend/internal/catalog/tag"
	"github.com/nullopt-t/prim-backend/internal/catalog/variant"
	"github.com/nullopt-t/prim-backend/internal/catalog/wishlist"
	"github.com/nullopt-t/prim-backend/internal/checkout"
	"github.com/nullopt-t/prim-backend/internal/http/swagger"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/internal/object"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/internal/admin"
	"github.com/minio/minio-go/v7"

	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/api/security"
	brokerkafka "github.com/nullopt-t/prim-backend/pkg/broker/kafka"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"

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

	// kafka producer
	kafkaProducer *brokerkafka.Producer

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
	router.Use(middleware.Locale())

	v1 := router.Group("/api/v1")
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	txRunner := database.NewTxRunner(app.db)

	kafkaCfg := brokerkafka.DefaultConfig(app.config.KafkaCfg.Brokers, app.config.KafkaCfg.GroupID)
	app.kafkaProducer = brokerkafka.NewProducer(kafkaCfg)
	notifier := notifier.NewKafkaNotifier(app.kafkaProducer, app.logger)

	rateLimiter := security.NewRateLimiter(
		app.redisClient,
	)

	userRepo := user.NewPostgresRepository()
	userService := user.NewService(
		txRunner,
		userRepo,
		app.logger,
	)


	challengeService := auth.NewChallengeService(
		app.redisClient,
		notifier,
		app.logger,
		app.config.AuthCfg.ChallengeTTL,
		app.config.IsProduction,
		app.config.KeysCfg.OTPHashPepper,
	)

	sessionService := auth.NewSessionService(
		app.redisClient,
		app.logger,
		app.config.AuthCfg.SessionTTL,
	)
	userService.SetSessionManager(sessionService)

	swagger.SetUpDocs(v1, sessionService)

	authService := auth.NewAuthService(
		app.logger,
		challengeService,
		userService,
		sessionService,
		notifier,
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
	brandRouter := brand.NewRouter(brandHandler, sessionService)
	brandRouter.MapRoutes(v1)

	tagRepo := tag.NewRepository()
	tagService := tag.NewService(
		txRunner,
		tagRepo,
	)
	tagHandler := tag.NewHandler(
		tagService,
	)
	tagRouter := tag.NewRouter(tagHandler, sessionService)
	tagRouter.MapRoutes(v1)

	categoryRepo := category.NewRepository()
	categoryService := category.NewService(
		txRunner,
		categoryRepo,
	)
	categoryHandler := category.NewHandler(
		categoryService,
	)
	categoryRouter := category.NewRouter(categoryHandler, sessionService)
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
	variantRouter := variant.NewRouter(variantHandler, sessionService)
	variantRouter.MapRoutes(v1)

	// review
	reviewRepo := review.NewReviewRepository()
	reviewService := review.NewService(txRunner, reviewRepo)
	reviewHandler := review.NewHandler(reviewService)
	reviewRouter := review.NewRouter(reviewHandler, sessionService)
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
	productRouter := product.NewRouter(productHandler, sessionService)
	productRouter.MapRoutes(v1)

	// cart
	cartRepo := cart.NewRepository()
	cartService := cart.NewService(txRunner, cartRepo, variantService, productService, inventoryService)
	cartHandler := cart.NewHandler(cartService)
	cartRouter := cart.NewRouter(cartHandler, sessionService)
	cartRouter.MapRoutes(v1)

	// wishlist
	wishlistRepo := wishlist.NewWishlistRepository()
	wishlistService := wishlist.NewService(txRunner, wishlistRepo, productService, objectService, app.logger)
	wishlistHandler := wishlist.NewHandler(wishlistService)
	wishlistRouter := wishlist.NewRouter(wishlistHandler, sessionService)
	wishlistRouter.MapRoutes(v1)

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
		sessionService,
		app.config.RateLimitCfg,
	)
	authRouter.MapRoutes(v1)

	userHandler := user.NewHandler(
		userService,
		rateLimiter,
		app.config.KeysCfg,
		app.logger,
	)

	userRouter := user.NewRouter(userHandler, sessionService)

	userRouter.MapRoutes(v1)

	notificationRepo := user.NewNotificationRepository()
	notificationService := user.NewNotificationService(txRunner, notificationRepo, app.logger)
	notificationHandler := user.NewNotificationHandler(notificationService)
	notificationRouter := user.NewNotificationRouter(notificationHandler, sessionService)
	notificationRouter.MapRoutes(v1)

	addressRepo := user.NewAddressRepository()
	addressService := user.NewAddressService(txRunner, addressRepo, app.logger)
	addressHandler := user.NewAddressHandler(addressService)
	addressRouter := user.NewAddressRouter(addressHandler, sessionService)
	addressRouter.MapRoutes(v1)

	// order
	orderRepo := order.NewRepository()
	orderService := order.NewService(txRunner, orderRepo, app.logger, inventoryService)
	orderHandler := order.NewHandler(orderService)
	orderRouter := order.NewRouter(orderHandler, sessionService)
	orderRouter.MapRoutes(v1)

	// checkout
	checkoutService := checkout.NewService(cartService, orderService, userService, addressService, inventoryService, txRunner)
	checkoutHandler := checkout.NewHandler(checkoutService)
	checkoutRouter := checkout.NewRouter(checkoutHandler, sessionService)
	checkoutRouter.MapRoutes(v1)

	// admin backoffice
	adminRepo := admin.NewRepository()
	adminService := admin.NewService(txRunner, adminRepo, objectService)
	adminHandler := admin.NewHandler(adminService)
	adminRouter := admin.NewRouter(adminHandler, sessionService)
	adminRouter.MapRoutes(v1)
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

	if app.kafkaProducer != nil {
		_ = app.kafkaProducer.Close()
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

	app.config, err = config.Load()
	if err != nil {
		app.logger.Error(
			"failed to load configuration",
			log.Meta{
				"Error": err.Error(),
			},
		)
		return err
	}
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

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Trust no proxies by default: c.ClientIP() then falls back to RemoteAddr,
	// so client-supplied X-Forwarded-For/X-Real-IP cannot spoof rate limits.
	// If the deployment sits behind known reverse proxies, replace nil with
	// their CIDRs here: router.SetTrustedProxies([]string{"..."}).
	if err := router.SetTrustedProxies(nil); err != nil {
		app.logger.Warn("failed to configure trusted proxies", log.Meta{"Error": err})
	}

	app.setupRoutes(router)

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetHTTP2(true)
	protocols.SetUnencryptedHTTP2(true)

	app.server = &http.Server{
		Addr:      fmt.Sprintf(":%s", app.config.SvPort),
		Handler:   router,
		Protocols: protocols,
	}

	app.logger.Info(
		"Server started with HTTP/1.1 and HTTP/2 (h2c) support",
		log.Meta{
			"URL":  fmt.Sprintf("http://localhost:%s", app.config.SvPort),
			"Port": app.config.SvPort,
		},
	)
	return app.server.ListenAndServe()
}
