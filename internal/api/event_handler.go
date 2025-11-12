package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/forgeutah/taikai/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

// EventHandler handles event-related requests
type EventHandler struct {
	db      *sql.DB
	checker *auth.PermissionChecker
}

// NewEventHandler creates a new event handler
func NewEventHandler(db *sql.DB, checker *auth.PermissionChecker) *EventHandler {
	return &EventHandler{
		db:      db,
		checker: checker,
	}
}

// Event response structure
type EventResponse struct {
	ID                 string     `json:"id"`
	GroupID            string     `json:"group_id"`
	Title              string     `json:"title"`
	Slug               string     `json:"slug"`
	Description        string     `json:"description"`
	Status             string     `json:"status"`
	StartTime          string     `json:"start_time"`
	EndTime            string     `json:"end_time"`
	Timezone           string     `json:"timezone"`
	VenueID            *string    `json:"venue_id"`
	Capacity           *int       `json:"capacity"`
	YouTubeURL         *string    `json:"youtube_url"`
	ParentEventID      *string    `json:"parent_event_id"`
	RecurrenceRule     *string    `json:"recurrence_rule"`
	RecurrenceEndDate  *string    `json:"recurrence_end_date"`
	RecurrenceCount    *int       `json:"recurrence_count"`
	IsRecurringParent  bool       `json:"is_recurring_parent"`
	CreatedBy          string     `json:"created_by"`
	CreatedAt          string     `json:"created_at"`
	UpdatedAt          string     `json:"updated_at"`
}

// CreateEventRequest for creating a new event
type CreateEventRequest struct {
	GroupID       string  `json:"group_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Status        string  `json:"status"` // draft, published, cancelled
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	Timezone      string  `json:"timezone"`
	VenueID       *string `json:"venue_id"`
	Capacity      *int    `json:"capacity"`
	YouTubeURL    *string `json:"youtube_url"`
}

// UpdateEventRequest for updating event
type UpdateEventRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Timezone    *string `json:"timezone"`
	VenueID     *string `json:"venue_id"`
	Capacity    *int    `json:"capacity"`
	YouTubeURL  *string `json:"youtube_url"`
}

// ListEventsResponse for paginated event list
type ListEventsResponse struct {
	Events []EventResponse `json:"events"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// GetEvent handles GET /api/v1/events/:eventId
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	if _, err := uuid.Parse(eventID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid event ID format")
		return
	}

	event, err := h.getEventByID(r, eventID)
	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Event not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get event")
		return
	}

	RespondSuccess(w, http.StatusOK, event)
}

// GetEventBySlug handles GET /api/v1/events/slug/:slug
func (h *EventHandler) GetEventBySlug(w http.ResponseWriter, r *http.Request) {
	eventSlug := chi.URLParam(r, "slug")
	if eventSlug == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event slug is required")
		return
	}

	query := `
		SELECT id, group_id, title, slug, description, status,
		       start_time, end_time, timezone, venue_id, capacity,
		       youtube_url, parent_event_id, recurrence_rule,
		       recurrence_end_date, recurrence_count, is_recurring_parent,
		       created_by, created_at, updated_at
		FROM events
		WHERE slug = $1
	`

	event, err := h.scanEvent(h.db.QueryRowContext(r.Context(), query, eventSlug))
	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Event not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get event")
		return
	}

	RespondSuccess(w, http.StatusOK, event)
}

