package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	_ "api-structure/docs"
	"api-structure/internal/adapter/auth/jwt"
	"api-structure/internal/adapter/config"
	httpLib "api-structure/internal/adapter/handler/http"
	"api-structure/internal/adapter/logger"
	"api-structure/internal/adapter/storage/postgres"
	"api-structure/internal/adapter/storage/postgres/repository"
	"api-structure/internal/adapter/storage/redis"
	"api-structure/internal/core/service"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// @title						Savely
// @version					1.0
// @description				A personal finance application
//
// @contact.name				Emmanuel Jonathan
// @contact.url				https://github.com/emmrys-jay
// @contact.email				jonathanemma121@gmail.com
//
// @host						localhost:8080
// @BasePath					/api/v1
// @schemes					http https
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and the access token.
func main() {
	// Load environment variables
	config := config.Setup()

	// Set logger
	l := logger.Get()

	l.Info("Starting the application",
		zap.String("app", config.App.Name),
		zap.String("env", config.App.Env))

	// Init database
	ctx := context.Background()
	db, err := postgres.New(ctx, &config.Database)
	if err != nil {
		l.Error("Error initializing database connection", zap.Error(err))
		os.Exit(1)
	}
	defer db.Close()

	l.Info("Successfully connected to the database",
		zap.String("db", config.Database.Protocol))

	// Migrate postgres database
	err = db.Migrate()
	if err != nil {
		l.Error("Error migrating database", zap.Error(err))
		os.Exit(1)
	}

	l.Info("Successfully migrated the database")

	// Init cache service
	cache, err := redis.New(ctx, &config.Redis)
	if err != nil {
		l.Error("Error initializing cache connection", zap.Error(err))
		// os.Exit(1) // Cache is not being used at the moment
	}
	defer cache.Close()

	l.Info("Successfully connected to the cache server")

	// Init token service
	tokenService := jwt.New(&config.Token)

	// Dependency injection
	// Ping
	pingRepo := repository.NewPingRepository(db)
	pingService := service.NewPingService(pingRepo, cache)
	pingHandler := httpLib.NewPingHandler(pingService, validator.New())

	// User
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cache)
	userHandler := httpLib.NewUserHandler(userService, validator.New())

	// Auth
	authService := service.NewAuthService(userRepo, tokenService, cache)
	authHandler := httpLib.NewAuthHandler(authService, validator.New())

	// Init router
	router, err := httpLib.NewRouter(
		&config.Server,
		tokenService,
		l,
		*pingHandler,
		*userHandler,
		*authHandler,
	)
	if err != nil {
		l.Error("Error initializing router ", zap.Error(err))
		os.Exit(1)
	}

	// Create Admin User
	if err := userService.CreateAdminUser(context.Background(), config.Admin.Email, config.Admin.Password); err != nil {
		if err.Code() != 500 {
			l.Info("Admin user already exists", zap.String("info", err.Error()))
		} else {
			l.Error("Could not create admin user, exiting...", zap.Error(err))
			os.Exit(1)
		}
	} else {
		l.Info("Successfully created admin user with email: ", zap.String("email", config.Admin.Email))
	}

	// Start server
	listenAddr := fmt.Sprintf("%s:%s", config.Server.HttpUrl, config.Server.HttpPort)
	l.Info("Starting the HTTP server", zap.String("listen_address", listenAddr))

	err = http.ListenAndServe(listenAddr, router)
	if err != nil {
		l.Error("Error starting the HTTP server", zap.Error(err))
		os.Exit(1)
	}
}
