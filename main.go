package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"load-test/internal/api"
	"load-test/internal/cache"
	"load-test/internal/config"
	"load-test/internal/database"
	"load-test/internal/repository"
	"load-test/pkg/postgres"
	"load-test/pkg/redis"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	if os.Getenv("ENVIRONMENT") == "production" {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func main() {
	log.Info("Starting Books API application")

	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to PostgreSQL")
	}
	defer pool.Close()

	migrationsPath := "file://db/migrations"
	if err := database.RunMigrations(cfg.Database.DSN(), migrationsPath); err != nil {
		log.WithError(err).Fatal("Failed to run migrations")
	}

	redisClient, err := redis.NewClient(ctx, cfg.Redis.Address(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	repo := repository.NewPostgresBookRepository(pool)
	log.Info("PostgreSQL repository initialized successfully")

	c := cache.NewRedisCache(redisClient, cfg.Redis.CacheTTL)
	log.WithField("ttl", cfg.Redis.CacheTTL).Info("Redis cache initialized")

	apiServer := api.NewBooksAPIServer(repo, c)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogMethod: true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.WithFields(log.Fields{
					"method": v.Method,
					"uri":    v.URI,
					"status": v.Status,
				}).WithError(v.Error).Error("Request failed")
			} else {
				log.WithFields(log.Fields{
					"method": v.Method,
					"uri":    v.URI,
					"status": v.Status,
				}).Info("Request completed")
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Load OpenAPI spec
	spec, err := api.GetSwagger()
	if err != nil {
		log.WithError(err).Fatal("Failed to load OpenAPI spec")
	}
	spec.Servers = nil // Allow any server

	// Register generated handlers first
	api.RegisterHandlers(e, apiServer)
	log.Info("API handlers registered from OpenAPI spec")

	// Add OpenAPI validation middleware after routes are registered
	if cfg.Swagger.Enabled {
		// Note: Validation middleware is optional and can cause issues
		// Uncomment the line below to enable request validation
		// e.Use(echomiddleware.OapiRequestValidator(spec))
		log.Info("OpenAPI validation available (currently disabled for compatibility)")
	}

	// Serve OpenAPI spec
	e.GET("/openapi.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, spec)
	})

	// Serve Swagger UI
	if cfg.Swagger.Enabled {
		e.GET("/swagger", api.ServeSwaggerUI)
		e.GET("/docs", api.ServeSwaggerUI)
		log.Info("Swagger UI available at /swagger or /docs")
		log.Info("OpenAPI spec available at /openapi.json")
	}

	go func() {
		addr := ":" + cfg.Server.Port
		log.WithField("port", cfg.Server.Port).Info("Server starting")
		if err := e.Start(addr); err != nil {
			log.WithError(err).Info("Server stopped")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server forced to shutdown")
	}

	log.Info("Server exited gracefully")
}
