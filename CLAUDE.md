# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Taikai** is an open-source, self-hosted event management platform for technology communities, built as an alternative to Meetup.com. It supports a hierarchical structure: Organizations → User Groups → Events, with flexible subscriptions, recurring events, and comprehensive permission management.

**Tech Stack:**
- Backend: Go 1.21+, Chi router, PostgreSQL 17+
- Database: PostgreSQL with goose migrations (embedded, auto-run on startup), sqlc for type-safe queries
- Frontend: Server-side Go templates, HTMX, Alpine.js, Tailwind CSS
- Email: AWS SES (dev: Mailpit)

## Development Commands

### Setup and Running
```bash
# Complete setup (Docker services + migrations)
make setup

# Start Docker services only (PostgreSQL, Mailpit)
make docker-up

# Run database migrations
make migrate-up

# Seed database with test data
make seed

# Start development server (http://localhost:8080)
make dev

# Start background worker (for recurring events, notifications)
make worker

# View development emails (Mailpit UI)
open http://localhost:8025
```

### Database Operations
```bash
# Create new migration
make migrate-create NAME=add_feature_name

# Check migration status
make migrate-status

# Rollback last migration
make migrate-down

# Generate Go code from SQL queries (after modifying internal/storage/queries/*.sql)
make sqlc-generate
```

### Testing and Quality
```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage
# Then open coverage.html in browser

# Run single test
go test -v ./internal/api -run TestCreateOrUpdateRSVP_Success

# Run tests for specific package
go test -v ./internal/auth

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Run go vet
make vet
```

### Building
```bash
# Build all binaries (creates bin/server, bin/worker, bin/migrate)
make build

# Clean build artifacts
make clean
```

## Architecture

### Project Structure
```
taikai/
├── cmd/
│   ├── server/          # Main HTTP server entrypoint
│   ├── worker/          # Background job worker (not yet implemented)
│   └── migrate/         # Database seeder
├── internal/            # Private application code
│   ├── api/             # HTTP handlers (handlers grouped by resource)
│   ├── auth/            # Authentication, JWT, permissions, password hashing
│   ├── middleware/      # HTTP middleware (auth, permissions)
│   └── storage/         # Database layer
│       └── queries/     # SQL queries for sqlc generation
├── pkg/                 # Public/reusable packages
│   ├── email/           # Email service abstraction
│   ├── jwt/             # JWT token management
│   └── recurrence/      # RRULE parsing and event generation
├── migrations/          # Goose SQL migrations
├── templates/           # HTML templates (if using server-side rendering)
├── static/              # Static assets (CSS, JS, images)
└── uploads/             # User-uploaded files (images, etc.)
```

### Database Architecture

**Core Entity Hierarchy:**
- Organizations (parent orgs like "Forge Utah Foundation")
- Groups (user groups within org: "Kubernetes Meetup", "Go Meetup")
- Events (one-time or recurring)
- Users
- Venues

**Permission Hierarchy:**
- Org Admins: Full control over organization and all groups
- Group Admins: Manage specific group(s) and create events
- Event Hosts: Edit only their assigned events (scoped permissions)

