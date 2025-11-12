# Architectural Design Document: Ford Utah Foundation Meetup Platform

## Document Information

**Version:** 0.1 (Draft)  
**Last Updated:** 2025-11-12  
**Status:** In Progress  
**Related Documents:** [Product Requirements Document](./ford-utah-meetup-prd.md)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Architecture Overview](#system-architecture-overview)
3. [Technology Stack](#technology-stack)
4. [Database Schema](#database-schema)
5. [API Design](#api-design)
6. [Authentication & Authorization](#authentication--authorization)
7. [Notification System](#notification-system)
8. [Frontend Architecture](#frontend-architecture)
9. [Infrastructure & Deployment](#infrastructure--deployment)
10. [Security Considerations](#security-considerations)
11. [Performance & Scalability](#performance--scalability)
12. [Third-Party Integrations](#third-party-integrations)

---

## Executive Summary

This document outlines the technical architecture for a self-hosted event management platform designed for the Ford Utah Foundation. The system follows a modern web application architecture with a RESTful API backend, PostgreSQL database, and responsive web frontend. The platform is designed to be self-hosted but architected for scalability and maintainability.

### Key Architectural Decisions

- **Monolithic Backend with API-First Design**: Single application server with clear API boundaries for future microservices migration if needed
- **PostgreSQL Database**: Leverages relational integrity for hierarchical organization structure and complex permission models
- **Server-Side Rendering + Progressive Enhancement**: Fast initial loads with JavaScript enhancement for interactivity
- **Background Job Processing**: Asynchronous handling of notifications and long-running tasks
- **Stateless API**: JWT-based authentication for horizontal scaling capability

---

## System Architecture Overview

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Web App    │  │ Mobile App   │  │  Calendar    │         │
│  │  (Browser)   │  │   (Future)   │  │ Clients (ICS)│         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼──────────────────┼──────────────────┼────────────────┘
          │                  │                  │
          └──────────────────┼──────────────────┘
                             │ HTTPS
          ┌──────────────────▼──────────────────┐
          │         API GATEWAY / LOAD          │
          │            BALANCER                  │
          └──────────────────┬──────────────────┘
                             │
          ┌──────────────────▼──────────────────┐
          │                                      │
          │    APPLICATION SERVER (Backend)      │
          │                                      │
          │  ┌────────────────────────────────┐ │
          │  │    API Layer (REST/JSON)       │ │
          │  └────────────┬───────────────────┘ │
          │               │                      │
          │  ┌────────────▼───────────────────┐ │
          │  │   Business Logic Layer         │ │
          │  │  - Event Management            │ │
          │  │  - User Management             │ │
          │  │  - Subscription Management     │ │
          │  │  - Permission Management       │ │
          │  └────────────┬───────────────────┘ │
          │               │                      │
          │  ┌────────────▼───────────────────┐ │
          │  │   Data Access Layer (ORM)      │ │
          │  └────────────┬───────────────────┘ │
          └───────────────┼──────────────────────┘
                          │
          ┌───────────────▼──────────────────────┐
          │     PRIMARY DATABASE (PostgreSQL)    │
          │  - Relational data                   │
          │  - User accounts, Events, Groups     │
          └──────────────────────────────────────┘

          ┌──────────────────────────────────────┐
          │     CACHE LAYER (Redis)              │
          │  - Session storage                   │
          │  - Rate limiting                     │
          │  - Event listing cache               │
          └──────────────────────────────────────┘

          ┌──────────────────────────────────────┐
          │   BACKGROUND JOB PROCESSOR           │
          │   (Sidekiq/Celery/Bull)              │
          │                                      │
          │  - Email notifications               │
          │  - SMS notifications                 │
          │  - Scheduled reminders               │
          │  - Capacity notifications            │
          └─────────┬────────────────────────────┘
                    │
          ┌─────────▼────────────────────────────┐
          │   MESSAGE QUEUE (Redis/PostgreSQL)   │
          └──────────────────────────────────────┘

          ┌──────────────────────────────────────┐
          │   EXTERNAL SERVICES                  │
          │                                      │
          │  - Email (SendGrid/Mailgun/SES)     │
          │  - SMS (Twilio)                      │
          │  - OAuth (Google, GitHub)            │
          │  - YouTube (embed)                   │
          └──────────────────────────────────────┘
```

### Component Responsibilities

#### API Gateway / Load Balancer
- SSL/TLS termination
- Rate limiting (first line of defense)
- Request routing
- Static asset serving (if not using CDN)

#### Application Server
- RESTful API endpoints
- Business logic execution
- Permission checking
- Session management
- Input validation
- Response formatting

#### Database (PostgreSQL)
- Primary data store
- ACID transactions
- Complex relational queries
- Full-text search capabilities

#### Cache Layer (Redis)
- Session storage
- Rate limiting counters
- Frequently accessed data (event listings, user counts)
- Temporary data (email verification tokens)

#### Background Job Processor
- Asynchronous task execution
- Scheduled tasks (event reminders)
- Notification dispatch
- Long-running operations

---

## Technology Stack

### Backend

**Primary Framework: Go with Chi or Echo** (Selected)
- **Rationale**: 
  - Excellent performance and low resource usage (critical for self-hosting)
  - Strong concurrency model for handling notifications and background tasks
  - Self-contained binary deployment (easy to deploy)
  - Matches your expertise (Twitch bot, Pedro development)
  - Chi: Lightweight, idiomatic, good middleware ecosystem
  - Echo: Slightly more features out-of-box, excellent performance
- **Router**: Chi (more idiomatic) or Echo (more batteries-included) - your preference
- **Database Migrations**: Goose (simple, Go-native, supports both SQL and Go migrations)
- **ORM Option**: sqlc (type-safe SQL), GORM (if you prefer ORM), or sqlx (lightweight)

**Why NOT Gin:**
- While performant, Chi and Echo are more actively maintained
- Chi is more idiomatic Go
- Echo has better built-in validation and error handling

**Migration Tool: Goose**
```bash
# Example migration structure
migrations/
  00001_create_organizations.sql
  00002_create_groups.sql
  00003_create_users.sql
  00004_create_events.sql
```

### Database

**Primary Database: PostgreSQL 15+**
- Strong relational integrity for hierarchical data
- JSON/JSONB support for flexible fields
- Full-text search for event discovery
- Excellent timezone support
- Mature replication and backup solutions

### Caching

**Redis 7+**
- Session management
- Rate limiting
- Query result caching
- Background job queue

### Background Jobs

**Recommended: Asynq (Go-native)**
- Built specifically for Go
- Redis-backed job queue
- Excellent for distributed systems
- Built-in retry logic and monitoring
- Web UI for job monitoring (asynqmon)
- Perfect fit with Chi/Echo

**Alternative: River (PostgreSQL-backed)**
- Uses PostgreSQL as job queue (reduces dependencies)
- Native Go implementation
- Good if you want to avoid Redis dependency
- Simpler operational overhead

**Alternative: Machinery**
- More mature but heavier
- Supports multiple brokers (Redis, AMQP)

**Recommendation**: **Asynq** is the best choice for your use case - modern, performant, and Go-native with excellent monitoring.

### Frontend

**Recommended: Server-Side Rendering with Templ (Go) or Django Templates**
- Fast initial page loads
- SEO-friendly
- Progressive enhancement with Alpine.js or HTMX
- Reduced JavaScript complexity

**Alternative: React SPA with Next.js**
- Rich interactivity
- Strong ecosystem
- Server-side rendering option
- Better for future mobile app (React Native)

**Styling: Tailwind CSS**
- Rapid UI development
- Consistent design system
- Luma-inspired aesthetic achievable
- Excellent accessibility utilities

### Authentication

**JWT (JSON Web Tokens)**
- Stateless authentication
- Easy to scale horizontally
- Include user ID, roles, and permissions in payload

**OAuth 2.0 Integration**
- Google OAuth
- GitHub OAuth
- Delegate to established providers

### Email

**Recommended: AWS SES** (Most Cost-Effective)
- Extremely affordable: $0.10 per 1,000 emails
- High deliverability rates
- No monthly minimums
- Easy integration with Go AWS SDK
- Production-ready for self-hosting

**Alternative: Self-Hosted SMTP**
- Postfix or OpenSMTPD
- Free but requires careful configuration for deliverability
- Risk of emails marked as spam without proper SPF/DKIM/DMARC
- More maintenance overhead

**Development/Testing:**
- Mailpit (modern) or Mailhog - local SMTP testing with web UI
- Both are open source and perfect for development

**Not Recommended:**
- SendGrid: Requires paid plan for production volume
- Mailgun: More expensive than SES

### SMS

**Twilio**
- Industry standard
- Reliable delivery
- Reasonable pricing
- Good API

### File Storage

**Local Filesystem (MVP)**
- Images: user avatars, group logos, event images
- Organized directory structure

**Future: S3-compatible Object Storage**
- Backblaze B2, AWS S3, MinIO
- Scalability and backups

---

## Database Schema

### Design Principles

1. **Hierarchical Integrity**: Enforce organization → group → event hierarchy at database level
2. **Subscription Flexibility**: Support both org-level and group-level subscriptions
3. **Permission Scoping**: Track permissions at different levels (org admin, group admin, event host)
4. **Audit Trail**: Track created_at, updated_at for all entities
5. **Soft Deletes**: Use soft deletes for critical data (users, events) to maintain referential integrity
6. **Timezone Awareness**: Store all timestamps in UTC, store timezone separately

### Entity Relationship Diagram

```
┌──────────────────┐
│  organizations   │
├──────────────────┤
│ id (PK)          │
│ name             │
│ slug (unique)    │
│ description      │
│ logo_url         │
│ website_url      │
│ community_chat_url│
│ community_chat_name│
│ created_at       │
│ updated_at       │
└────────┬─────────┘
         │ 1:N
         ▼
┌──────────────────┐
│     groups       │
├──────────────────┤
│ id (PK)          │
│ organization_id (FK)│
│ name             │
│ slug (unique)    │
│ description      │
│ logo_url         │
│ primary_location │
│ community_chat_url│ (nullable, overrides org)
│ community_chat_name│ (nullable)
│ created_at       │
│ updated_at       │
└────────┬─────────┘
         │ 1:N
         ▼
┌──────────────────┐
│     events       │
├──────────────────┤
│ id (PK)          │
│ group_id (FK)    │
│ title            │
│ slug (unique)    │
│ description      │
│ status (enum)    │ draft, published, cancelled
│ start_time (UTC) │
│ end_time (UTC)   │
│ timezone         │ IANA timezone
│ venue_id (FK)    │
│ capacity (nullable)│
│ youtube_url (nullable)│
│ created_by (FK → users)│
│ created_at       │
│ updated_at       │
└──────────────────┘

┌──────────────────┐
│     venues       │
├──────────────────┤
│ id (PK)          │
│ name             │
│ address          │
│ city             │
│ state            │
│ postal_code      │
│ country          │
│ directions       │
│ capacity         │
│ timezone         │ IANA timezone
│ is_virtual       │ boolean
│ created_at       │
└──────────────────┘

┌──────────────────┐
│      users       │
├──────────────────┤
│ id (PK)          │
│ email (unique)   │
│ password_hash    │
│ name             │
│ avatar_url       │
│ bio              │
│ phone (nullable) │
│ timezone         │ IANA timezone
│ discord_id (nullable)│
│ email_verified   │
│ is_active        │
│ created_at       │
│ updated_at       │
│ last_login_at    │
└────────┬─────────┘
         │
         │ Multiple relationships:
         │
         ├─────1:N─────┐
         │              ▼
         │     ┌──────────────────┐
         │     │ org_subscriptions│
         │     ├──────────────────┤
         │     │ id (PK)          │
         │     │ user_id (FK)     │
         │     │ organization_id (FK)│
         │     │ notify_email     │
         │     │ notify_sms       │
         │     │ notify_discord   │
         │     │ created_at       │
         │     └──────────────────┘
         │
         ├─────1:N─────┐
         │              ▼
         │     ┌──────────────────┐
         │     │group_subscriptions│
         │     ├──────────────────┤
         │     │ id (PK)          │
         │     │ user_id (FK)     │
         │     │ group_id (FK)    │
         │     │ notify_email     │
         │     │ notify_sms       │
         │     │ notify_discord   │
         │     │ created_at       │
         │     └──────────────────┘
         │
         ├─────1:N─────┐
         │              ▼
         │     ┌──────────────────┐
         │     │   event_rsvps    │
         │     ├──────────────────┤
         │     │ id (PK)          │
         │     │ user_id (FK)     │
         │     │ event_id (FK)    │
         │     │ status (enum)    │ attending, not_attending
         │     │ created_at       │
         │     │ updated_at       │
         │     └──────────────────┘
         │     (unique constraint: user_id, event_id)
         │
         ├─────1:N─────┐
         │              ▼
         │     ┌──────────────────┐
         │     │   org_admins     │
         │     ├──────────────────┤
         │     │ id (PK)          │
         │     │ user_id (FK)     │
         │     │ organization_id (FK)│
         │     │ created_at       │
         │     └──────────────────┘
         │     (unique constraint: user_id, organization_id)
         │
         ├─────1:N─────┐
         │              ▼
         │     ┌──────────────────┐
         │     │   group_admins   │
         │     ├──────────────────┤
         │     │ id (PK)          │
         │     │ user_id (FK)     │
         │     │ group_id (FK)    │
         │     │ created_at       │
         │     └──────────────────┘
         │     (unique constraint: user_id, group_id)
         │
         └─────1:N─────┐
                       ▼
              ┌──────────────────┐
              │   event_hosts    │
              ├──────────────────┤
              │ id (PK)          │
              │ user_id (FK)     │
              │ event_id (FK)    │
              │ created_at       │
              └──────────────────┘
              (unique constraint: user_id, event_id)

┌──────────────────┐
│  event_speakers  │
├──────────────────┤
│ id (PK)          │
│ event_id (FK)    │
│ name             │
│ bio              │
│ avatar_url       │
│ order            │ display order
│ created_at       │
└──────────────────┘

┌──────────────────┐
│ event_schedules  │
├──────────────────┤
│ id (PK)          │
│ event_id (FK)    │
│ time (relative)  │ minutes from event start
│ title            │
│ description      │
│ order            │
│ created_at       │
└──────────────────┘

┌──────────────────┐
│   oauth_accounts │
├──────────────────┤
│ id (PK)          │
│ user_id (FK)     │
│ provider (enum)  │ google, github
│ provider_user_id │
│ access_token (encrypted)│
│ refresh_token (encrypted)│
│ expires_at       │
│ created_at       │
│ updated_at       │
└──────────────────┘
(unique constraint: provider, provider_user_id)

┌──────────────────┐
│ email_verification_tokens │
├──────────────────┤
│ id (PK)          │
│ user_id (FK)     │
│ token (unique)   │
│ expires_at       │
│ created_at       │
└──────────────────┘

┌──────────────────┐
│password_reset_tokens│
├──────────────────┤
│ id (PK)          │
│ user_id (FK)     │
│ token (unique)   │
│ expires_at       │
│ created_at       │
└──────────────────┘
```

### Schema Details

#### organizations

```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    logo_url VARCHAR(512),
    website_url VARCHAR(512),
    community_chat_url VARCHAR(512),
    community_chat_name VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
```

#### groups

```sql
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    logo_url VARCHAR(512),
    primary_location TEXT,
    community_chat_url VARCHAR(512), -- Nullable, overrides org if set
    community_chat_name VARCHAR(100), -- Nullable
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_groups_organization_id ON groups(organization_id);
CREATE INDEX idx_groups_slug ON groups(slug);
```

#### users

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- Nullable for OAuth-only users
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(512),
    bio TEXT,
    phone VARCHAR(20),
    timezone VARCHAR(100) NOT NULL DEFAULT 'America/Denver', -- IANA timezone
    discord_id VARCHAR(100),
    email_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE, -- Soft delete flag
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);
```

#### venues

```sql
CREATE TABLE venues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    directions TEXT,
    capacity INTEGER,
    timezone VARCHAR(100) NOT NULL, -- IANA timezone
    is_virtual BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_venues_city_state ON venues(city, state);
```

#### events

```sql
CREATE TYPE event_status AS ENUM ('draft', 'published', 'cancelled');

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    status event_status DEFAULT 'draft',
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    timezone VARCHAR(100) NOT NULL, -- IANA timezone for display
    venue_id UUID REFERENCES venues(id) ON DELETE SET NULL,
    capacity INTEGER, -- Nullable for unlimited capacity
    youtube_url VARCHAR(512),
    
    -- Recurring event fields
    parent_event_id UUID REFERENCES events(id) ON DELETE CASCADE, -- NULL for non-recurring or series parent
    recurrence_rule TEXT, -- RRULE format (RFC 5545)
    recurrence_end_date TIMESTAMP WITH TIME ZONE, -- When recurrence stops
    recurrence_count INTEGER, -- Alternative to end_date: stop after N occurrences
    is_recurring_parent BOOLEAN DEFAULT FALSE, -- True for the template event
    
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_time_range CHECK (end_time > start_time)
);

CREATE INDEX idx_events_group_id ON events(group_id);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_slug ON events(slug);
CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_parent_event_id ON events(parent_event_id); -- For recurring series

-- Composite index for common query: upcoming published events for a group
CREATE INDEX idx_events_group_status_time ON events(group_id, status, start_time);
```

#### org_admins

```sql
CREATE TABLE org_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, organization_id)
);

CREATE INDEX idx_org_admins_user_id ON org_admins(user_id);
CREATE INDEX idx_org_admins_organization_id ON org_admins(organization_id);
```

#### group_admins

```sql
CREATE TABLE group_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, group_id)
);

CREATE INDEX idx_group_admins_user_id ON group_admins(user_id);
CREATE INDEX idx_group_admins_group_id ON group_admins(group_id);
```

#### event_hosts

```sql
CREATE TABLE event_hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, event_id)
);

CREATE INDEX idx_event_hosts_user_id ON event_hosts(user_id);
CREATE INDEX idx_event_hosts_event_id ON event_hosts(event_id);
```

#### org_subscriptions

```sql
CREATE TABLE org_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    notify_email BOOLEAN DEFAULT TRUE,
    notify_sms BOOLEAN DEFAULT FALSE,
    notify_discord BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, organization_id)
);

CREATE INDEX idx_org_subscriptions_user_id ON org_subscriptions(user_id);
CREATE INDEX idx_org_subscriptions_organization_id ON org_subscriptions(organization_id);
```

#### group_subscriptions

```sql
CREATE TABLE group_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    notify_email BOOLEAN DEFAULT TRUE,
    notify_sms BOOLEAN DEFAULT FALSE,
    notify_discord BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, group_id)
);

CREATE INDEX idx_group_subscriptions_user_id ON group_subscriptions(user_id);
CREATE INDEX idx_group_subscriptions_group_id ON group_subscriptions(group_id);
```

#### event_rsvps

```sql
CREATE TYPE rsvp_status AS ENUM ('attending', 'not_attending');

CREATE TABLE event_rsvps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    status rsvp_status NOT NULL DEFAULT 'attending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, event_id)
);

CREATE INDEX idx_event_rsvps_user_id ON event_rsvps(user_id);
CREATE INDEX idx_event_rsvps_event_id ON event_rsvps(event_id);
CREATE INDEX idx_event_rsvps_status ON event_rsvps(status);
```

#### event_speakers

```sql
CREATE TABLE event_speakers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    bio TEXT,
    avatar_url VARCHAR(512),
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_event_speakers_event_id ON event_speakers(event_id);
```

#### event_schedules

```sql
CREATE TABLE event_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    time_offset_minutes INTEGER NOT NULL, -- Minutes from event start_time
    title VARCHAR(255) NOT NULL,
    description TEXT,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_event_schedules_event_id ON event_schedules(event_id);
```

#### oauth_accounts

```sql
CREATE TYPE oauth_provider AS ENUM ('google', 'github');

CREATE TABLE oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider oauth_provider NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT, -- Should be encrypted at application level
    refresh_token TEXT, -- Should be encrypted at application level
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);
```

#### email_verification_tokens

```sql
CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_email_verification_tokens_token ON email_verification_tokens(token);
CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
```

#### password_reset_tokens

```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
```

### Key Database Design Decisions

#### UUID Primary Keys
- **Rationale**: UUIDs prevent enumeration attacks, allow distributed ID generation, and avoid sequential ID leakage

#### Timestamps with Time Zone
- **Rationale**: Proper timezone support required for multi-timezone events. All times stored in UTC.

#### Soft Deletes via is_active
- **Rationale**: Preserve referential integrity, allow data recovery, maintain audit trail

#### Composite Indexes
- **Rationale**: Optimize common query patterns (e.g., "get upcoming published events for a group")

#### Cascading Deletes
- **Rationale**: 
  - Organization deletion cascades to groups and all children (rare operation, should be protected at app level)
  - User deletion cascades to all their associations (GDPR compliance)
  - Event deletion cascades to RSVPs, hosts, speakers (clean data)

#### Unique Constraints
- **Rationale**: Prevent duplicate subscriptions, multiple RSVPs, duplicate admin assignments

#### Recurring Events Implementation
- **Approach**: Create actual event instances rather than virtual occurrences
- **Rationale**: 
  - Simplifies queries (no complex RRULE expansion at runtime)
  - Allows individual event modifications (change time, cancel single occurrence)
  - Better for RSVP tracking and notifications
  - More straightforward for API consumers
- **Trade-off**: More database rows, but better query performance and flexibility
- **Process**: Background job generates event instances when parent is created, typically 6-12 months ahead
- **RRULE Storage**: Store original recurrence rule for reference and future generation

---

## API Design

### API Principles

1. **RESTful Design**: Resources as nouns, HTTP verbs for actions
2. **Consistent Response Format**: Uniform success/error structure
3. **Versioning**: API versioned via URL path (`/api/v1/...`)
4. **Pagination**: Cursor-based pagination for lists
5. **Filtering & Sorting**: Query parameters for filtering, sorting
6. **Rate Limiting**: Protect against abuse
7. **CORS**: Allow specified origins for web client

### Base URL Structure

```
https://meetup.forgeutah.org/api/v1/
```

### Authentication

All authenticated requests include JWT in Authorization header:
```
Authorization: Bearer <jwt_token>
```

### Response Format

#### Success Response (200, 201)
```json
{
  "data": {
    // Resource or array of resources
  },
  "meta": {
    // Optional metadata (pagination, counts, etc.)
  }
}
```

#### Error Response (4xx, 5xx)
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "The requested event could not be found",
    "details": {
      "event_id": "123e4567-e89b-12d3-a456-426614174000"
    }
  }
}
```

### API Endpoints

#### Authentication & User Management

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
POST   /api/v1/auth/refresh
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
POST   /api/v1/auth/verify-email
GET    /api/v1/auth/oauth/:provider
GET    /api/v1/auth/oauth/:provider/callback

GET    /api/v1/users/me
PATCH  /api/v1/users/me
DELETE /api/v1/users/me
GET    /api/v1/users/:userId (public profile)
```

#### Organizations

```
GET    /api/v1/organizations/:orgId
PATCH  /api/v1/organizations/:orgId (org admin only)
GET    /api/v1/organizations/:orgId/groups
GET    /api/v1/organizations/:orgId/events
GET    /api/v1/organizations/:orgId/subscribers (org admin only)
POST   /api/v1/organizations/:orgId/subscribe
DELETE /api/v1/organizations/:orgId/subscribe
```

#### Groups

```
GET    /api/v1/groups
GET    /api/v1/groups/:groupId
POST   /api/v1/groups (org admin only)
PATCH  /api/v1/groups/:groupId (group admin or org admin)
DELETE /api/v1/groups/:groupId (org admin only)

GET    /api/v1/groups/:groupId/events
GET    /api/v1/groups/:groupId/subscribers (group admin or org admin)
POST   /api/v1/groups/:groupId/subscribe
DELETE /api/v1/groups/:groupId/subscribe

GET    /api/v1/groups/:groupId/admins (group admin or org admin)
POST   /api/v1/groups/:groupId/admins (group admin or org admin)
DELETE /api/v1/groups/:groupId/admins/:userId (group admin or org admin)
```

#### Events

```
GET    /api/v1/events
GET    /api/v1/events/:eventId
POST   /api/v1/events (group admin only)
PATCH  /api/v1/events/:eventId (event host or group admin or org admin)
DELETE /api/v1/events/:eventId (group admin or org admin)

GET    /api/v1/events/:eventId/rsvps (host/admin: full list, public: count)
POST   /api/v1/events/:eventId/rsvp
DELETE /api/v1/events/:eventId/rsvp

GET    /api/v1/events/:eventId/hosts (host or group admin or org admin)
POST   /api/v1/events/:eventId/hosts (event host or group admin or org admin)
DELETE /api/v1/events/:eventId/hosts/:userId (event host or group admin or org admin)

GET    /api/v1/events/:eventId/speakers
POST   /api/v1/events/:eventId/speakers (event host or group admin or org admin)
PATCH  /api/v1/events/:eventId/speakers/:speakerId (event host or group admin or org admin)
DELETE /api/v1/events/:eventId/speakers/:speakerId (event host or group admin or org admin)

GET    /api/v1/events/:eventId/schedule
POST   /api/v1/events/:eventId/schedule (event host or group admin or org admin)
PATCH  /api/v1/events/:eventId/schedule/:scheduleId (event host or group admin or org admin)
DELETE /api/v1/events/:eventId/schedule/:scheduleId (event host or group admin or org admin)

GET    /api/v1/events/:eventId/ics (calendar export)
```

#### Venues

```
GET    /api/v1/venues (with search/filter params)
GET    /api/v1/venues/:venueId
POST   /api/v1/venues (group admin or org admin)
PATCH  /api/v1/venues/:venueId (group admin or org admin)
```

#### User Dashboard

```
GET    /api/v1/me/subscriptions
GET    /api/v1/me/rsvps (upcoming events user RSVP'd to)
GET    /api/v1/me/events (events user is hosting)
GET    /api/v1/me/groups (groups user admins)
GET    /api/v1/me/organization (organization user admins)
```

### Example API Request/Response

#### GET /api/v1/events/:eventId

**Request:**
```http
GET /api/v1/events/123e4567-e89b-12d3-a456-426614174000 HTTP/1.1
Host: meetup.forgeutah.org
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**
```json
{
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Kubernetes 1.30: What's New",
    "slug": "kubernetes-1-30-whats-new",
    "description": "Join us for an in-depth look at the latest features...",
    "status": "published",
    "start_time": "2025-12-15T19:00:00Z",
    "end_time": "2025-12-15T21:00:00Z",
    "timezone": "America/Denver",
    "display_time": "7:00 PM MST", // Computed based on user's timezone
    "capacity": 50,
    "rsvp_count": 32,
    "youtube_url": "https://youtube.com/live/abc123",
    "group": {
      "id": "789e4567-e89b-12d3-a456-426614174000",
      "name": "Kubernetes Meetup",
      "slug": "kubernetes-meetup"
    },
    "venue": {
      "id": "456e4567-e89b-12d3-a456-426614174000",
      "name": "DevHub Coworking",
      "address": "123 Main St",
      "city": "Salt Lake City",
      "state": "UT",
      "timezone": "America/Denver"
    },
    "speakers": [
      {
        "id": "speaker-001",
        "name": "Jane Doe",
        "bio": "Platform Engineer at TechCo",
        "avatar_url": "https://..."
      }
    ],
    "schedule": [
      {
        "id": "schedule-001",
        "time_offset_minutes": 0,
        "title": "Networking & Pizza",
        "description": null
      },
      {
        "id": "schedule-002",
        "time_offset_minutes": 30,
        "title": "Main Presentation",
        "description": "Deep dive into K8s 1.30 features"
      }
    ],
    "community_chat": {
      "name": "Discord",
      "url": "https://discord.gg/forgeutah"
    },
    "user_rsvp": "attending", // Only if user is authenticated
    "is_host": false, // Only if user is authenticated
    "created_at": "2025-11-01T10:00:00Z",
    "updated_at": "2025-11-10T15:30:00Z"
  }
}
```

#### POST /api/v1/events/:eventId/rsvp

**Request:**
```http
POST /api/v1/events/123e4567-e89b-12d3-a456-426614174000/rsvp HTTP/1.1
Host: meetup.forgeutah.org
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "status": "attending"
}
```

**Response:**
```json
{
  "data": {
    "event_id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "user-789",
    "status": "attending",
    "created_at": "2025-11-12T16:45:00Z"
  },
  "meta": {
    "message": "RSVP confirmed! You'll receive a calendar invite via email."
  }
}
```

---

## Authentication & Authorization

### Authentication Strategy

#### JWT-Based Authentication

**Token Structure:**
```json
{
  "sub": "user-id-123",
  "email": "user@example.com",
  "name": "John Doe",
  "exp": 1700000000,
  "iat": 1699990000,
  "permissions": {
    "org_admin": ["org-id-1"],
    "group_admin": ["group-id-1", "group-id-2"],
    "event_host": ["event-id-1", "event-id-2"]
  }
}
```

**Token Lifecycle:**
- **Access Token**: Short-lived (15 minutes), contains user identity and permissions
- **Refresh Token**: Long-lived (7 days), used to obtain new access tokens
- Refresh tokens stored in httpOnly cookie or in Redis with user session

**Security Measures:**
- Tokens signed with HS256 or RS256
- Refresh token rotation on use
- Token blacklist in Redis for logout
- Rate limiting on token refresh endpoint

#### OAuth 2.0 Flow

**Google OAuth:**
1. User clicks "Sign in with Google"
2. Redirect to Google authorization endpoint
3. User grants permissions
4. Google redirects to callback with authorization code
5. Exchange code for access token and user info
6. Create or link user account
7. Issue application JWT tokens

**GitHub OAuth:** Similar flow

### Authorization & Permission Model

#### Permission Hierarchy

```
Organization Admin (Super User)
  ↓ Inherits all permissions below
  ├── Can manage organization settings
  ├── Can create/delete groups
  ├── Can assign group admins
  └── Can view all data in organization
  
Group Admin
  ↓ Has full control of their group
  ├── Can manage group settings
  ├── Can create/edit/delete events in their group
  ├── Can assign group admins (co-admins)
  ├── Can assign event hosts
  └── Can view group subscribers and event RSVPs
  
Event Host
  ↓ Scoped permissions to specific event
  ├── Can edit event details
  ├── Can add/remove co-hosts
  ├── Can view event RSVPs
  └── CANNOT delete event or manage group
  
Subscriber
  ↓ Basic user permissions
  ├── Can subscribe to groups/org
  ├── Can RSVP to events
  ├── Can manage their own profile
  └── Can view public content
```

#### Permission Checks

**Event Host Subscription Enforcement:**

When assigning a user as event host:
```
IF user is NOT subscribed to parent organization AND
   user is NOT subscribed to event's group
THEN
   REJECT with error: "Host must be subscribed to organization or group"
END IF
```

**Implementation in code (pseudo-code):**
```go
func CanAddEventHost(userID, eventID, requestorID string) error {
    // 1. Check requestor has permission to add hosts
    if !IsEventHost(requestorID, eventID) && 
       !IsGroupAdmin(requestorID, event.GroupID) &&
       !IsOrgAdmin(requestorID, event.Group.OrganizationID) {
        return ErrUnauthorized
    }
    
    // 2. Check user to be added is subscribed
    isOrgSubscriber := IsOrgSubscriber(userID, event.Group.OrganizationID)
    isGroupSubscriber := IsGroupSubscriber(userID, event.GroupID)
    
    if !isOrgSubscriber && !isGroupSubscriber {
        return ErrHostMustBeSubscribed
    }
    
    return nil
}
```

### Rate Limiting

**Strategy: Token Bucket Algorithm**

Limits per IP and per authenticated user:
- **Unauthenticated requests**: 100 req/hour per IP
- **Authenticated requests**: 1000 req/hour per user
- **Authentication endpoints**: 5 req/minute (login, register)
- **Password reset**: 3 req/hour

Implementation using Redis:
```go
func RateLimit(key string, limit int, window time.Duration) bool {
    count := redis.Incr(key)
    if count == 1 {
        redis.Expire(key, window)
    }
    return count <= limit
}
```

---

## Notification System

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│              APPLICATION SERVER                         │
│                                                         │
│  Event occurs (new event, RSVP, etc.)                  │
│           │                                             │
│           ▼                                             │
│  ┌─────────────────────────────┐                       │
│  │  Notification Dispatcher     │                       │
│  │  - Determines recipients     │                       │
│  │  - Checks notification prefs │                       │
│  │  - Enqueues jobs             │                       │
│  └─────────────┬────────────────┘                       │
└────────────────┼────────────────────────────────────────┘
                 │
                 ▼
        ┌────────────────┐
        │  MESSAGE QUEUE  │
        │  (Redis/PgBoss) │
        └────────┬───────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│         BACKGROUND JOB PROCESSOR                        │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │ Email Worker │  │  SMS Worker  │  │Discord Worker│ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
└─────────┼──────────────────┼──────────────────┼─────────┘
          │                  │                  │
          ▼                  ▼                  ▼
┌────────────────┐  ┌────────────────┐  ┌────────────────┐
│ Email Service  │  │  SMS Service   │  │ Discord API    │
│ (SendGrid/SES) │  │    (Twilio)    │  │                │
└────────────────┘  └────────────────┘  └────────────────┘
```

### Notification Triggers

| Trigger | Recipients | Channels | Priority |
|---------|-----------|----------|----------|
| New Event Published | Org subscribers + Group subscribers | Email, SMS, Discord | Normal |
| Event Updated (time/venue) | RSVPs | Email, SMS | High |
| Event Cancelled | RSVPs | Email, SMS | High |
| Event Reminder (24hr) | RSVPs with reminder enabled | Email, SMS, Discord | Normal |
| Event Reminder (1hr) | RSVPs with reminder enabled | SMS, Discord | High |
| Event Capacity 80% | Event hosts + Group admins | Email | Low |
| Event Capacity 100% | Event hosts + Group admins | Email | Normal |
| RSVP Confirmation | User who RSVP'd | Email (with ICS) | High |

### Notification Logic

#### Determining Recipients

**For new event published:**
```sql
-- Get all org subscribers with email notifications enabled
SELECT u.id, u.email, u.timezone
FROM users u
JOIN org_subscriptions os ON u.id = os.user_id
WHERE os.organization_id = :org_id
  AND os.notify_email = true
  AND u.email_verified = true
  AND u.is_active = true

UNION

-- Get all group subscribers with email notifications enabled
-- (excluding those already in org subscribers to avoid duplicates)
SELECT u.id, u.email, u.timezone
FROM users u
JOIN group_subscriptions gs ON u.id = gs.user_id
WHERE gs.group_id = :group_id
  AND gs.notify_email = true
  AND u.email_verified = true
  AND u.is_active = true
  AND u.id NOT IN (
    SELECT user_id FROM org_subscriptions 
    WHERE organization_id = :org_id
  )
```

#### Notification Deduplication

Store notification log to prevent duplicate sends within time window:
```sql
CREATE TABLE notification_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    notification_type VARCHAR(50) NOT NULL,
    resource_id UUID NOT NULL, -- Event ID, Group ID, etc.
    channel VARCHAR(20) NOT NULL, -- email, sms, discord
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'pending' -- pending, sent, failed
);

CREATE INDEX idx_notification_log_user_type_resource ON 
    notification_log(user_id, notification_type, resource_id);
```

### Email Templates

**Template System: Handlebars or similar**

Each template includes:
- HTML version (styled)
- Plain text version (fallback)
- Transactional metadata (unsubscribe links, sender info)

**Example: New Event Notification**

```handlebars
Subject: [{{group.name}}] New Event: {{event.title}}

Hi {{user.name}},

A new event has been scheduled for {{group.name}}:

**{{event.title}}**
📅 {{event.display_time}} {{event.timezone}}
📍 {{venue.name}}, {{venue.city}}

{{event.description}}

[View Event Details & RSVP]({{event.url}})

{{#if event.youtube_url}}
🎥 This event will be streamed live on YouTube
{{/if}}

{{#if community_chat}}
Have questions? Join us on {{community_chat.name}}: {{community_chat.url}}
{{/if}}

---
You're receiving this because you're subscribed to {{subscription_type}}.
[Manage Subscription Preferences]({{preferences_url}})
```

### SMS Templates

**Character limit: 160 characters (single SMS)**

```
[{{group.name}}] New event: {{event.title}}
{{event.short_date}} at {{venue.name}}
RSVP: {{short_url}}
```

### Calendar Integration (ICS)

**ICS Generation for RSVP Confirmation:**

```
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Forge Utah//Meetup Platform//EN
CALSCALE:GREGORIAN
METHOD:REQUEST
BEGIN:VEVENT
UID:{{event.id}}@meetup.forgeutah.org
DTSTAMP:{{now_utc}}
DTSTART;TZID={{event.timezone}}:{{event.start_time}}
DTEND;TZID={{event.timezone}}:{{event.end_time}}
SUMMARY:{{event.title}}
DESCRIPTION:{{event.description}}\n\nView event: {{event.url}}
LOCATION:{{venue.name}}, {{venue.address}}
URL:{{event.url}}
ORGANIZER;CN="{{group.name}}":mailto:noreply@meetup.forgeutah.org
STATUS:CONFIRMED
SEQUENCE:0
BEGIN:VALARM
TRIGGER:-PT24H
ACTION:DISPLAY
DESCRIPTION:Event reminder: {{event.title}}
END:VALARM
END:VEVENT
END:VCALENDAR
```

---

## Frontend Architecture

### Recommended Stack: Server-Side Rendering + Progressive Enhancement

**Technology Choice: Go + Templ + HTMX + Alpine.js**

**Rationale:**
- Fast server-side rendering for initial page load
- SEO-friendly
- Progressive enhancement: works without JS
- HTMX for dynamic updates without full SPA complexity
- Alpine.js for lightweight client-side interactions
- Matches your Go backend expertise

**Alternative: React SPA (if interactive features are priority)**

### Page Structure

#### Public Pages (No Auth Required)
- `/` - Homepage (upcoming events)
- `/groups` - Group directory
- `/groups/:slug` - Group detail page
- `/events/:slug` - Event detail page
- `/login` - Login page
- `/register` - Registration page
- `/forgot-password` - Password reset

#### Authenticated User Pages
- `/dashboard` - User dashboard (RSVPs, subscriptions)
- `/profile` - User profile settings
- `/events/:slug/rsvp` - RSVP to event

#### Admin Pages
- `/admin/dashboard` - Org/group admin dashboard
- `/admin/events/new` - Create event
- `/admin/events/:id/edit` - Edit event
- `/admin/groups/:id` - Manage group
- `/admin/organization` - Org settings (org admin only)

### Component Hierarchy (React example)

```
App
├── Layout
│   ├── Header
│   │   ├── Logo
│   │   ├── Navigation
│   │   └── UserMenu
│   ├── Main (content area)
│   └── Footer
│
├── Public Routes
│   ├── HomePage
│   │   ├── EventList
│   │   │   └── EventCard
│   │   └── GroupList
│   │       └── GroupCard
│   ├── GroupPage
│   │   ├── GroupHeader
│   │   ├── SubscribeButton
│   │   └── EventList
│   ├── EventPage
│   │   ├── EventHeader
│   │   ├── EventDetails
│   │   ├── VenueMap
│   │   ├── SpeakerList
│   │   ├── Schedule
│   │   ├── RSVPButton
│   │   └── YouTubeEmbed (if streaming)
│   └── Auth Pages (Login, Register)
│
└── Protected Routes
    ├── Dashboard
    │   ├── UpcomingRSVPs
    │   ├── MySubscriptions
    │   └── MyHostedEvents (if host)
    ├── ProfileSettings
    └── AdminDashboard
        ├── GroupManagement
        ├── EventManagement
        └── SubscriberManagement
```

### Design System (Luma-inspired)

**Color Palette:**
- Primary: Indigo/Purple tones
- Accent: Vibrant color for CTAs
- Neutrals: Clean grays for text and backgrounds
- Success/Warning/Error states

**Typography:**
- Headings: Sans-serif, bold
- Body: Sans-serif, medium weight
- Monospace: For code examples (if needed)

**Components:**
- Cards with subtle shadows
- Rounded corners (8px-12px)
- Smooth transitions
- High-contrast CTAs
- Spacious layout

**Accessibility:**
- WCAG 2.1 AA compliance
- Focus indicators
- Keyboard navigation
- ARIA labels
- Color contrast ratios > 4.5:1

---

## Infrastructure & Deployment

### Self-Hosting Requirements

**Minimum Server Specs:**
- CPU: 2 vCPUs
- RAM: 4GB
- Storage: 50GB SSD
- Network: 1 Gbps

**Recommended Specs (for growth):**
- CPU: 4 vCPUs
- RAM: 8GB
- Storage: 100GB SSD
- Separate Redis server

### Deployment Architecture

**Option 1: Single Server (MVP)**
```
┌─────────────────────────────────────────┐
│         VPS / Dedicated Server          │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │   Nginx (reverse proxy + TLS)     │ │
│  └────────────┬──────────────────────┘ │
│               │                         │
│  ┌────────────▼──────────────────────┐ │
│  │  Application Server (Backend)     │ │
│  └────────────┬──────────────────────┘ │
│               │                         │
│  ┌────────────▼──────────────────────┐ │
│  │  PostgreSQL Database              │ │
│  └───────────────────────────────────┘ │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │  Redis (cache + job queue)        │ │
│  └───────────────────────────────────┘ │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │  Background Job Worker            │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

**Option 2: Docker Compose (Recommended for portability)**

```yaml
version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
  
  app:
    build: ./app
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/meetup
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis
  
  postgres:
    image: postgres:15-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=meetup
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=secure_password
  
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
  
  worker:
    build: ./app
    command: worker
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/meetup
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

volumes:
  postgres_data:
  redis_data:
```

**Option 3: Kubernetes (Future scaling)**
- Horizontal scaling of app servers
- Managed PostgreSQL (RDS, Cloud SQL)
- Managed Redis
- Auto-scaling based on load

### CI/CD Pipeline

**GitHub Actions Example:**

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: make test
  
  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build Docker image
        run: docker build -t meetup-app .
      - name: Push to registry
        run: docker push ghcr.io/forgeutah/meetup-app:latest
  
  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to server
        run: |
          ssh user@server "cd /app && docker-compose pull && docker-compose up -d"
```

### Backup Strategy

**Database Backups:**
- Daily full backups
- Hourly incremental backups (pg_basebackup or WAL archiving)
- Retention: 30 days
- Offsite storage (S3, Backblaze B2)

**Backup Script Example:**
```bash
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
FILENAME="meetup_db_$DATE.sql.gz"

pg_dump -U user meetup | gzip > "$BACKUP_DIR/$FILENAME"
# Upload to S3
aws s3 cp "$BACKUP_DIR/$FILENAME" s3://backups/meetup-db/
# Cleanup old backups (keep 30 days)
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +30 -delete
```

### Monitoring & Logging

**Metrics (Prometheus + Grafana):**
- Request rate, latency, error rate
- Database connection pool usage
- Cache hit/miss ratio
- Background job queue depth
- System metrics (CPU, memory, disk)

**Logging (Structured JSON logs):**
- Application logs (errors, warnings, info)
- Access logs (Nginx)
- Database slow query logs
- Background job logs

**Log aggregation:**
- Loki (lightweight option)
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Cloud options (AWS CloudWatch, Datadog)

**Alerting:**
- Email/SMS alerts for critical errors
- PagerDuty integration for on-call
- Slack notifications for warnings

---

## Security Considerations

### Threat Model

**Assets to Protect:**
1. User credentials and personal data
2. Event and group data integrity
3. Admin access and permissions
4. API availability

**Threat Actors:**
- Malicious users (spam, abuse)
- Automated bots (scraping, DDoS)
- Compromised accounts
- Insider threats (rogue admins)

### Security Measures

#### Application Security

**Input Validation:**
- Validate all user inputs server-side
- Sanitize HTML to prevent XSS
- Use prepared statements to prevent SQL injection
- Validate file uploads (images)

**Output Encoding:**
- Escape all user-generated content in HTML
- Use Content Security Policy headers

**Authentication Security:**
- Bcrypt/Argon2 for password hashing (cost factor 12+)
- Email verification required
- Password strength requirements
- Rate limiting on auth endpoints
- Multi-factor authentication (future enhancement)

**Session Security:**
- httpOnly cookies for refresh tokens
- Secure flag on cookies (HTTPS only)
- SameSite=Strict for CSRF protection
- Session expiration and rotation

**Authorization:**
- Check permissions on every request
- Never trust client-side permission checks
- Audit trail for admin actions

#### Infrastructure Security

**Network Security:**
- Firewall rules (only ports 80, 443, 22 exposed)
- SSH key-based authentication only
- Fail2ban for SSH brute force protection
- VPN for admin access (optional)

**TLS/SSL:**
- Let's Encrypt for free SSL certificates
- TLS 1.3 minimum
- HSTS headers
- Certificate pinning (future)

**Database Security:**
- Strong passwords
- Limited user privileges (principle of least privilege)
- No direct external access (behind firewall)
- Encrypted backups
- Encryption at rest (LUKS, dm-crypt)

**Secrets Management:**
- Environment variables for secrets (never in code)
- Secret management tools (Vault, AWS Secrets Manager)
- Rotate secrets regularly

#### Compliance

**GDPR Compliance:**
- User data export functionality
- User account deletion (right to be forgotten)
- Clear privacy policy
- Consent for email/SMS notifications
- Data processing agreement (if in EU)

**CCPA Compliance (if applicable):**
- Similar to GDPR requirements

---

## Performance & Scalability

### Performance Targets

- **Page Load Time**: < 2 seconds (LCP)
- **API Response Time**: < 200ms (p95)
- **Database Query Time**: < 50ms (p95)
- **Time to First Byte**: < 600ms

### Optimization Strategies

#### Database Optimization

**Indexing:**
- Composite indexes for common query patterns
- Partial indexes for filtered queries
- Full-text search indexes for event/group search

**Query Optimization:**
- Use EXPLAIN ANALYZE to identify slow queries
- Avoid N+1 queries (use JOINs or eager loading)
- Pagination with cursor-based approach for large datasets

**Connection Pooling:**
- Limit max connections (25-50 for small deployments)
- Use PgBouncer for connection pooling

#### Caching Strategy

**Cache Layers:**
1. **HTTP Cache**: Browser cache for static assets (CSS, JS, images)
2. **CDN Cache**: Edge caching for static assets (future)
3. **Application Cache**: Redis for frequently accessed data
4. **Database Cache**: Query result cache

**Cache Keys:**
```
events:upcoming:{group_id}:{page}
groups:directory
user:profile:{user_id}
event:detail:{event_id}
```

**Cache Invalidation:**
- Time-based expiration (TTL)
- Event-based invalidation (on update/delete)
- Cache tags for bulk invalidation

**Example Caching Logic:**
```go
func GetUpcomingEvents(groupID string, page int) ([]Event, error) {
    cacheKey := fmt.Sprintf("events:upcoming:%s:%d", groupID, page)
    
    // Try cache first
    cached, err := redis.Get(cacheKey)
    if err == nil {
        return json.Unmarshal(cached, &events)
    }
    
    // Cache miss - query database
    events, err := db.Query("SELECT ... FROM events WHERE ...")
    if err != nil {
        return nil, err
    }
    
    // Store in cache (5 minutes TTL)
    redis.Set(cacheKey, json.Marshal(events), 5*time.Minute)
    
    return events, nil
}
```

#### Frontend Optimization

**Asset Optimization:**
- Minify CSS/JS
- Image optimization (WebP format, lazy loading)
- Code splitting for JavaScript
- Tree shaking to remove unused code

**Critical Rendering Path:**
- Inline critical CSS
- Defer non-critical JavaScript
- Preload key resources
- Use HTTP/2 for multiplexing

#### Background Job Optimization

**Job Prioritization:**
- High priority: Event cancellation notifications
- Normal priority: New event notifications, reminders
- Low priority: Analytics, reporting

**Batch Processing:**
- Batch email sends (up to 100 per API call to SendGrid)
- Batch database inserts

### Scalability Roadmap

**Phase 1: Single Server (0-10,000 users)**
- Vertical scaling (more CPU/RAM)
- Database tuning
- Redis caching

**Phase 2: Horizontal Scaling (10,000-100,000 users)**
- Multiple app servers behind load balancer
- Managed PostgreSQL (read replicas)
- CDN for static assets
- Separate background job workers

**Phase 3: Microservices (100,000+ users)**
- Split notification service
- Separate search service (Elasticsearch)
- Event sourcing for audit trail
- GraphQL API (optional)

---

## Third-Party Integrations

### Email Service Integration

**SendGrid API Example:**
```go
import "github.com/sendgrid/sendgrid-go"

func SendEmail(to, subject, body string, icsAttachment []byte) error {
    message := mail.NewV3Mail()
    message.SetFrom(mail.NewEmail("Forge Utah", "noreply@forgeutah.org"))
    message.AddContent(mail.NewContent("text/html", body))
    
    personalization := mail.NewPersonalization()
    personalization.AddTos(mail.NewEmail("", to))
    personalization.Subject = subject
    message.AddPersonalizations(personalization)
    
    // Attach ICS file
    attachment := mail.NewAttachment()
    attachment.SetContent(base64.StdEncoding.EncodeToString(icsAttachment))
    attachment.SetType("text/calendar")
    attachment.SetFilename("event.ics")
    attachment.SetDisposition("attachment")
    message.AddAttachment(attachment)
    
    client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
    response, err := client.Send(message)
    return err
}
```

### SMS Integration (Twilio)

```go
import "github.com/twilio/twilio-go"

func SendSMS(to, body string) error {
    client := twilio.NewRestClientWithParams(twilio.ClientParams{
        Username: os.Getenv("TWILIO_ACCOUNT_SID"),
        Password: os.Getenv("TWILIO_AUTH_TOKEN"),
    })
    
    params := &twilioApi.CreateMessageParams{}
    params.SetTo(to)
    params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
    params.SetBody(body)
    
    _, err := client.Api.CreateMessage(params)
    return err
}
```

### YouTube Embed

**Simple iframe embed:**
```html
<div class="video-container">
  <iframe 
    width="560" 
    height="315" 
    src="https://www.youtube.com/embed/{{video_id}}" 
    title="Live Stream"
    frameborder="0" 
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
    allowfullscreen>
  </iframe>
</div>
```

**Extract video ID from URL:**
```go
func ExtractYouTubeVideoID(url string) (string, error) {
    patterns := []string{
        `(?:youtube\.com\/watch\?v=|youtu\.be\/)([a-zA-Z0-9_-]{11})`,
        `youtube\.com\/embed\/([a-zA-Z0-9_-]{11})`,
        `youtube\.com\/live\/([a-zA-Z0-9_-]{11})`,
    }
    
    for _, pattern := range patterns {
        re := regexp.MustCompile(pattern)
        matches := re.FindStringSubmatch(url)
        if len(matches) > 1 {
            return matches[1], nil
        }
    }
    
    return "", errors.New("invalid YouTube URL")
}
```

### Discord Integration (Future)

**Discord Bot for Event Notifications:**
- Create Discord bot application
- OAuth2 for user linking
- Webhook for posting events to channels
- DM notifications for subscribed users

### OAuth Integration

**Google OAuth Example:**
```go
import "golang.org/x/oauth2/google"

func HandleGoogleCallback(code string) (*User, error) {
    token, err := googleOauthConfig.Exchange(context.Background(), code)
    if err != nil {
        return nil, err
    }
    
    client := googleOauthConfig.Client(context.Background(), token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var googleUser struct {
        ID      string `json:"id"`
        Email   string `json:"email"`
        Name    string `json:"name"`
        Picture string `json:"picture"`
    }
    json.NewDecoder(resp.Body).Decode(&googleUser)
    
    // Find or create user
    user, err := FindOrCreateOAuthUser("google", googleUser.ID, googleUser.Email, googleUser.Name)
    return user, err
}
```

---

## Next Steps & Implementation Phases

### Phase 1: MVP (Weeks 1-8)

**Week 1-2: Project Setup**
- Initialize repository
- Set up development environment
- Database schema implementation
- Docker Compose for local dev

**Week 3-4: Backend Core**
- User authentication (email/password)
- Organization & group management
- Permission system

**Week 5-6: Event Management**
- Event CRUD operations
- RSVP functionality
- Venue management

**Week 7-8: Frontend & Integration**
- Public event pages
- User dashboard
- RSVP flow
- Email notifications (basic)

### Phase 2: Polish & Launch (Weeks 9-12)

**Week 9: Subscription System**
- Org and group subscriptions
- Notification preferences

**Week 10: Calendar & Notifications**
- ICS generation
- Email notification templates
- Scheduled reminders

**Week 11: Admin Features**
- Admin dashboards
- Subscriber management
- Event analytics (basic)

**Week 12: Testing & Deployment**
- End-to-end testing
- Security audit
- Production deployment
- Migration from Meetup.com

### Phase 3: Enhancements (Post-Launch)

**Months 2-3:**
- SMS notifications
- OAuth (Google, GitHub)
- Discord integration
- Mobile-responsive improvements

**Months 4-6:**
- Recurring events
- Waitlist functionality
- Advanced analytics
- API documentation

---

## Appendices

### A. Glossary

- **Organization**: Parent entity (e.g., Forge Utah Foundation)
- **Group**: Sub-entity representing a user group (e.g., Kubernetes Meetup)
- **Event**: Scheduled meetup with time, venue, and details
- **RSVP**: User confirmation of attendance
- **Host**: User with scoped permissions to manage a specific event
- **Subscription**: User opt-in to receive notifications from org or group

### B. References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [REST API Design Best Practices](https://restfulapi.net/)

### C. Environment Variables Reference

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/meetup
DATABASE_POOL_SIZE=25

# Redis
REDIS_URL=redis://localhost:6379
REDIS_POOL_SIZE=10

# Application
APP_ENV=production
APP_SECRET_KEY=your-secret-key-here
APP_PORT=8080
BASE_URL=https://meetup.forgeutah.org

# Email
SENDGRID_API_KEY=your-sendgrid-api-key
EMAIL_FROM=noreply@forgeutah.org

# SMS
TWILIO_ACCOUNT_SID=your-twilio-sid
TWILIO_AUTH_TOKEN=your-twilio-token
TWILIO_PHONE_NUMBER=+1234567890

# OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# JWT
JWT_SECRET=your-jwt-secret
JWT_EXPIRATION=900 # 15 minutes in seconds
REFRESH_TOKEN_EXPIRATION=604800 # 7 days in seconds
```

---

**Document End**

Questions or feedback? Contact: [Your Email]
