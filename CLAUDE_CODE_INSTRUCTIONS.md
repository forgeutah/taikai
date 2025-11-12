# Instructions for Claude Code: Building Taikai

## Project Overview

**Repository**: https://github.com/forgeutah/taikai  
**Project Name**: Taikai (Meeting/Event platform for Forge Utah Foundation)  
**Tech Stack**: Go + Chi + HTMX + Alpine.js + Tailwind CSS + PostgreSQL + Redis  
**Purpose**: Self-hosted Meetup.com replacement with parent organization → groups → events hierarchy

## CRITICAL: Read Documentation First

**Before writing ANY code**, you MUST read these three documents in the repository:

1. **docs/PRD.md** - Product Requirements Document
2. **docs/ARCHITECTURE.md** - Architectural Design Document  
3. **docs/IMPLEMENTATION_PLAN.md** - 12-Week Implementation Plan

These documents contain ALL the details you need, including:
- Complete database schema (copy the exact SQL)
- API endpoint specifications
- Permission model (org admin → group admin → event host)
- Subscription deduplication logic
- Auto-subscription for event hosts
- Recurring events (RRULE format)
- Timezone handling patterns

## Setup Instructions for Repository

### Step 1: Add Documentation Files

Create a `docs/` directory in the repository and add these files:
- `docs/PRD.md` (Product Requirements Document)
- `docs/ARCHITECTURE.md` (Architectural Design Document)
- `docs/IMPLEMENTATION_PLAN.md` (Implementation Plan)

### Step 2: Follow the 12-Week Plan

The implementation plan breaks down development into 12 weeks with day-by-day tasks. **Follow this plan sequentially**. Each week builds on the previous week.

## Quick Start Guide for Claude Code

### Week 1, Day 1-2: Environment Setup

Create these foundational files:

1. **docker-compose.yml** (PostgreSQL, Redis, Mailpit for local email testing)
2. **.env.example** (all environment variables with dev defaults)
3. **Makefile** (common tasks: dev, migrate-up, migrate-down, test, build)
4. **go.mod** (`go mod init github.com/forgeutah/taikai`)
5. **README.md** (setup and running instructions)

Install dependencies:
```bash
# Router
go get github.com/go-chi/chi/v5

# Database  
go get github.com/lib/pq
go get github.com/jmoiron/sqlx

# Migrations
go get github.com/pressly/goose/v3

# Background jobs
go get github.com/hibiken/asynq

# JWT
go get github.com/golang-jwt/jwt/v5

# Other essentials (see Architecture Doc for full list)
```

### Week 1, Day 3-5: Database Migrations

Create 10+ migration files in `migrations/` directory using Goose.

**CRITICAL**: Copy the exact SQL from docs/ARCHITECTURE.md for each table. The schema is comprehensive and includes:
- All foreign keys and indexes
- UUID primary keys
- Timezone-aware timestamps
- Recurring event fields (parent_event_id, recurrence_rule, etc.)
- Proper constraints

Migration files needed:
```
migrations/
├── 00001_create_organizations.sql
├── 00002_create_groups.sql
├── 00003_create_users.sql
├── 00004_create_venues.sql
├── 00005_create_events.sql
├── 00006_create_permissions.sql
├── 00007_create_subscriptions.sql
├── 00008_create_rsvps.sql
├── 00009_create_oauth_accounts.sql
├── 00010_create_supporting_tables.sql
└── seed.sql
```

## Critical Requirements to Remember

### 1. Auto-Subscription for Event Hosts

When adding a user as event host:
```
1. Check if user is subscribed to org OR group
2. If NOT subscribed → auto-create group subscription
3. Set default preferences (email=true, SMS=false, Discord=false)
4. Send notification email about auto-subscription
5. Add user to event_hosts table
```

This is MANDATORY. See docs/PRD.md for full requirements.

### 2. Two-Tier Subscription Model

- **Org subscription**: Receives ALL events from ALL groups
- **Group subscription**: Receives only that group's events
- **Deduplication required**: When getting event subscribers, don't send duplicate notifications to users who have both

### 3. Recurring Events (RRULE)

- Use RFC 5545 RRULE format (same as Google Calendar)
- Generate actual event instances (not virtual)
- Support editing: single event, future events, or all events
- Background job maintains 6-month horizon of future events
- See docs/IMPLEMENTATION_PLAN.md Week 5 for full details

### 4. Timezone Handling

- Store all timestamps in UTC (PostgreSQL TIMESTAMP WITH TIME ZONE)
- Store event timezone separately (IANA format: "America/Denver")
- Display times in user's timezone with original timezone noted
- Example: "7:00 PM (your time) • 8:00 PM MST (event time)"

### 5. Permission Hierarchy

```
Org Admin (super user)
├── Can manage organization settings
├── Can create/delete groups
├── Has all group admin permissions for all groups
└── Can view everything

Group Admin
├── Can manage group settings
├── Can create/edit/delete events in their group
├── Can assign event hosts
└── Can view group subscribers

Event Host (scoped permissions)
├── Can edit ONLY their event
├── Can add co-hosts
├── Can view event RSVPs
└── CANNOT delete event or manage group
```