// ListEvents handles GET /api/v1/events with filters and pagination
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Parse filter parameters
	groupID := r.URL.Query().Get("group_id")
	status := r.URL.Query().Get("status")
	startAfter := r.URL.Query().Get("start_after")
	startBefore := r.URL.Query().Get("start_before")

	// Parse UUID if provided
	var groupUUID *uuid.UUID
	if groupID != "" {
		parsed, err := uuid.Parse(groupID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid group_id format")
			return
		}
		groupUUID = &parsed
	}

	// Parse dates
	var startAfterTime, startBeforeTime *time.Time
	if startAfter != "" {
		t, err := time.Parse(time.RFC3339, startAfter)
		if err != nil {
			RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid start_after format (use RFC3339)")
			return
		}
		startAfterTime = &t
	}
	if startBefore != "" {
		t, err := time.Parse(time.RFC3339, startBefore)
		if err != nil {
			RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid start_before format (use RFC3339)")
			return
		}
		startBeforeTime = &t
	}

	// Count total
	countQuery := `
		SELECT COUNT(*)
		FROM events e
		WHERE
			CASE WHEN $1::uuid IS NOT NULL THEN e.group_id = $1 ELSE TRUE END
			AND CASE WHEN $2::text != '' THEN e.status::text = $2 ELSE TRUE END
			AND CASE WHEN $3::timestamptz IS NOT NULL THEN e.start_time >= $3 ELSE TRUE END
			AND CASE WHEN $4::timestamptz IS NOT NULL THEN e.start_time <= $4 ELSE TRUE END
	`

	var total int
	err := h.db.QueryRowContext(r.Context(), countQuery, groupUUID, status, startAfterTime, startBeforeTime).Scan(&total)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to count events")
		return
	}

	// Query events
	query := `
		SELECT id, group_id, title, slug, description, status,
		       start_time, end_time, timezone, venue_id, capacity,
		       youtube_url, parent_event_id, recurrence_rule,
		       recurrence_end_date, recurrence_count, is_recurring_parent,
		       created_by, created_at, updated_at
		FROM events
		WHERE
			CASE WHEN $1::uuid IS NOT NULL THEN group_id = $1 ELSE TRUE END
			AND CASE WHEN $2::text != '' THEN status::text = $2 ELSE TRUE END
			AND CASE WHEN $3::timestamptz IS NOT NULL THEN start_time >= $3 ELSE TRUE END
			AND CASE WHEN $4::timestamptz IS NOT NULL THEN start_time <= $4 ELSE TRUE END
		ORDER BY start_time ASC
		LIMIT $5 OFFSET $6
	`

	rows, err := h.db.QueryContext(r.Context(), query, groupUUID, status, startAfterTime, startBeforeTime, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to list events")
		return
	}
	defer rows.Close()

	events := []EventResponse{}
	for rows.Next() {
		event, err := h.scanEvent(rows)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan event")
			return
		}
		events = append(events, event)
	}

	response := ListEventsResponse{
		Events: events,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}

	RespondSuccess(w, http.StatusOK, response)
}

// CreateEvent handles POST /api/v1/events
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.GroupID == "" || req.Title == "" || req.Description == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "group_id, title, and description are required")
		return
	}
	if req.StartTime == "" || req.EndTime == "" || req.Timezone == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "start_time, end_time, and timezone are required")
		return
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid start_time format (use RFC3339)")
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid end_time format (use RFC3339)")
		return
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "end_time must be after start_time")
		return
	}

	// Generate slug from title
	eventSlug := slug.Make(req.Title)

	// Check for slug uniqueness, append UUID if needed
	var exists bool
	err = h.db.QueryRowContext(r.Context(), "SELECT EXISTS(SELECT 1 FROM events WHERE slug = $1)", eventSlug).Scan(&exists)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to check slug uniqueness")
		return
	}
	if exists {
		eventSlug = eventSlug + "-" + uuid.New().String()[:8]
	}

	// Default status to draft if not provided
	if req.Status == "" {
		req.Status = "draft"
	}

	// Validate status
	if req.Status != "draft" && req.Status != "published" && req.Status != "cancelled" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "status must be draft, published, or cancelled")
		return
	}

	// Create event
	query := `
		INSERT INTO events (
			group_id, title, slug, description, status, start_time, end_time,
			timezone, venue_id, capacity, youtube_url, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, group_id, title, slug, description, status, start_time, end_time,
		          timezone, venue_id, capacity, youtube_url, parent_event_id,
		          recurrence_rule, recurrence_end_date, recurrence_count,
		          is_recurring_parent, created_by, created_at, updated_at
	`

	row := h.db.QueryRowContext(r.Context(), query,
		req.GroupID, req.Title, eventSlug, req.Description, req.Status,
		startTime, endTime, req.Timezone, req.VenueID, req.Capacity,
		req.YouTubeURL, userID,
	)

	event, err := h.scanEvent(row)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create event")
		return
	}

	// Auto-assign creator as event host
	_, err = h.db.ExecContext(r.Context(),
		"INSERT INTO event_hosts (user_id, event_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		userID, event.ID,
	)
	if err != nil {
		// Log but don't fail the request
		// In production, we'd log this properly
	}

	RespondSuccess(w, http.StatusOK, event)
}

