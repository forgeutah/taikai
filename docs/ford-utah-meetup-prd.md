# Product Requirements Document: Ford Utah Foundation Meetup Platform

## Executive Summary

A self-hosted event management platform designed specifically for the Ford Utah Foundation's network of technology user groups. The platform replaces Meetup.com with a Luma-inspired aesthetic while supporting parent organization structures and Discord integration.

## Problem Statement

Ford Utah Foundation manages multiple technology user groups (Kubernetes, Go, Data Engineering, ML, Engineering Leadership, Foundry Hack Night, Saint George, Architecture, Python at the Point). Current solutions (Meetup.com, Luma) lack:
- Parent organization/sub-group hierarchy management
- Proper admin permission scoping across organization levels
- Integration with existing community infrastructure (Discord, YouTube streaming)
- Tailored subscription models for multi-group organizations

## Goals

### Primary Goals
1. Enable hierarchical organization structure (parent org → user groups → events)
2. Provide flexible subscription models (org-level or group-level)
3. Integrate with existing community tools (Discord, YouTube, calendars)
4. Maintain accessibility and aesthetic excellence inspired by Luma

### Secondary Goals
1. Support future mobile app development
2. Enable easy event discovery across the Ford Utah network
3. Reduce administrative overhead for multi-group management

## Non-Goals (V1)
- Integrated chat functionality (Discord handles this)
- Video hosting (YouTube handles streaming)
- Payment processing for paid events
- Multi-organization support (single parent org to start)

---

## User Personas

### Organization Admin (e.g., Ford Utah Foundation Leadership)
**Responsibilities:**
- Manage organization-level settings and branding
- Add/remove user groups within the organization
- Manage organization-wide subscriptions
- Oversee group admins (super user permissions)

**Key Needs:**
- Dashboard view of all groups and upcoming events
- Ability to delegate group management
- Organization-level analytics and metrics

### Group Admin (e.g., Kubernetes Meetup Organizer)
**Responsibilities:**
- Create and manage events for their specific group
- Manage group-specific subscriptions
- Add/remove group admins
- Coordinate with venues and speakers

**Key Needs:**
- Event creation workflow with all necessary fields
- Subscriber management for their group
- Communication tools for event updates

### Event Host (May not be Group Admin)
**Responsibilities:**
- Manage single event details
- Add co-hosts to event
- Update event information (description, schedule, speakers)

**Key Needs:**
- Scoped permissions to specific event only
- Ability to add co-hosts
- Event-specific communication

### Subscriber/Attendee
**Responsibilities:**
- Subscribe to organizations or specific groups
- RSVP to events
- Receive notifications based on preferences

**Key Needs:**
- Easy subscription management
- Multiple notification channels (email, SMS, Discord)
- Calendar integration (ICS export, Google Calendar, etc.)
- Clear event information including streaming availability

---

## Feature Requirements

### 1. Organization Management

#### 1.1 Organization Structure
**Priority:** P0 (MVP)

**Requirements:**
- Each deployment represents one parent organization (Ford Utah Foundation)
- Organization has branding elements: name, logo, description, website URL
- Organization settings page for org admins
- Organization-level community chat link (Discord, Slack, Mattermost, IRC, etc.) and display name
- Community chat settings: URL and display name (e.g., "Discord", "Slack Community")

**User Stories:**
- As an org admin, I can configure organization branding and settings
- As an org admin, I can view all user groups under the organization

#### 1.2 Organization Subscriptions
**Priority:** P0 (MVP)

**Requirements:**
- Users can subscribe to all events across all groups in the organization
- Org subscription automatically includes new groups added later
- Org admins can view all organization subscribers
- Org admins can manually add/remove organization subscribers

**User Stories:**
- As a subscriber, I can subscribe to the entire Ford Utah organization and receive notifications for all groups
- As an org admin, I can see a list of all organization-level subscribers

### 2. User Group Management

#### 2.1 Group Creation & Settings
**Priority:** P0 (MVP)

**Requirements:**
- Groups MUST be tied to a parent organization (no orphan groups)
- Can be 1:1 (org with single group) or 1:many (org with multiple groups)
- Group settings include: name, description, logo/image, primary location, community chat link (URL + display name)
- Groups can override organization-level community chat or inherit it
- Only org admins can create or delete groups
- Group admins can edit group settings (except deletion)