### 6. Community Chat Inheritance

- Groups can set their own community chat URL (Discord, Slack, etc.)
- If NOT set, inherit from organization's community chat URL
- Display on every event page: "Have questions? Join our [Community Name]"

### 7. RSVP Privacy

- Public users see only RSVP count
- Event hosts and group admins see full attendee list (names, emails, timestamps)
- This is enforced at API level

### 8. Capacity Management

- Capacity is optional (NULL = unlimited)
- When set, enforce hard limit (reject RSVPs at capacity)
- Notify hosts at 80% and 100% capacity
- Use database transactions to prevent race conditions

## Project Structure

```
taikai/
├── cmd/
│   ├── server/          # Main HTTP server
│   ├── worker/          # Background job worker
│   └── migrate/         # Migration runner
├── internal/
│   ├── api/             # HTTP handlers and routes
│   ├── auth/            # Authentication & permissions
│   ├── domain/          # Business logic
│   │   ├── organization/
│   │   ├── group/
│   │   ├── event/
│   │   ├── user/
│   │   ├── venue/
│   │   ├── subscription/
│   │   └── rsvp/
│   ├── notification/    # Email notifications
│   ├── storage/         # Database queries
│   └── middleware/      # HTTP middleware
├── migrations/          # Goose migrations
├── pkg/                 # Public packages
│   ├── jwt/
│   ├── email/
│   └── recurrence/      # RRULE handling
├── templates/           # HTML templates
├── static/              # CSS, JS, images
├── docs/                # Documentation
│   ├── PRD.md
│   ├── ARCHITECTURE.md
│   └── IMPLEMENTATION_PLAN.md
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## Development Workflow

1. **Start services**: `docker-compose up -d`
2. **Run migrations**: `make migrate-up`
3. **Run server**: `make dev`
4. **Run tests**: `make test`

## Testing Requirements

- Write tests alongside features
- Aim for 80%+ coverage on core business logic
- Test critical flows:
  - User registration → verification → login
  - Event creation (one-time and recurring)
  - RSVP with capacity enforcement
  - Auto-subscription for hosts
  - Email notifications
  - Permission checks

## API Response Format

```go
// Success
{
  "data": { ... },
  "meta": { ... } // optional
}

// Error  
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": { ... }
  }
}
```

## Common Patterns

### Permission Check Pattern
```go
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    eventID := chi.URLParam(r, "eventId")
    
    if !h.auth.CanManageEvent(userID, eventID) {
        respondError(w, http.StatusForbidden, "insufficient permissions")
        return
    }
    
    // Proceed...
}
```

### Subscription Deduplication Pattern
```go
func GetEventSubscribers(eventID string) ([]User, error) {
    // Get org subscribers (all events)
    orgSubs := getOrgSubscribers(event.Group.OrganizationID)
    
    // Get group subscribers (this group only)
    groupSubs := getGroupSubscribers(event.GroupID)
    
    // Deduplicate (org subscribers take precedence)
    unique := make(map[string]User)
    for _, u := range orgSubs { unique[u.ID] = u }
    for _, u := range groupSubs { 
        if _, exists := unique[u.ID]; !exists {
            unique[u.ID] = u 
        }
    }
    
    return values(unique), nil
}
```

## External Services

### Email (AWS SES)
- Development: Use Mailpit (localhost:1025 SMTP, localhost:8025 web UI)
- Production: AWS SES ($0.10 per 1,000 emails)
- Requires: SPF, DKIM, DMARC DNS records for deliverability

### SMS (Twilio)
- Phase 2 feature (post-launch)
- ~$0.0079 per SMS

## Security Checklist

- [ ] Bcrypt for password hashing (cost factor 12)
- [ ] JWT tokens with short expiry (15 min access, 7 day refresh)
- [ ] SQL injection prevention (prepared statements)
- [ ] XSS prevention (escape all user input)
- [ ] CSRF protection
- [ ] Rate limiting on all endpoints
- [ ] Permission checks on all write operations
- [ ] HTTPS only in production
- [ ] Security headers (CSP, HSTS, X-Frame-Options)

## Performance Targets

- Page load time: < 2 seconds
- API response time: < 500ms (p95)
- Support 100+ concurrent users
- Database queries: < 50ms (p95)

## Launch Readiness Checklist

- [ ] All MVP features implemented
- [ ] Security audit passed
- [ ] Performance targets met
- [ ] All tests passing (80%+ coverage)
- [ ] Works on mobile and desktop
- [ ] Accessibility audit passed (WCAG 2.1 AA)
- [ ] Production deployment ready
- [ ] SSL configured
- [ ] Backups automated
- [ ] Monitoring and alerting set up
- [ ] Documentation complete

## Need Help?

If anything is unclear:
1. Re-read the relevant section in docs/PRD.md, docs/ARCHITECTURE.md, or docs/IMPLEMENTATION_PLAN.md
2. The Implementation Plan has day-by-day tasks with detailed requirements
3. The Architecture Doc has complete database schema and API specs
4. The PRD has all feature requirements and user stories

## Let's Build! 🚀

Start with Week 1, Day 1 of the Implementation Plan. Follow the plan sequentially. Test as you build. You've got this!
