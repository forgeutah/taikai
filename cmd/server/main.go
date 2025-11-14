package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/forgeutah/taikai"
	"github.com/forgeutah/taikai/internal/api"
	"github.com/forgeutah/taikai/internal/auth"
	"github.com/forgeutah/taikai/internal/middleware"
	"github.com/forgeutah/taikai/pkg/email"
	"github.com/forgeutah/taikai/pkg/jwt"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services
	jwtManager := initJWTManager()
	emailService := initEmailService()
	blacklist := auth.NewPostgresBlacklist(db)
	permissionChecker := auth.NewPermissionChecker(db)

	// Initialize handlers
	authHandler := api.NewAuthHandler(db, jwtManager, emailService, blacklist)
	userHandler := api.NewUserHandler(db)
	orgHandler := api.NewOrganizationHandler(db, permissionChecker)
	groupHandler := api.NewGroupHandler(db, permissionChecker)
	venueHandler := api.NewVenueHandler(db)
	eventHandler := api.NewEventHandler(db, permissionChecker)
	rsvpHandler := api.NewRSVPHandler(db, permissionChecker)
	subscriptionHandler := api.NewSubscriptionHandler(db, permissionChecker)

	// Initialize upload handler
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	uploadHandler := api.NewUploadHandler(uploadDir, baseURL)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	permissionMiddleware := middleware.NewPermissionMiddleware(permissionChecker)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Serve uploaded files
	fileServer := http.FileServer(http.Dir(uploadDir))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)
		r.Post("/auth/forgot-password", authHandler.ForgotPassword)
		r.Post("/auth/reset-password", authHandler.ResetPassword)
		r.Get("/auth/verify-email", authHandler.VerifyEmail)

		// Public organization routes (read-only, no auth required)
		r.Get("/organizations", orgHandler.ListOrganizations)
		r.Get("/organizations/{orgId}", orgHandler.GetOrganization)
		r.Get("/organizations/slug/{slug}", orgHandler.GetOrganizationBySlug)
		r.Get("/organizations/{orgId}/admins", orgHandler.GetOrgAdmins)

		// Public group routes (read-only, no auth required)
		r.Get("/groups", groupHandler.ListGroups)
		r.Get("/groups/{groupId}", groupHandler.GetGroup)
		r.Get("/groups/slug/{slug}", groupHandler.GetGroupBySlug)
		r.Get("/groups/{groupId}/admins", groupHandler.GetGroupAdmins)

		// Public venue routes (read-only, no auth required)
		r.Get("/venues", venueHandler.ListVenues)
		r.Get("/venues/{venueId}", venueHandler.GetVenue)

		// Public event routes (read-only, no auth required)
		r.Get("/events", eventHandler.ListEvents)
		r.Get("/events/{eventId}", eventHandler.GetEvent)
		r.Get("/events/slug/{slug}", eventHandler.GetEventBySlug)
		r.Get("/events/{eventId}/hosts", eventHandler.GetEventHosts)
		r.Get("/events/{eventId}/speakers", eventHandler.GetEventSpeakers)
		r.Get("/events/{eventId}/sponsors", eventHandler.GetEventSponsors)
		r.Get("/events/{eventId}/schedule", eventHandler.GetEventSchedule)

		// Public RSVP routes (count public, details for hosts/admins)
		r.Get("/events/{eventId}/rsvps", rsvpHandler.GetEventRSVPs)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			// User routes
			r.Get("/users/me", userHandler.GetMe)
			r.Patch("/users/me", userHandler.UpdateMe)
			r.Delete("/users/me", userHandler.DeleteMe)

			// Auth routes that require authentication
			r.Post("/auth/logout", authHandler.Logout)

			// File upload routes (authenticated users only)
			r.Post("/upload/image", uploadHandler.UploadImage)

			// Organization management routes (org admin only)
			r.With(permissionMiddleware.RequireOrgAdmin).Patch("/organizations/{orgId}", orgHandler.UpdateOrganization)
			r.With(permissionMiddleware.RequireOrgAdmin).Post("/organizations/{orgId}/admins", orgHandler.AddOrgAdmin)
			r.With(permissionMiddleware.RequireOrgAdmin).Delete("/organizations/{orgId}/admins/{userId}", orgHandler.RemoveOrgAdmin)

			// Group management routes
			r.With(permissionMiddleware.RequireOrgAdmin).Post("/groups", groupHandler.CreateGroup)
			r.With(permissionMiddleware.RequireGroupAdmin).Patch("/groups/{groupId}", groupHandler.UpdateGroup)
			r.With(permissionMiddleware.RequireOrgAdmin).Delete("/groups/{groupId}", groupHandler.DeleteGroup)
			r.With(permissionMiddleware.RequireGroupAdmin).Post("/groups/{groupId}/admins", groupHandler.AddGroupAdmin)
			r.With(permissionMiddleware.RequireGroupAdmin).Delete("/groups/{groupId}/admins/{userId}", groupHandler.RemoveGroupAdmin)

			// Venue management routes (any group admin can manage venues)
			r.With(permissionMiddleware.RequireAnyGroupAdmin).Post("/venues", venueHandler.CreateVenue)
			r.With(permissionMiddleware.RequireAnyGroupAdmin).Patch("/venues/{venueId}", venueHandler.UpdateVenue)

			// Event management routes
			r.With(permissionMiddleware.RequireGroupAdmin).Post("/events", eventHandler.CreateEvent)
			r.Post("/events/preview-recurrence", eventHandler.PreviewRecurrence) // Preview recurring events
			r.With(permissionMiddleware.RequireEventManagement).Patch("/events/{eventId}", eventHandler.UpdateEvent)
			r.With(permissionMiddleware.RequireGroupAdmin).Delete("/events/{eventId}", eventHandler.DeleteEvent)

			// Event hosts management
			r.With(permissionMiddleware.RequireEventManagement).Post("/events/{eventId}/hosts", eventHandler.AddEventHost)
			r.With(permissionMiddleware.RequireEventManagement).Delete("/events/{eventId}/hosts/{userId}", eventHandler.RemoveEventHost)

			// Event speakers management
			r.With(permissionMiddleware.RequireEventManagement).Post("/events/{eventId}/speakers", eventHandler.CreateSpeaker)
			r.With(permissionMiddleware.RequireEventManagement).Patch("/events/{eventId}/speakers/{speakerId}", eventHandler.UpdateSpeaker)
			r.With(permissionMiddleware.RequireEventManagement).Delete("/events/{eventId}/speakers/{speakerId}", eventHandler.DeleteSpeaker)

			// Event sponsors management
			r.With(permissionMiddleware.RequireEventManagement).Post("/events/{eventId}/sponsors", eventHandler.CreateSponsor)
			r.With(permissionMiddleware.RequireEventManagement).Patch("/events/{eventId}/sponsors/{sponsorId}", eventHandler.UpdateSponsor)
			r.With(permissionMiddleware.RequireEventManagement).Delete("/events/{eventId}/sponsors/{sponsorId}", eventHandler.DeleteSponsor)

			// Event schedule management
			r.With(permissionMiddleware.RequireEventManagement).Post("/events/{eventId}/schedule", eventHandler.CreateScheduleItem)
			r.With(permissionMiddleware.RequireEventManagement).Patch("/events/{eventId}/schedule/{scheduleId}", eventHandler.UpdateScheduleItem)
			r.With(permissionMiddleware.RequireEventManagement).Delete("/events/{eventId}/schedule/{scheduleId}", eventHandler.DeleteScheduleItem)

			// RSVP management
			r.Post("/events/{eventId}/rsvp", rsvpHandler.CreateOrUpdateRSVP)
			r.Delete("/events/{eventId}/rsvp", rsvpHandler.DeleteRSVP)

			// User dashboard routes
			r.Get("/me/rsvps", rsvpHandler.GetMyRSVPs)
			r.Get("/me/events", rsvpHandler.GetMyHostedEvents)
			r.Get("/me/subscriptions", subscriptionHandler.GetMySubscriptions)

			// Organization subscription routes
			r.Post("/organizations/{orgId}/subscribe", subscriptionHandler.SubscribeToOrganization)
			r.Delete("/organizations/{orgId}/subscribe", subscriptionHandler.UnsubscribeFromOrganization)
			r.Patch("/organizations/{orgId}/subscribe", subscriptionHandler.UpdateOrgSubscription)
			r.With(permissionMiddleware.RequireOrgAdmin).Get("/organizations/{orgId}/subscribers", subscriptionHandler.GetOrgSubscribers)

			// Group subscription routes
			r.Post("/groups/{groupId}/subscribe", subscriptionHandler.SubscribeToGroup)
			r.Delete("/groups/{groupId}/subscribe", subscriptionHandler.UnsubscribeFromGroup)
			r.Patch("/groups/{groupId}/subscribe", subscriptionHandler.UpdateGroupSubscription)
			r.With(permissionMiddleware.RequireGroupAdmin).Get("/groups/{groupId}/subscribers", subscriptionHandler.GetGroupSubscribers)
		})
	})

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("🚀 Taikai server starting on port %s", port)
	log.Printf("📧 Email service configured for: %s", os.Getenv("SMTP_HOST"))
	log.Printf("🗄️  Database: Connected")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped")
}

func initDatabase() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func initJWTManager() *jwt.Manager {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Get token durations from env or use defaults
	accessTokenDuration := 15 * time.Minute  // 15 minutes
	refreshTokenDuration := 7 * 24 * time.Hour // 7 days

	return jwt.NewManager(secretKey, accessTokenDuration, refreshTokenDuration)
}

func initEmailService() *email.Service {
	config := &email.Config{
		SMTPHost:  os.Getenv("SMTP_HOST"),
		SMTPPort:  os.Getenv("SMTP_PORT"),
		FromEmail: os.Getenv("SMTP_FROM_EMAIL"),
		FromName:  os.Getenv("SMTP_FROM_NAME"),
		// For production with AWS SES or similar
		// SMTPUsername: os.Getenv("SMTP_USERNAME"),
		// SMTPPassword: os.Getenv("SMTP_PASSWORD"),
	}

	if config.SMTPHost == "" {
		config.SMTPHost = "localhost"
	}
	if config.SMTPPort == "" {
		config.SMTPPort = "1025"
	}
	if config.FromEmail == "" {
		config.FromEmail = "noreply@taikai.local"
	}
	if config.FromName == "" {
		config.FromName = "Taikai"
	}

	return email.NewService(config)
}

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(taikai.Migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}