**Subscription Model (Two-Tier):**
- Org Subscriptions: Receive notifications for ALL events in ALL groups
- Group Subscriptions: Receive notifications for specific group's events
- **Critical:** Deduplication required when sending notifications (users with org subscription shouldn't receive duplicate group notifications)

**Key Tables:**
- `organizations`, `groups`, `events`, `users`, `venues`
- `org_admins`, `group_admins`, `event_hosts` (permission tables)
- `org_subscriptions`, `group_subscriptions` (subscription management)
- `event_rsvps` (RSVP tracking with capacity enforcement)
- `event_speakers`, `event_schedules`, `event_sponsors` (event details)
- `oauth_accounts` (OAuth integration, Phase 2)

### Code Generation with sqlc

All database queries are written in `internal/storage/queries/*.sql` and type-safe Go code is generated using sqlc. After modifying queries, run:
```bash
make sqlc-generate
```

Configuration: `sqlc.yaml`

### Recurring Events (RRULE)

Recurring events use RFC 5545 RRULE format (same as Google Calendar):
- Store recurrence rule in `events.recurrence_rule` field
- Generate actual event instances (not virtual) via background job
- Support editing: single occurrence, future occurrences, or all occurrences
- Background worker maintains 6-month horizon of future event instances
- Implementation: `pkg/recurrence/rrule.go`

**When creating/editing recurring events:**
1. Validate RRULE format
2. Generate preview of occurrences (use `PreviewRecurrence` endpoint)
3. Create parent event with recurrence_rule
4. Background job generates child events with `parent_event_id` reference

### Permission System

**Permission Checks Pattern:**
All write operations require permission validation. Use `auth.PermissionChecker`:

```go
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    eventID := chi.URLParam(r, "eventId")

    if !h.checker.CanManageEvent(userID, eventID) {
        respondError(w, http.StatusForbidden, "insufficient permissions")
        return
    }
    // Proceed with update
}
```

**Middleware Usage:**
Routes use permission middleware in `cmd/server/main.go`:
- `RequireOrgAdmin`: Organization management
- `RequireGroupAdmin`: Group management, event creation
- `RequireEventManagement`: Event/speaker/schedule editing (hosts or admins)
- `RequireAnyGroupAdmin`: Venue management (any group admin)

### Auto-Subscription for Event Hosts

**CRITICAL REQUIREMENT:** When adding a user as event host:
1. Check if user has org OR group subscription
2. If NO subscription → auto-create group subscription
3. Set default preferences: `email=true, sms=false, discord=false`
4. Send notification about auto-subscription
5. Add to `event_hosts` table

This ensures event hosts always receive notifications about their events.

### Testing Patterns

Tests use `go-sqlmock` for database mocking. Pattern:
```go
func TestCreateRSVP_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    // Set up expected queries and results
    mock.ExpectQuery("SELECT capacity FROM events").
        WithArgs(eventID).
        WillReturnRows(sqlmock.NewRows([]string{"capacity"}).AddRow(10))

    // Execute handler
    handler := NewRSVPHandler(db, permChecker)
    // ... test execution

    // Verify all expectations met
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

See `internal/api/rsvp_handler_test.go` for comprehensive examples.

### API Response Format

**Success:**
```json
{
  "data": { ... },
  "meta": { ... }  // optional
}
```

**Error:**
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": { ... }
  }
}
```

Helper functions in `internal/api/response.go`.

## Important Implementation Details

### Timezone Handling
- Store all timestamps in UTC (`TIMESTAMP WITH TIME ZONE`)
- Store event timezone separately in IANA format (`America/Denver`)
- Display times in user's timezone with original timezone noted
- Example: "7:00 PM (your time) • 8:00 PM MST (event time)"

### Capacity Management
- Capacity is optional (`NULL` = unlimited)
- When set, enforce hard limit (reject RSVPs when at capacity)
- Use database transactions to prevent race conditions
- Notify event hosts at 80% and 100% capacity

### Subscription Deduplication
When fetching event subscribers for notifications:
```go
// Get org subscribers (receive ALL events)
orgSubs := getOrgSubscribers(event.Group.OrganizationID)

// Get group subscribers (receive only this group's events)
groupSubs := getGroupSubscribers(event.GroupID)

// Deduplicate (org subscribers take precedence)
unique := make(map[string]User)
for _, u := range orgSubs { unique[u.ID] = u }
for _, u := range groupSubs {
    if _, exists := unique[u.ID]; !exists {
        unique[u.ID] = u
    }
}
```

### Community Chat Inheritance
- Groups can set their own community chat URL (Discord, Slack)
- If NOT set, inherit from organization's community chat URL
- Display on every event page for community engagement

### RSVP Privacy
- Public users: See only RSVP count
- Event hosts & group admins: See full attendee list with names, emails, timestamps
- Enforced at API level in `GetEventRSVPs` handler

## Development Workflow

1. **Read documentation first:** `docs/CLAUDE_CODE_INSTRUCTIONS.md`, `docs/PRD.md`, `docs/ARCHITECTURE.md`
2. **Start services:** `make docker-up`
3. **Run migrations:** `make migrate-up`
4. **Seed test data:** `make seed` (creates admin user, test users, sample events)
5. **Start server:** `make dev`
6. **Write tests alongside features** (aim for 80%+ coverage on business logic)
7. **After modifying SQL queries:** `make sqlc-generate`
8. **Before committing:** `make fmt && make test`

## Test Data (after make seed)

**Admin User:**
- Email: admin@forgeutah.org
- Password: admin123

**Test Users:**
- john@example.com / password123
- jane@example.com / password123
- bob@example.com / password123

**Test Organization:** Forge Utah Foundation
**Test Groups:** Kubernetes, Go, Data Engineering

## Environment Variables

Key variables (see `.env.example` for full list):
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT signing (change in production!)
- `SMTP_HOST`, `SMTP_PORT`: Email server (dev: Mailpit on localhost:1025)
- `APP_PORT`: HTTP server port (default: 8080)
- `UPLOAD_DIR`: Directory for uploaded files (default: ./uploads)
- `BASE_URL`: Base URL for the application (for email links)

## Common Gotchas

1. **Permission checks:** Always verify permissions before write operations
2. **Transaction safety:** Use transactions for RSVP capacity enforcement
3. **Auto-subscriptions:** Don't forget to auto-subscribe event hosts
4. **Deduplication:** Don't send duplicate notifications to users with org subscriptions
5. **Timezone handling:** Always store in UTC, display in user/event timezone
6. **sqlc regeneration:** Run `make sqlc-generate` after modifying query files
7. **Migration ordering:** Migrations must maintain referential integrity

## Security Checklist

- Bcrypt password hashing (cost factor 12)
- JWT tokens with short expiry (15 min access, 7 day refresh)
- SQL injection prevention (sqlc uses prepared statements)
- Permission checks on all write operations
- Input validation and sanitization
- CORS configuration for production
- Rate limiting (to be implemented)

## Resources

- Chi Router: https://go-chi.io
- sqlc: https://sqlc.dev
- Goose Migrations: https://github.com/pressly/goose
- RRULE Spec: https://icalendar.org/iCalendar-RFC-5545/3-8-5-3-recurrence-rule.html
- Project Docs: `/docs` directory in repository
