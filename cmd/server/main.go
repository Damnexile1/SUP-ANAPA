package main

import (
	"context"
	"net/http"
	"sup-anapa/internal/config"
	"sup-anapa/internal/handlers"
	"sup-anapa/internal/logger"
	"sup-anapa/internal/middleware"
	"sup-anapa/internal/repository"
	"sup-anapa/internal/services"
	bookinguc "sup-anapa/internal/usecase/booking"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	appLogger, err := logger.Init(cfg.LogFile)
	if err != nil {
		panic(err)
	}
	defer appLogger.Close()

	// Подключение к БД
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		appLogger.Fatal("Unable to connect to database", map[string]interface{}{"error": err.Error()})
	}
	defer pool.Close()

	// Проверка подключения
	if err := pool.Ping(context.Background()); err != nil {
		appLogger.Fatal("Unable to ping database", map[string]interface{}{"error": err.Error()})
	}
	appLogger.Info("Successfully connected to database", nil)

	// Инициализация репозиториев
	bookingRepo := repository.NewBookingRepository(pool)
	instructorRepo := repository.NewInstructorRepository(pool)
	slotRepo := repository.NewSlotRepository(pool)
	walkTypeRepo := repository.NewWalkTypeRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	adminRepo := repository.NewAdminRepository(pool)

	// Инициализация сервисов
	notificationService := services.NewNotificationService(cfg.VKBotToken)
	bookingService := services.NewBookingService(bookingRepo, notificationService)
	weatherService := services.NewWeatherService(cfg.WeatherAPIKey)
	authService := services.NewAuthService(adminRepo)
	userAuthService := services.NewUserAuthService(userRepo)

	// Инициализация handlers
	handlers.Init(cfg.SessionSecret, authService)
	handlers.SetUserAuthService(userAuthService)
	handlers.SetRepositories(bookingRepo, instructorRepo, slotRepo)
	handlers.SetUserRepository(userRepo)
	handlers.SetBookingUseCase(bookinguc.NewCreateBookingUseCase(slotRepo, bookingRepo, userRepo, 20))

	// Инициализация middleware
	middleware.InitAuth(handlers.GetStore())

	// Инициализация API handlers
	instructorHandler := handlers.NewInstructorHandler(instructorRepo)
	slotHandler := handlers.NewSlotHandler(slotRepo, instructorRepo, walkTypeRepo)
	walkTypeHandler := handlers.NewWalkTypeHandler(walkTypeRepo)
	bookingHandler := handlers.NewBookingHandler(bookingRepo, slotRepo, instructorRepo)

	// TODO: Передать сервисы в handlers
	_ = bookingService
	_ = weatherService

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.CorrelationID)
	r.Use(middleware.AccessLog)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Static files
	fileServer := http.FileServer(http.Dir("./web/static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Public routes
	r.Get("/", handlers.Home)
	r.Get("/booking", handlers.BookingPage)
	r.Get("/booking/", handlers.BookingPage)
	r.Get("/user/register", handlers.UserRegisterPage)
	r.Post("/user/register", handlers.UserRegisterPost)
	r.Get("/user/login", handlers.UserLoginPage)
	r.Post("/user/login", handlers.UserLoginPost)
	r.Get("/user/logout", handlers.UserLogout)
	r.Get("/lk", handlers.UserCabinet)
	r.Get("/instructors", handlers.InstructorsPage)
	r.Post("/booking", handlers.CreateBooking)
	r.Post("/booking/", handlers.CreateBooking)
	r.Post("/api/booking", handlers.CreateBooking)
	r.Post("/api/booking/", handlers.CreateBooking)
	r.Get("/favicon.ico", handlers.Favicon)

	// Public API (no auth required)
	r.Get("/api/instructors", instructorHandler.List)
	r.Get("/api/slots", slotHandler.List)
	r.Get("/api/instructors/{id}/walk-types", walkTypeHandler.ListByInstructor)

	// Admin routes
	r.Route("/admin", func(r chi.Router) {
		r.Get("/login", handlers.AdminLogin)
		r.Post("/login", handlers.AdminLoginPost)
		r.Get("/logout", handlers.AdminLogout)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth)
			r.Get("/", handlers.AdminDashboard)
			r.Get("/instructors", handlers.AdminInstructors)
			r.Get("/slots", handlers.AdminSlots)
			r.Get("/walk-types", handlers.AdminWalkTypes)
			r.Get("/walk-types/", handlers.AdminWalkTypes)
			r.Get("/walk_types", handlers.AdminWalkTypes)
			r.Get("/bookings", handlers.AdminBookings)

			// API routes
			r.Route("/api/instructors", func(r chi.Router) {
				r.Get("/", instructorHandler.List)
				r.Post("/", instructorHandler.Create)
				r.Get("/{id}", instructorHandler.Get)
				r.Put("/{id}", instructorHandler.Update)
				r.Delete("/{id}", instructorHandler.Delete)
				r.Get("/{id}/walk-types", walkTypeHandler.ListByInstructor)
			})

			r.Route("/api/walk-types", func(r chi.Router) {
				r.Post("/", walkTypeHandler.Create)
				r.Delete("/{id}", walkTypeHandler.Delete)
			})

			r.Route("/api/slots", func(r chi.Router) {
				r.Get("/", slotHandler.List)
				r.Post("/", slotHandler.Create)
				r.Get("/{id}", slotHandler.Get)
				r.Put("/{id}", slotHandler.Update)
				r.Delete("/{id}", slotHandler.Delete)
			})

			r.Route("/api/bookings", func(r chi.Router) {
				r.Get("/", bookingHandler.List)
				r.Get("/{id}", bookingHandler.Get)
				r.Put("/{id}/status", bookingHandler.UpdateStatus)
			})
		})
	})

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ctx := context.Background()
			n, err := slotRepo.ExpireHolds(ctx)
			if err != nil {
				appLogger.Error("Error expiring slot holds", map[string]interface{}{"error": err.Error()})
			} else if n > 0 {
				appLogger.Info("Expired pending slot holds", map[string]interface{}{"count": n})
			}
		}
	}()

	addr := ":" + cfg.Port
	appLogger.Info("Server starting", map[string]interface{}{"addr": addr, "log_file": cfg.LogFile})
	if err := http.ListenAndServe(addr, r); err != nil {
		appLogger.Fatal("ListenAndServe failed", map[string]interface{}{"error": err.Error()})
	}
}
