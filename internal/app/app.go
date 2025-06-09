package app

import (
	"fmt"
	"server/internal/config"
	"server/internal/handler/http"
	"server/internal/models"
	"server/internal/pkg/logger"
	postgr "server/internal/repository/postgres"
	"server/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Run() {
	// 1. Инициализация
	log := logger.New()
	cfg := config.New(log)
	
	// 2. Подключение к БД
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSslMode)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 3. Миграция моделей
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Info("Database migration completed")

	// 4. Инициализация слоев (Repository, Service, Handler)
	userRepo := postgr.NewUserRepository(db)
	
	tokenService := service.NewTokenService(cfg)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, tokenService, cfg)
	
	authHandler := http.NewAuthHandler(authService)
	userHandler := http.NewUserHandler(userService)
	
	mw := http.NewMiddleware(tokenService, userService, cfg)
	
	// 5. Настройка Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// 6. Роутинг
	api := e.Group("/api")
	
	// Auth routes
	authGroup := api.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.RefreshToken)
	
	// User routes (защищенные)
	userGroup := api.Group("/users")
	userGroup.Use(mw.AuthMiddleware) // Middleware для проверки авторизации
	
	// Эндпоинты только для админов
	userGroup.POST("", userHandler.CreateUser, mw.AdminMiddleware)
	userGroup.GET("", userHandler.GetAllUsers, mw.AdminMiddleware)
	userGroup.DELETE("/:id", userHandler.DeleteUser, mw.AdminMiddleware)
	
	// Эндпоинты для всех авторизованных пользователей
	userGroup.GET("/:id", userHandler.GetUser)
	userGroup.PUT("/:id", userHandler.UpdateUser)


	// 7. Запуск сервера
	log.Infof("Starting server on port %s", cfg.ServerPort)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}