**User Stories:**
- As an org admin, I can create a new user group within the organization
- As an org admin, I can delete a user group
- As a group admin, I can update my group's description and branding

#### 2.2 Group Admin Management
**Priority:** P0 (MVP)

**Requirements:**
- Each group has one or more group admins
- Group admins have full permissions within their group (event creation, subscription management)
- Org admins can add/remove group admins
- Group admins can add/remove other group admins (with at least one remaining)
- Group admins do NOT automatically become event hosts

**User Stories:**
- As an org admin, I can assign group admin permissions to users
- As a group admin, I can invite others to be co-admins of my group
- As a group admin, I have full control over events and subscriptions for my group

#### 2.3 Group Subscriptions
**Priority:** P0 (MVP)

**Requirements:**
- Users can subscribe to individual groups
- Group subscription only includes events from that specific group
- Group admins can view group subscribers
- Group admins can manually add/remove group subscribers
- Users can be subscribed to multiple groups independently

**User Stories:**
- As a subscriber, I can subscribe only to the Kubernetes meetup without getting notifications from other groups
- As a group admin, I can manage subscribers to my specific group

### 3. Event Management

#### 3.1 Event Creation
**Priority:** P0 (MVP)

**Requirements:**
- Only group admins can create events
- Event creator automatically becomes the event host
- Required fields: title, description, date/time with timezone, venue
- Optional fields: speaker(s), schedule, YouTube stream link, capacity limit, recurrence rule
- Event timezone defaults to venue location timezone or user's timezone
- Event has status: Draft, Published, Cancelled
- **Recurring Events (P0 - MVP)**: Support for monthly/weekly recurring events with flexible patterns

**User Stories:**
- As a group admin, I can create a new event with all necessary details
- As a group admin, I can save an event as a draft before publishing
- As a group admin, I can create recurring events (e.g., "Second Tuesday of every month")

#### 3.1.1 Recurring Events
**Priority:** P0 (MVP)

**Requirements:**
- Support standard recurrence patterns: daily, weekly, monthly, yearly
- Support intervals (e.g., every 2 weeks, every 3 months)
- Support day-of-week patterns (e.g., "every Monday" or "first Tuesday of month")
- Support end conditions: never end, end after N occurrences, end by specific date
- Recurring events create individual event instances (not virtual)
- Editing a recurring event offers options:
  - Edit this event only
  - Edit this and future events (updates recurrence rule)
  - Edit all events in series
- Cancelling a recurring event offers same options
- Series are linked via parent_event_id and recurrence_id
- Event creator UI provides recurrence pattern builder (similar to Google Calendar)

**User Stories:**
- As a group admin, I can create a monthly meetup that repeats on the second Tuesday forever
- As a group admin, I can update a single event in a recurring series without affecting others
- As a group admin, I can cancel all future events in a series

#### 3.2 Event Hosts & Permissions
**Priority:** P0 (MVP)

**Requirements:**
- Event creator is the initial host
- Host can add additional co-hosts manually
- Co-hosts can be group admins OR non-admins
- If host is NOT a group admin, they have scoped permissions only to that event
- Hosts can edit event details, add/remove co-hosts, communicate with attendees
- Only group admins can delete events
- **CRITICAL RULE:** Event hosts must be subscribed to either:
  - The parent organization (org-level subscription receives all events), OR
  - The specific group whose event they're hosting
- System enforces this subscription requirement when assigning hosts

**User Stories:**
- As an event host, I can add co-hosts to help manage the event
- As a non-admin event host, I can only modify the specific event I'm hosting
- As a group admin, I maintain full control even if I'm not listed as a host

#### 3.3 Event Details & Display
**Priority:** P0 (MVP)

**Requirements:**
- Event page displays: title, description, date/time, venue information, RSVP count (not list), capacity indicator
- If YouTube stream link provided, embed video player on event page
- Community chat link displayed prominently: "Have questions? Join our [Community Name]" (configured at org/group level)
- Speaker information with optional bio/photo
- Schedule breakdown for multi-session events
- Capacity indication (X of Y spots filled) if capacity is set
- Clear indication if event is streaming vs. in-person only
- All times displayed in user's timezone (with original event timezone noted)

