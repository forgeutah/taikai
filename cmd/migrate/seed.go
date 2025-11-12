package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	log.Println("Starting database seeding...")

	// Create organization
	orgID := seedOrganization(ctx, db)
	log.Printf("Created organization: %s", orgID)

	// Create admin user
	adminID := seedAdminUser(ctx, db)
	log.Printf("Created admin user: %s", adminID)

	// Make admin user an org admin
	makeOrgAdmin(ctx, db, adminID, orgID)
	log.Println("Assigned org admin role")

	// Create groups
	groupIDs := seedGroups(ctx, db, orgID)
	log.Printf("Created %d groups", len(groupIDs))

	// Create test users
	userIDs := seedTestUsers(ctx, db)
	log.Printf("Created %d test users", len(userIDs))

	// Create venues
	venueIDs := seedVenues(ctx, db)
	log.Printf("Created %d venues", len(venueIDs))

	// Create events
	eventIDs := seedEvents(ctx, db, groupIDs, adminID, venueIDs)
	log.Printf("Created %d events", len(eventIDs))

	log.Println("Database seeding completed successfully!")
}

func seedOrganization(ctx context.Context, db *sql.DB) string {
	query := `
		INSERT INTO organizations (name, slug, description, logo_url, website_url, community_chat_url, community_chat_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id string
	err := db.QueryRowContext(
		ctx, query,
		"Forge Utah Foundation",
		"forge-utah",
		"Forge Utah Foundation is a local tech community in Utah built for the developers, engineers, data scientists-- for the true builders of tech.",
		"https://forgeutah.org/logo.png",
		"https://forgeutah.org",
		"https://discord.gg/forgeutah",
		"Discord",
	).Scan(&id)
	if err != nil {
		log.Fatalf("Failed to create organization: %v", err)
	}
	return id
}

func seedAdminUser(ctx context.Context, db *sql.DB) string {
	// Hash password: "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	query := `
		INSERT INTO users (email, password_hash, name, timezone, email_verified)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id string
	err = db.QueryRowContext(
		ctx, query,
		"admin@forgeutah.org",
		string(hashedPassword),
		"Admin User",
		"America/Denver",
		true,
	).Scan(&id)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}
	return id
}

func makeOrgAdmin(ctx context.Context, db *sql.DB, userID, orgID string) {
	query := `INSERT INTO org_admins (user_id, organization_id) VALUES ($1, $2)`
	_, err := db.ExecContext(ctx, query, userID, orgID)
	if err != nil {
		log.Fatalf("Failed to make user org admin: %v", err)
	}
}

func seedGroups(ctx context.Context, db *sql.DB, orgID string) []string {
	groups := []struct {
		name        string
		slug        string
		description string
	}{
		{"Kubernetes Meetup", "kubernetes", "Monthly meetup for Kubernetes enthusiasts"},
		{"Go User Group", "go-lang", "Go programming language user group"},
		{"Data Engineering", "data-engineering", "Data engineering and analytics community"},
	}

	var ids []string
	for _, g := range groups {
		query := `
			INSERT INTO groups (organization_id, name, slug, description, primary_location)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`
		var id string
		err := db.QueryRowContext(ctx, query, orgID, g.name, g.slug, g.description, "Salt Lake City, UT").Scan(&id)
		if err != nil {
			log.Fatalf("Failed to create group %s: %v", g.name, err)
		}
		ids = append(ids, id)
	}
	return ids
}

func seedTestUsers(ctx context.Context, db *sql.DB) []string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	users := []struct {
		email string
		name  string
	}{
		{"john@example.com", "John Doe"},
		{"jane@example.com", "Jane Smith"},
		{"bob@example.com", "Bob Johnson"},
	}

	var ids []string
	for _, u := range users {
		query := `
			INSERT INTO users (email, password_hash, name, timezone, email_verified)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`
		var id string
		err := db.QueryRowContext(ctx, query, u.email, string(hashedPassword), u.name, "America/Denver", true).Scan(&id)
		if err != nil {
			log.Fatalf("Failed to create user %s: %v", u.name, err)
		}
		ids = append(ids, id)
	}
	return ids
}

func seedVenues(ctx context.Context, db *sql.DB) []string {
	venues := []struct {
		name     string
		address  string
		city     string
		state    string
		timezone string
	}{
		{"DevHub Coworking", "123 Main St", "Salt Lake City", "UT", "America/Denver"},
		{"TechSpace Downtown", "456 State St", "Salt Lake City", "UT", "America/Denver"},
		{"Virtual", "", "", "", "America/Denver"},
	}

	var ids []string
	for _, v := range venues {
		query := `
			INSERT INTO venues (name, address, city, state, timezone, is_virtual)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		isVirtual := v.name == "Virtual"
		var id string
		err := db.QueryRowContext(ctx, query, v.name, v.address, v.city, v.state, v.timezone, isVirtual).Scan(&id)
		if err != nil {
			log.Fatalf("Failed to create venue %s: %v", v.name, err)
		}
		ids = append(ids, id)
	}
	return ids
}

func seedEvents(ctx context.Context, db *sql.DB, groupIDs []string, creatorID string, venueIDs []string) []string {
	if len(groupIDs) == 0 || len(venueIDs) == 0 {
		log.Println("No groups or venues available, skipping event creation")
		return nil
	}

	now := time.Now()
	nextWeek := now.AddDate(0, 0, 7)
	nextMonth := now.AddDate(0, 1, 0)

	events := []struct {
		title       string
		slug        string
		description string
		startTime   time.Time
		endTime     time.Time
		groupID     string
		venueID     string
	}{
		{
			"Kubernetes 1.30: What's New",
			"kubernetes-1-30-whats-new",
			"Join us for an in-depth look at the latest features in Kubernetes 1.30",
			nextWeek,
			nextWeek.Add(2 * time.Hour),
			groupIDs[0],
			venueIDs[0],
		},
		{
			"Go Concurrency Patterns",
			"go-concurrency-patterns",
			"Learn about advanced concurrency patterns in Go",
			nextMonth,
			nextMonth.Add(2 * time.Hour),
			groupIDs[1],
			venueIDs[1],
		},
	}

	var ids []string
	for _, e := range events {
		query := `
			INSERT INTO events (group_id, title, slug, description, status, start_time, end_time, timezone, venue_id, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`
		var id string
		err := db.QueryRowContext(
			ctx, query,
			e.groupID, e.title, e.slug, e.description, "published",
			e.startTime, e.endTime, "America/Denver", e.venueID, creatorID,
		).Scan(&id)
		if err != nil {
			log.Fatalf("Failed to create event %s: %v", e.title, err)
		}
		ids = append(ids, id)

		// Make creator an event host
		_, err = db.ExecContext(ctx, `INSERT INTO event_hosts (user_id, event_id) VALUES ($1, $2)`, creatorID, id)
		if err != nil {
			log.Fatalf("Failed to assign event host: %v", err)
		}
	}
	return ids
}