// UpdateEvent handles PATCH /api/v1/events/:eventId
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	if _, err := uuid.Parse(eventID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid event ID format")
		return
	}

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Parse times if provided
	var startTime, endTime *time.Time
	if req.StartTime != nil {
		t, err := time.Parse(time.RFC3339, *req.StartTime)
		if err != nil {
			RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid start_time format (use RFC3339)")
			return
		}
		startTime = &t
	}
	if req.EndTime != nil {
		t, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid end_time format (use RFC3339)")
			return
		}
		endTime = &t
	}

	// Validate status if provided
	if req.Status != nil {
		if *req.Status != "draft" && *req.Status != "published" && *req.Status != "cancelled" {
			RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "status must be draft, published, or cancelled")
			return
		}
	}

	query := `
		UPDATE events
		SET
			title = COALESCE($2, title),
			description = COALESCE($3, description),
			status = COALESCE($4, status),
			start_time = COALESCE($5, start_time),
			end_time = COALESCE($6, end_time),
			timezone = COALESCE($7, timezone),
			venue_id = COALESCE($8, venue_id),
			capacity = COALESCE($9, capacity),
			youtube_url = COALESCE($10, youtube_url),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, group_id, title, slug, description, status, start_time, end_time,
		          timezone, venue_id, capacity, youtube_url, parent_event_id,
		          recurrence_rule, recurrence_end_date, recurrence_count,
		          is_recurring_parent, created_by, created_at, updated_at
	`

	row := h.db.QueryRowContext(r.Context(), query,
		eventID, req.Title, req.Description, req.Status, startTime, endTime,
		req.Timezone, req.VenueID, req.Capacity, req.YouTubeURL,
	)

	event, err := h.scanEvent(row)
	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Event not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update event")
		return
	}

	RespondSuccess(w, http.StatusOK, event)
}