**User Stories:**
- As an attendee, I can see all event details including whether streaming is available
- As an attendee, I can click on the embedded YouTube link to watch the stream during the event

### 4. Venue Management

#### 4.1 Venue Information
**Priority:** P0 (MVP)

**Requirements:**
- Venue includes: name, address, directions/notes, capacity
- Venues can be reused across events
- Venue library for quick selection
- Online/virtual venue option for remote events

**User Stories:**
- As a group admin, I can select from previously used venues or create a new one
- As an attendee, I can see the venue address and any special instructions

### 5. RSVP & Attendance

#### 5.1 RSVP Functionality
**Priority:** P0 (MVP)

**Requirements:**
- Users must be logged in to RSVP
- RSVP options: Attending, Not Attending
- Users can change RSVP up until event start time
- Capacity enforcement: If capacity is set, prevent RSVPs over venue capacity
- RSVP confirmation shown to user after RSVPing
- Public event view shows only RSVP count, not attendee names
- **Host/Admin view:** Event hosts and group admins can see full attendee list with names, emails, and RSVP timestamps
- User dashboard shows their upcoming RSVPs
- Hosts receive notifications at 80% and 100% capacity (if capacity is set)

**User Stories:**
- As an attendee, I can RSVP to an event and receive confirmation
- As an event host, I can see the list of attendees who have RSVP'd
- As an attendee, I receive a notification if an event I RSVP'd to is cancelled

### 6. Subscription & Notification System

#### 6.1 Subscription Model
**Priority:** P0 (MVP)

**Requirements:**
- Two-tier subscription model:
  - Organization-level: receive notifications for ALL groups
  - Group-level: receive notifications for specific group(s)
- Users can be subscribed to org + additional individual groups (no duplicate notifications)
- Subscription preferences stored per user
- Users can manage their subscriptions from their profile

**User Stories:**
- As a subscriber, I can choose to follow the entire organization or just specific groups
- As a subscriber, I can easily add or remove group subscriptions

#### 6.2 Notification Channels
**Priority:** 
- Email: P0 (MVP)
- SMS: P1 (Post-launch)
- Discord: P1 (Post-launch)
- Push (mobile): P2 (Future)

**Requirements:**
- Users opt-in to notification channels independently
- Notification preferences per subscription (can set different preferences for different groups)
- Notification triggers:
  - New event published in subscribed group
  - Event updates (time/venue changes)
  - Event reminders (24hr, 1hr before)
  - Event cancellations
  - RSVP confirmations

**Email Requirements (P0):**
- Event announcement emails
- Event reminder emails (configurable: 1 day, 1 hour before)
- Event update/cancellation emails
- ICS calendar file attachment for RSVP confirmations

**SMS Requirements (P1):**
- Same triggers as email
- User provides phone number and opts in
- SMS should be concise with link to event page

**Discord Requirements (P1):**
- Integration with Ford Utah Discord server
- Bot posts event announcements to relevant channels
- Users can opt-in to Discord DMs for reminders
- Link Discord account to platform account

**User Stories:**
- As a subscriber, I can choose to receive email and SMS notifications but not Discord
- As a subscriber, I receive a reminder 24 hours before an event I RSVP'd to

### 7. Calendar Integration

#### 7.1 ICS Export
**Priority:** P0 (MVP)

**Requirements:**
- Each event has an "Add to Calendar" button
- Generates ICS file compatible with Google Calendar, Apple Calendar, Outlook
- ICS includes: event title, description, location, time, organizer info
- Include YouTube link in description if available
- Users can export their full RSVP calendar

**User Stories:**
- As an attendee, I can add an event directly to my Google Calendar
- As a subscriber, I can subscribe to a calendar feed of all events in my subscribed groups

#### 7.2 Calendar Feeds (P1)
**Priority:** P1 (Post-launch)

**Requirements:**
- Webcal/ICS feed URL for organization (all events)
- Webcal/ICS feed URL per group
- Webcal/ICS feed for user's RSVP'd events
- Calendar feeds auto-update as events change

**User Stories:**
- As a subscriber, I can subscribe to a calendar feed that automatically updates with new events

### 8. User Management & Authentication

#### 8.1 User Accounts
**Priority:** P0 (MVP)

