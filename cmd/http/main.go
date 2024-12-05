package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	_ "savely/docs"
	"savely/internal/adapter/auth/jwt"
	"savely/internal/adapter/config"
	httpLib "savely/internal/adapter/handler/http"
	"savely/internal/adapter/logger"
	"savely/internal/adapter/storage/postgres"
	"savely/internal/adapter/storage/postgres/repository"
	"savely/internal/adapter/storage/redis"
	"savely/internal/core/service"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// @title						Savely (Smart Personal Finance Manager) API
// @version					1.0
// @description				This is a RESTful personal finance API written in Go using go-chi, PostgreSQL database, and Redis cache.
//
// @contact.name				Emmanuel Jonathan
// @contact.url				https://github.com/emmrys-jay/savely
// @contact.email				jonathanemma121@gmail.com
//
// @host						http://localhost:8080
// @BasePath					/v1
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

	// Migrate database
	// err = db.Migrate()
	// if err != nil {
	// 	l.Error("Error migrating database", zap.Error(err))
	// 	os.Exit(1)
	// }

	// l.Info("Successfully migrated the database")

	// Init cache service
	cache, err := redis.New(ctx, &config.Redis)
	if err != nil {
		l.Error("Error initializing cache connection", zap.Error(err))
		os.Exit(1)
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

	// Init router
	router, err := httpLib.NewRouter(
		&config.Server,
		tokenService,
		*pingHandler,
	)
	if err != nil {
		l.Error("Error initializing router ", zap.Error(err))
		os.Exit(1)
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
