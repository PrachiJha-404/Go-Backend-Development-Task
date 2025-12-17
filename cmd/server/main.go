package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	database "user-api/db/sqlc"
	"user-api/internal/handler"
	"user-api/internal/logger"
	"user-api/internal/middleware"
	"user-api/internal/repository"
	"user-api/internal/routes"
	"user-api/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	logger, err := logger.NewLoggerFromEnv()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	} //Don't run the server if it's blind
	defer logger.Sync()

	middleware.SetLogger(logger)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/userdb?sslmode=disable"
		logger.Warn("DATABASE_URL not set, using default")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}
	logger.Info("successfully connected to database")

	queries := database.New(db)
	userRepo := repository.NewUserRepository(queries)
	userService := service.NewUserService(userRepo, logger)
	userHandler := handler.NewUserHandler(*userService, logger)

	app := fiber.New(fiber.Config{AppName: "User API v1.0",
		ErrorHandler: customErrorHandler(logger),
	})

	app.Use(recover.New())
	app.Use(middleware.CORS())
	app.Use(middleware.ErrorHandler())

	routes.SetupRoutes(app, userHandler)

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			logger.Error("server shutdown error", zap.Error(err))
		}
	}()

	logger.Info("starting sevrer", zap.String("port", port))
	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}

func customErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}
		logger.Error("error occured",
			zap.Int("status", code),
			zap.String("path", c.Path()),
			zap.Error(err),
		)
		return c.Status(code).JSON(fiber.Map{"error": err.Error()})
	}
}
