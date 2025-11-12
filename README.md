# Taikai - Open Source Community & Events Platform

Taikai is an open-source event management platform designed for technology communities. It provides a self-hosted alternative to Meetup.com with features tailored for parent organizations managing multiple user groups.

**Taikai** (大会) is a Japanese word meaning "large meeting" or "convention."

## Features

- **Hierarchical Organization Structure**: Parent organization → User groups → Events
- **Flexible Subscription Model**: Subscribe to entire organization or specific groups
- **Recurring Events**: Full support for recurring events with RRULE format
- **RSVP Management**: Capacity enforcement, waitlists, and attendance tracking
- **Multi-Channel Notifications**: Email, SMS (Phase 2), Discord (Phase 2)
- **Calendar Integration**: ICS export and calendar feeds
- **Permission System**: Org admins, group admins, and event hosts with scoped permissions
- **Community Integration**: Discord/Slack chat links, YouTube live streaming

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Chi (HTTP router)
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Background Jobs**: Asynq (Redis-backed)
- **Migrations**: Goose

### Frontend
- **Rendering**: Server-side with Go templates
- **Enhancement**: HTMX + Alpine.js
- **Styling**: Tailwind CSS

### External Services
- **Email**: AWS SES (Development: Mailpit)
- **SMS**: Twilio (Phase 2)

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, but recommended)
- PostgreSQL 15+ (via Docker)
- Redis 7+ (via Docker)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/forgeutah/taikai.git
cd taikai
```

### 2. Copy Environment File

```bash
cp .env.example .env
```

Edit `.env` if you need to customize any settings. Default values work for local development.

### 3. Install Dependencies

```bash
make install-deps
make install-tools
```

This will install:
- Go dependencies (Chi, sqlx, goose, asynq, etc.)
- Development tools (goose, sqlc)

### 4. Start Docker Services

```bash
make docker-up
```

This starts:
- PostgreSQL on port 5432
- Redis on port 6379
- Mailpit (email testing) on ports 1025 (SMTP) and 8025 (Web UI)

### 5. Run Database Migrations

```bash
make migrate-up
```

This creates all necessary database tables.

### 6. Seed the Database (Optional)

```bash
make seed
```

This populates the database with:
- 1 Organization (Forge Utah Foundation)
- 3 Groups (Kubernetes, Go, Data Engineering)
- 1 Admin user (admin@forgeutah.org / password: admin123)
- 3 Test users
- 3 Venues
- 2 Sample events

### 7. Start the Development Server

```bash
make dev
```

The application will be available at http://localhost:8080

## Development Workflow

### Running the Server

```bash
make dev
```

### Running Database Migrations

Create a new migration:
```bash
make migrate-create NAME=add_new_feature
```

Run migrations:
```bash
make migrate-up
```

Rollback last migration:
```bash
make migrate-down
```

Check migration status:
```bash
make migrate-status
```

### Code Generation with sqlc

After modifying SQL queries in `internal/storage/queries/*.sql`:

```bash
make sqlc-generate
```

### Running Tests

```bash
make test
```

With coverage report:
```bash
make test-coverage
open coverage.html
```

### Viewing Emails (Development)

Mailpit web interface: http://localhost:8025

All emails sent during development are captured here.

### Formatting and Linting

Format code:
```bash
make fmt
```

Run linter:
```bash
make lint
```

Run go vet:
```bash
make vet
```

## Project Structure

```
taikai/
├── cmd/
│   ├── server/          # Main HTTP server
│   ├── worker/          # Background job worker
│   └── migrate/         # Migration runner & seeder
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
│   ├── storage/         # Database queries (sqlc generated)
│   └── middleware/      # HTTP middleware
├── migrations/          # Database migrations
├── templates/           # HTML templates
├── static/              # CSS, JS, images
├── pkg/                 # Public packages
│   ├── jwt/
│   ├── email/
│   └── recurrence/      # RRULE handling
├── docs/                # Documentation
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

## Configuration

All configuration is done via environment variables. See `.env.example` for all available options.

### Key Environment Variables

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `JWT_SECRET`: Secret key for JWT token signing (change in production!)
- `SMTP_HOST`, `SMTP_PORT`: Email server configuration
- `APP_PORT`: HTTP server port (default: 8080)

## Database Schema

The database includes the following main tables:

- `organizations`: Parent organizations
- `groups`: User groups within organizations
- `users`: User accounts
- `events`: Events (with recurring event support)
- `venues`: Event venues
- `org_admins`, `group_admins`, `event_hosts`: Permission tables
- `org_subscriptions`, `group_subscriptions`: Subscription management
- `event_rsvps`: RSVP tracking
- `event_speakers`, `event_schedules`: Event details

See `migrations/` for complete schema definitions.

## Testing

### Default Test Credentials

After running `make seed`:

**Admin User:**
- Email: admin@forgeutah.org
- Password: admin123

**Test Users:**
- john@example.com / password123
- jane@example.com / password123
- bob@example.com / password123

## Deployment

See `docs/DEPLOYMENT.md` for production deployment instructions.

## Contributing

Contributions are welcome! Please see `CONTRIBUTING.md` for guidelines.

## License

MIT License - see `LICENSE` file for details.

## History

Forge Utah Foundation is a local tech community in Utah built for developers, engineers, and data scientists. We host several free and open developer groups and were big users of Meetup.com. Unfortunately, Meetup doesn't provide discounts for not-for-profit organizations, and the cost became prohibitive. Since we are a group of technologists, we thought "let's build our own!" Hopefully, other communities will find it beneficial as well.

## Support

- GitHub Issues: https://github.com/forgeutah/taikai/issues
- Documentation: https://docs.forgeutah.org/taikai
- Community: https://discord.gg/forgeutah

## Roadmap

### Phase 1 (MVP - Current)
- ✅ Project setup and infrastructure
- ✅ Database schema and migrations
- 🚧 User authentication and profiles
- 🚧 Organization and group management
- 🚧 Event creation and management
- 🚧 RSVP system
- 🚧 Email notifications

### Phase 2 (Post-Launch)
- SMS notifications (Twilio)
- OAuth (Google, GitHub)
- Discord integration
- Calendar feed subscriptions

### Phase 3 (Future)
- Mobile apps (iOS/Android)
- Advanced analytics
- Event check-in system (QR codes)
- Waitlist functionality

## Acknowledgments

Built with ❤️ by the Forge Utah Foundation community.

Special thanks to all contributors and the open-source community.