// DeleteEvent handles DELETE /api/v1/events/:eventId
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	if _, err := uuid.Parse(eventID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid event ID format")
		return
	}

	_, err := h.db.ExecContext(r.Context(), "DELETE FROM events WHERE id = $1", eventID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to delete event")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func (h *EventHandler) getEventByID(r *http.Request, eventID string) (EventResponse, error) {
	query := `
		SELECT id, group_id, title, slug, description, status,
		       start_time, end_time, timezone, venue_id, capacity,
		       youtube_url, parent_event_id, recurrence_rule,
		       recurrence_end_date, recurrence_count, is_recurring_parent,
		       created_by, created_at, updated_at
		FROM events
		WHERE id = $1
	`
	return h.scanEvent(h.db.QueryRowContext(r.Context(), query, eventID))
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func (h *EventHandler) scanEvent(s scanner) (EventResponse, error) {
	var event EventResponse
	var venueID, youtubeURL, parentEventID, recurrenceRule, recurrenceEndDate sql.NullString
	var capacity, recurrenceCount sql.NullInt64

	err := s.Scan(
		&event.ID, &event.GroupID, &event.Title, &event.Slug, &event.Description,
		&event.Status, &event.StartTime, &event.EndTime, &event.Timezone,
		&venueID, &capacity, &youtubeURL, &parentEventID, &recurrenceRule,
		&recurrenceEndDate, &recurrenceCount, &event.IsRecurringParent,
		&event.CreatedBy, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return event, err
	}

	if venueID.Valid {
		event.VenueID = &venueID.String
	}
	if capacity.Valid {
		cap := int(capacity.Int64)
		event.Capacity = &cap
	}
	if youtubeURL.Valid {
		event.YouTubeURL = &youtubeURL.String
	}
	if parentEventID.Valid {
		event.ParentEventID = &parentEventID.String
	}
	if recurrenceRule.Valid {
		event.RecurrenceRule = &recurrenceRule.String
	}
	if recurrenceEndDate.Valid {
		event.RecurrenceEndDate = &recurrenceEndDate.String
	}
	if recurrenceCount.Valid {
		count := int(recurrenceCount.Int64)
		event.RecurrenceCount = &count
	}

	return event, nil
}

// ====================
// Event Hosts Management
// ====================

// EventHostResponse represents an event host
type EventHostResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	EventID   string  `json:"event_id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
	CreatedAt string  `json:"created_at"`
}

// AddEventHostRequest for adding a host
type AddEventHostRequest struct {
	UserID string `json:"user_id"`
}

// GetEventHosts handles GET /api/v1/events/:eventId/hosts
func (h *EventHandler) GetEventHosts(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	query := `
		SELECT eh.id, eh.user_id, eh.event_id, eh.created_at,
		       u.email, u.name, u.avatar_url
		FROM event_hosts eh
		JOIN users u ON eh.user_id = u.id
		WHERE eh.event_id = $1
		ORDER BY eh.created_at ASC
	`

	rows, err := h.db.QueryContext(r.Context(), query, eventID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get event hosts")
		return
	}
	defer rows.Close()

	hosts := []EventHostResponse{}
	for rows.Next() {
		var host EventHostResponse
		var avatarURL sql.NullString

		err := rows.Scan(&host.ID, &host.UserID, &host.EventID, &host.CreatedAt,
			&host.Email, &host.Name, &avatarURL)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan host")
			return
		}

		if avatarURL.Valid {
			host.AvatarURL = &avatarURL.String
		}

		hosts = append(hosts, host)
	}

	RespondSuccess(w, http.StatusOK, hosts)
}

// AddEventHost handles POST /api/v1/events/:eventId/hosts
func (h *EventHandler) AddEventHost(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	var req AddEventHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "user_id is required")
		return
	}

	// Add host
	query := `
		INSERT INTO event_hosts (user_id, event_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, event_id) DO NOTHING
		RETURNING id, user_id, event_id, created_at
	`

	var host struct {
		ID        string
		UserID    string
		EventID   string
		CreatedAt string
	}

	err := h.db.QueryRowContext(r.Context(), query, req.UserID, eventID).Scan(
		&host.ID, &host.UserID, &host.EventID, &host.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Host already exists, return success
			RespondSuccess(w, http.StatusOK, map[string]string{"message": "Host already added"})
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to add event host")
		return
	}

	// TODO: Check if user is subscribed to the group and auto-subscribe if not
	// This will be implemented when we add the subscription system

	RespondSuccess(w, http.StatusOK, host)
}

// RemoveEventHost handles DELETE /api/v1/events/:eventId/hosts/:userId
func (h *EventHandler) RemoveEventHost(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	userIDToRemove := chi.URLParam(r, "userId")

	if eventID == "" || userIDToRemove == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID and user ID are required")
		return
	}

	_, err := h.db.ExecContext(r.Context(),
		"DELETE FROM event_hosts WHERE user_id = $1 AND event_id = $2",
		userIDToRemove, eventID,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to remove event host")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ====================
// Speakers Management
// ====================

// SpeakerResponse represents a speaker
type SpeakerResponse struct {
	ID           string  `json:"id"`
	EventID      string  `json:"event_id"`
	Name         string  `json:"name"`
	Bio          *string `json:"bio"`
	AvatarURL    *string `json:"avatar_url"`
	DisplayOrder int     `json:"display_order"`
	CreatedAt    string  `json:"created_at"`
}

// CreateSpeakerRequest for creating a speaker
type CreateSpeakerRequest struct {
	Name         string  `json:"name"`
	Bio          *string `json:"bio"`
	AvatarURL    *string `json:"avatar_url"`
	DisplayOrder int     `json:"display_order"`
}

// UpdateSpeakerRequest for updating a speaker
type UpdateSpeakerRequest struct {
	Name         *string `json:"name"`
	Bio          *string `json:"bio"`
	AvatarURL    *string `json:"avatar_url"`
	DisplayOrder *int    `json:"display_order"`
}

// GetEventSpeakers handles GET /api/v1/events/:eventId/speakers
func (h *EventHandler) GetEventSpeakers(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	query := `
		SELECT id, event_id, name, bio, avatar_url, display_order, created_at
		FROM event_speakers
		WHERE event_id = $1
		ORDER BY display_order ASC, created_at ASC
	`

	rows, err := h.db.QueryContext(r.Context(), query, eventID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get speakers")
		return
	}
	defer rows.Close()

	speakers := []SpeakerResponse{}
	for rows.Next() {
		var speaker SpeakerResponse
		var bio, avatarURL sql.NullString

		err := rows.Scan(&speaker.ID, &speaker.EventID, &speaker.Name, &bio, &avatarURL,
			&speaker.DisplayOrder, &speaker.CreatedAt)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan speaker")
			return
		}

		if bio.Valid {
			speaker.Bio = &bio.String
		}
		if avatarURL.Valid {
			speaker.AvatarURL = &avatarURL.String
		}

		speakers = append(speakers, speaker)
	}

	RespondSuccess(w, http.StatusOK, speakers)
}

// CreateSpeaker handles POST /api/v1/events/:eventId/speakers
func (h *EventHandler) CreateSpeaker(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	var req CreateSpeakerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "name is required")
		return
	}

	query := `
		INSERT INTO event_speakers (event_id, name, bio, avatar_url, display_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, event_id, name, bio, avatar_url, display_order, created_at
	`

	var speaker SpeakerResponse
	var bio, avatarURL sql.NullString

	err := h.db.QueryRowContext(r.Context(), query,
		eventID, req.Name, req.Bio, req.AvatarURL, req.DisplayOrder,
	).Scan(&speaker.ID, &speaker.EventID, &speaker.Name, &bio, &avatarURL,
		&speaker.DisplayOrder, &speaker.CreatedAt)

	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create speaker")
		return
	}

	if bio.Valid {
		speaker.Bio = &bio.String
	}
	if avatarURL.Valid {
		speaker.AvatarURL = &avatarURL.String
	}

	RespondSuccess(w, http.StatusOK, speaker)
}

// UpdateSpeaker handles PATCH /api/v1/events/:eventId/speakers/:speakerId
func (h *EventHandler) UpdateSpeaker(w http.ResponseWriter, r *http.Request) {
	speakerID := chi.URLParam(r, "speakerId")
	if speakerID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "speaker ID is required")
		return
	}

	var req UpdateSpeakerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	query := `
		UPDATE event_speakers
		SET
			name = COALESCE($2, name),
			bio = COALESCE($3, bio),
			avatar_url = COALESCE($4, avatar_url),
			display_order = COALESCE($5, display_order)
		WHERE id = $1
		RETURNING id, event_id, name, bio, avatar_url, display_order, created_at
	`

	var speaker SpeakerResponse
	var bio, avatarURL sql.NullString

	err := h.db.QueryRowContext(r.Context(), query,
		speakerID, req.Name, req.Bio, req.AvatarURL, req.DisplayOrder,
	).Scan(&speaker.ID, &speaker.EventID, &speaker.Name, &bio, &avatarURL,
		&speaker.DisplayOrder, &speaker.CreatedAt)

	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Speaker not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update speaker")
		return
	}

	if bio.Valid {
		speaker.Bio = &bio.String
	}
	if avatarURL.Valid {
		speaker.AvatarURL = &avatarURL.String
	}

	RespondSuccess(w, http.StatusOK, speaker)
}

// DeleteSpeaker handles DELETE /api/v1/events/:eventId/speakers/:speakerId
func (h *EventHandler) DeleteSpeaker(w http.ResponseWriter, r *http.Request) {
	speakerID := chi.URLParam(r, "speakerId")
	if speakerID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "speaker ID is required")
		return
	}

	_, err := h.db.ExecContext(r.Context(), "DELETE FROM event_speakers WHERE id = $1", speakerID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to delete speaker")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ====================
// Schedule Management
// ====================

// ScheduleItemResponse represents a schedule item
type ScheduleItemResponse struct {
	ID                string  `json:"id"`
	EventID           string  `json:"event_id"`
	TimeOffsetMinutes int     `json:"time_offset_minutes"`
	Title             string  `json:"title"`
	Description       *string `json:"description"`
	DisplayOrder      int     `json:"display_order"`
	CreatedAt         string  `json:"created_at"`
}

// CreateScheduleItemRequest for creating a schedule item
type CreateScheduleItemRequest struct {
	TimeOffsetMinutes int     `json:"time_offset_minutes"`
	Title             string  `json:"title"`
	Description       *string `json:"description"`
	DisplayOrder      int     `json:"display_order"`
}

// UpdateScheduleItemRequest for updating a schedule item
type UpdateScheduleItemRequest struct {
	TimeOffsetMinutes *int    `json:"time_offset_minutes"`
	Title             *string `json:"title"`
	Description       *string `json:"description"`
	DisplayOrder      *int    `json:"display_order"`
}

// GetEventSchedule handles GET /api/v1/events/:eventId/schedule
func (h *EventHandler) GetEventSchedule(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	query := `
		SELECT id, event_id, time_offset_minutes, title, description, display_order, created_at
		FROM event_schedules
		WHERE event_id = $1
		ORDER BY time_offset_minutes ASC, display_order ASC
	`

	rows, err := h.db.QueryContext(r.Context(), query, eventID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get schedule")
		return
	}
	defer rows.Close()

	scheduleItems := []ScheduleItemResponse{}
	for rows.Next() {
		var item ScheduleItemResponse
		var description sql.NullString

		err := rows.Scan(&item.ID, &item.EventID, &item.TimeOffsetMinutes, &item.Title,
			&description, &item.DisplayOrder, &item.CreatedAt)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan schedule item")
			return
		}

		if description.Valid {
			item.Description = &description.String
		}

		scheduleItems = append(scheduleItems, item)
	}

	RespondSuccess(w, http.StatusOK, scheduleItems)
}

// CreateScheduleItem handles POST /api/v1/events/:eventId/schedule
func (h *EventHandler) CreateScheduleItem(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	var req CreateScheduleItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "title is required")
		return
	}

	query := `
		INSERT INTO event_schedules (event_id, time_offset_minutes, title, description, display_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, event_id, time_offset_minutes, title, description, display_order, created_at
	`

	var item ScheduleItemResponse
	var description sql.NullString

	err := h.db.QueryRowContext(r.Context(), query,
		eventID, req.TimeOffsetMinutes, req.Title, req.Description, req.DisplayOrder,
	).Scan(&item.ID, &item.EventID, &item.TimeOffsetMinutes, &item.Title,
		&description, &item.DisplayOrder, &item.CreatedAt)

	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create schedule item")
		return
	}

	if description.Valid {
		item.Description = &description.String
	}

	RespondSuccess(w, http.StatusOK, item)
}

// UpdateScheduleItem handles PATCH /api/v1/events/:eventId/schedule/:scheduleId
func (h *EventHandler) UpdateScheduleItem(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "scheduleId")
	if scheduleID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "schedule ID is required")
		return
	}

	var req UpdateScheduleItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	query := `
		UPDATE event_schedules
		SET
			time_offset_minutes = COALESCE($2, time_offset_minutes),
			title = COALESCE($3, title),
			description = COALESCE($4, description),
			display_order = COALESCE($5, display_order)
		WHERE id = $1
		RETURNING id, event_id, time_offset_minutes, title, description, display_order, created_at
	`

	var item ScheduleItemResponse
	var description sql.NullString

	err := h.db.QueryRowContext(r.Context(), query,
		scheduleID, req.TimeOffsetMinutes, req.Title, req.Description, req.DisplayOrder,
	).Scan(&item.ID, &item.EventID, &item.TimeOffsetMinutes, &item.Title,
		&description, &item.DisplayOrder, &item.CreatedAt)

	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Schedule item not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update schedule item")
		return
	}

	if description.Valid {
		item.Description = &description.String
	}

	RespondSuccess(w, http.StatusOK, item)
}

// DeleteScheduleItem handles DELETE /api/v1/events/:eventId/schedule/:scheduleId
func (h *EventHandler) DeleteScheduleItem(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "scheduleId")
	if scheduleID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "schedule ID is required")
		return
	}

	_, err := h.db.ExecContext(r.Context(), "DELETE FROM event_schedules WHERE id = $1", scheduleID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to delete schedule item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