**Requirements:**
- User registration with email verification
- OAuth options: Google, GitHub (defer to P1 if time-constrained)
- User profile: name, email, phone (optional), timezone (required), Discord ID (optional), bio, avatar
- Users can update their profile information
- Users can delete their account (GDPR compliance)
- Timezone defaults to browser/system timezone on registration

**User Stories:**
- As a new user, I can create an account with my email
- As a user, I can link my Google account for easy login

#### 8.2 Password & Security
**Priority:** P0 (MVP)

**Requirements:**
- Password requirements: minimum 8 characters, mix of characters
- Password reset via email
- Session management with secure tokens
- Rate limiting on authentication endpoints

### 9. Discovery & Browsing

#### 9.1 Event Listing
**Priority:** P0 (MVP)

**Requirements:**
- Homepage shows upcoming events across all groups (chronological)
- Filter events by group
- Filter events by date range
- Search events by keyword
- Past events archive (view-only)

**User Stories:**
- As a visitor, I can browse upcoming events across all Ford Utah groups
- As a visitor, I can filter to see only Kubernetes events

#### 9.2 Group Directory
**Priority:** P0 (MVP)

**Requirements:**
- Directory page listing all user groups
- Each group card shows: name, description, logo, subscriber count, upcoming event count
- Click through to group page with group details and upcoming events

**User Stories:**
- As a visitor, I can explore all the user groups under Ford Utah Foundation
- As a potential subscriber, I can see how active each group is

### 10. Admin Dashboards

#### 10.1 Organization Admin Dashboard
**Priority:** P0 (MVP)

**Requirements:**
- Overview of all groups with key metrics
- List of organization-level subscribers
- Recent activity feed (new events, new subscribers)
- Quick links to manage groups and settings

**User Stories:**
- As an org admin, I have a dashboard showing the health of all user groups

#### 10.2 Group Admin Dashboard
**Priority:** P0 (MVP)

**Requirements:**
- Overview of group metrics (subscribers, upcoming events, total RSVPs)
- List of upcoming events with RSVP counts
- Quick event creation button
- List of group subscribers
- Recent activity for the group

**User Stories:**
- As a group admin, I can see at a glance how my events are performing

#### 10.3 Event Host View
**Priority:** P1 (Post-launch)

**Requirements:**
- Event-specific dashboard showing RSVP list
- Attendee contact information (email) for communication
- Export attendee list as CSV

**User Stories:**
- As an event host, I can see who has RSVP'd and reach out if needed

---

## User Flows

### Flow 1: New User Discovers and RSVPs to Event
1. User visits Ford Utah Foundation platform (public homepage)
2. Browses upcoming events or clicks on a specific group
3. Clicks on an event to see details
4. Clicks "RSVP" button → prompted to create account or log in
5. Creates account with email + password
6. Confirms RSVP
7. Receives RSVP confirmation email with ICS attachment
8. Receives reminder notification 24 hours before event

### Flow 2: User Subscribes to Organization
1. Logged-in user navigates to organization page
2. Clicks "Subscribe to All Events" button
3. Prompted to select notification preferences (email, SMS, Discord)
4. Confirms subscription
5. Receives welcome email
6. From that point forward, receives notifications for all new events across all groups

### Flow 3: Group Admin Creates Event
1. Group admin logs in and navigates to their group dashboard
2. Clicks "Create Event" button
3. Fills out event form:
   - Title, description
   - Date and time
   - Selects venue from library or creates new venue
   - Adds speaker information (optional)
   - Adds schedule breakdown (optional)
   - Adds YouTube stream link (optional)
   - Sets capacity
4. Saves as draft OR publishes immediately
5. Upon publishing:
   - All group subscribers receive notification
   - All org subscribers receive notification
   - Event appears on homepage and group page

### Flow 4: Event Host Adds Co-Host
1. Event host (who is not a group admin) navigates to their event
2. Clicks "Manage Hosts" button
3. Searches for user by email or username
4. Selects user to add as co-host
5. User receives invitation notification
6. User accepts co-host role
7. User now has edit permissions for that specific event

### Flow 5: Org Admin Creates New User Group
1. Org admin logs into admin dashboard
2. Clicks "Add New Group" button
3. Fills out group details: name, description, logo, primary location
4. Assigns initial group admin(s)
5. Saves group
6. Group appears in directory
7. Group admin receives notification of their new role

---

## Technical Requirements

### Performance
- Page load time < 2 seconds for event listings
- Support for 1000+ concurrent users
- Database queries optimized for hierarchical structure

### Accessibility
- WCAG 2.1 AA compliance
- Keyboard navigation support
- Screen reader compatibility
- Color contrast ratios meet standards
- Focus indicators visible

### Browser Support
- Modern browsers: Chrome, Firefox, Safari, Edge (latest 2 versions)
- Mobile responsive design (iOS Safari, Chrome Android)

### Security
- HTTPS only
- SQL injection prevention
- XSS protection
- CSRF tokens
- Rate limiting on all public endpoints
- Secure password storage (bcrypt or Argon2)

### Scalability
- Self-hosted but designed for horizontal scaling
- Database: PostgreSQL (chosen for robust relational model)
- Caching layer (Redis) for event listings
- Background job processing for notifications (Sidekiq/Celery)

---

## Open Questions & Decisions Needed

1. **Event Host Subscription Rule:** ✅ DECIDED - Event hosts must be subscribed to either the parent organization (all events) OR the specific group

4. **Recurring Events:** ✅ DECIDED - P0 MVP feature with full recurrence pattern support

7. **Email Service:** ✅ DECIDED - Use open source solution or AWS SES (most cost-effective)
   - **Rationale**: AWS SES offers $0.10 per 1,000 emails (extremely affordable)
   - Alternative: Self-hosted SMTP relay with Postfix/OpenSMTPD
   - Development: Mailhog or Mailpit for local testing

8. **SMS Service:** ✅ DECIDED - Twilio (industry standard, pay-as-you-go)
   - **Rationale**: No free open source alternatives for SMS that work reliably
   - Cost: ~$0.0079 per SMS (affordable at scale)
   - Can defer SMS to Phase 2 if budget is concern

9. **Analytics:** Do we want built-in analytics (attendance trends, popular events) for admins?

10. **API Access:** Should we expose a public API for third-party integrations?

11. **Host Assignment Enforcement:** ✅ DECIDED - Automatically subscribe users when they're assigned as event host
    - If user is not subscribed to org or group, system automatically creates group subscription
    - User receives notification that they've been subscribed due to host assignment
    - User can still manage their notification preferences (email/SMS/Discord)

12. **Event Time Display:** ✅ DECIDED - Show event in user's timezone with original timezone noted
    - Display format: "7:00 PM (in your timezone) • 8:00 PM MST (event time)"
    - Calendar exports use event timezone
    - Logged-out users see event timezone only

13. **Default Community Chat:** ✅ DECIDED - Groups inherit organization's community chat if not configured
    - Groups can override with their own community chat link
    - If group doesn't configure one, automatically use org-level chat
    - Display: "Have questions? Join our [Community Name]" with appropriate link

14. **Venue Timezone:** ✅ DECIDED - Venues should store default timezone
    - Helps with event creation (auto-fill timezone from venue)
    - Reduces manual timezone selection errors

---

## Success Metrics

### Launch Metrics (First 3 Months)
- All Ford Utah user groups migrated from Meetup.com
- 80%+ of previous Meetup members have accounts
- 50%+ of members subscribe to notifications
- Average event RSVP rate maintained or improved vs. Meetup

### Ongoing Metrics
- Event creation rate (events per month)
- Subscription growth (org-level and group-level)
- RSVP conversion rate (views to RSVPs)
- Notification open rates
- Actual attendance vs. RSVP rate
- Calendar integration usage (ICS downloads)

---

## Future Enhancements (Post-V1)

### Phase 2
- Recurring events
- Waitlist functionality
- Event check-in system (QR codes)
- Enhanced analytics dashboard
- Calendar feed subscriptions

### Phase 3
- Mobile apps (iOS/Android)
- Push notifications
- Multi-organization support (white-label platform)
- Event photos/galleries
- Attendee networking features

### Phase 4
- Paid events / ticketing
- Sponsor management
- Badge/achievement system for regular attendees
- Integration marketplace (Slack, Teams, etc.)

---

## Document Status

**Version:** 0.1 (Draft)  
**Last Updated:** 2025-11-12  
**Status:** In Progress - Needs review and decisions on open questions  
**Next Steps:** 
1. Review and answer open questions
2. Finalize notification channel priorities
3. Create architectural design document
4. Define API contracts
5. Create database schema document

