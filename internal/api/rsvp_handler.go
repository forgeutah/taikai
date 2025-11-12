package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/forgeutah/taikai/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RSVPHandler handles RSVP-related requests
type RSVPHandler struct {
	db      *sql.DB
	checker *auth.PermissionChecker
}

// NewRSVPHandler creates a new RSVP handler
func NewRSVPHandler(db *sql.DB, checker *auth.PermissionChecker) *RSVPHandler {
	return &RSVPHandler{
		db:      db,
		checker: checker,
	}
}

// RSVPResponse represents an RSVP
type RSVPResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	EventID   string  `json:"event_id"`
	Status    string  `json:"status"`
	Email     *string `json:"email,omitempty"`     // Only for hosts/admins
	Name      *string `json:"name,omitempty"`      // Only for hosts/admins
	AvatarURL *string `json:"avatar_url,omitempty"` // Only for hosts/admins
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CreateRSVPRequest for creating/updating an RSVP
type CreateRSVPRequest struct {
	Status string `json:"status"` // "attending" or "not_attending"
}

// EventRSVPsResponse for listing RSVPs
type EventRSVPsResponse struct {
	Count int            `json:"count"`
	RSVPs []RSVPResponse `json:"rsvps,omitempty"` // Only for hosts/admins
}

// UserRSVPWithEventResponse for user's RSVPs with event details
type UserRSVPWithEventResponse struct {
	RSVP  RSVPResponse  `json:"rsvp"`
	Event EventResponse `json:"event"`
}

// UserHostedEventResponse for events user is hosting with RSVP count
type UserHostedEventResponse struct {
	Event     EventResponse `json:"event"`
	RSVPCount int           `json:"rsvp_count"`
}

// CreateOrUpdateRSVP handles POST /api/v1/events/:eventId/rsvp
func (h *RSVPHandler) CreateOrUpdateRSVP(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	if _, err := uuid.Parse(eventID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid event ID format")
		return
	}

	var req CreateRSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate status
	if req.Status != "attending" && req.Status != "not_attending" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "status must be 'attending' or 'not_attending'")
		return
	}

	// Check if event exists and get capacity
	var capacity sql.NullInt64
	err := h.db.QueryRowContext(r.Context(), "SELECT capacity FROM events WHERE id = $1", eventID).Scan(&capacity)
	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Event not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to check event")
		return
	}

	// If attending, check capacity
	if req.Status == "attending" && capacity.Valid {
		// Get current RSVP count (excluding this user's existing RSVP)
		var currentCount int
		countQuery := `
			SELECT COUNT(*)
			FROM event_rsvps
			WHERE event_id = $1 AND status = 'attending' AND user_id != $2
		`
		err := h.db.QueryRowContext(r.Context(), countQuery, eventID, userID).Scan(&currentCount)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to check capacity")
			return
		}

		// Check if at capacity
		if currentCount >= int(capacity.Int64) {
			RespondError(w, http.StatusConflict, "event_full", "Event is at capacity")
			return
		}
	}

	// Create or update RSVP
	query := `
		INSERT INTO event_rsvps (user_id, event_id, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, event_id)
		DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = NOW()
		RETURNING id, user_id, event_id, status, created_at, updated_at
	`

	var rsvp RSVPResponse
	err = h.db.QueryRowContext(r.Context(), query, userID, eventID, req.Status).Scan(
		&rsvp.ID, &rsvp.UserID, &rsvp.EventID, &rsvp.Status, &rsvp.CreatedAt, &rsvp.UpdatedAt,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create RSVP")
		return
	}

	// Check if we should send capacity notifications
	if req.Status == "attending" && capacity.Valid {
		go h.checkCapacityNotifications(r.Context(), eventID, int(capacity.Int64))
	}

	RespondSuccess(w, http.StatusOK, rsvp)
}

// DeleteRSVP handles DELETE /api/v1/events/:eventId/rsvp
func (h *RSVPHandler) DeleteRSVP(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	_, err := h.db.ExecContext(r.Context(), "DELETE FROM event_rsvps WHERE user_id = $1 AND event_id = $2", userID, eventID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to delete RSVP")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEventRSVPs handles GET /api/v1/events/:eventId/rsvps
// Public: returns only count
// Host/Admin: returns full list with user details
func (h *RSVPHandler) GetEventRSVPs(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventId")
	if eventID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "event ID is required")
		return
	}

	// Get RSVP count
	var count int
	err := h.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM event_rsvps WHERE event_id = $1 AND status = 'attending'", eventID).Scan(&count)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get RSVP count")
		return
	}

	// Check if user is authenticated and has permission to see details
	userID := auth.GetUserIDFromContext(r.Context())
	canSeeDetails := false

	if userID != "" {
		// Check if user can manage this event
		canManage, err := h.checker.CanManageEvent(r.Context(), userID, eventID)
		if err == nil && canManage {
			canSeeDetails = true
		}
	}

	response := EventRSVPsResponse{
		Count: count,
	}

	// If user can see details, include full RSVP list
	if canSeeDetails {
		query := `
			SELECT r.id, r.user_id, r.event_id, r.status, r.created_at, r.updated_at,
			       u.email, u.name, u.avatar_url
			FROM event_rsvps r
			JOIN users u ON r.user_id = u.id
			WHERE r.event_id = $1
			ORDER BY r.created_at DESC
		`

		rows, err := h.db.QueryContext(r.Context(), query, eventID)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get RSVPs")
			return
		}
		defer rows.Close()

		rsvps := []RSVPResponse{}
		for rows.Next() {
			var rsvp RSVPResponse
			var email, name string
			var avatarURL sql.NullString

			err := rows.Scan(&rsvp.ID, &rsvp.UserID, &rsvp.EventID, &rsvp.Status,
				&rsvp.CreatedAt, &rsvp.UpdatedAt, &email, &name, &avatarURL)
			if err != nil {
				RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan RSVP")
				return
			}

			rsvp.Email = &email
			rsvp.Name = &name
			if avatarURL.Valid {
				rsvp.AvatarURL = &avatarURL.String
			}

			rsvps = append(rsvps, rsvp)
		}

		response.RSVPs = rsvps
	}

	RespondSuccess(w, http.StatusOK, response)
}

// GetMyRSVPs handles GET /api/v1/me/rsvps
// Returns user's upcoming RSVPs with event details
func (h *RSVPHandler) GetMyRSVPs(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM event_rsvps r
		JOIN events e ON r.event_id = e.id
		WHERE r.user_id = $1 AND r.status = 'attending' AND e.start_time > NOW()
	`
	err := h.db.QueryRowContext(r.Context(), countQuery, userID).Scan(&total)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to count RSVPs")
		return
	}

	// Get RSVPs with event details
	query := `
		SELECT r.id, r.user_id, r.event_id, r.status, r.created_at, r.updated_at,
		       e.id, e.group_id, e.title, e.slug, e.description, e.status,
		       e.start_time, e.end_time, e.timezone, e.venue_id, e.capacity,
		       e.youtube_url, e.parent_event_id, e.recurrence_rule,
		       e.recurrence_end_date, e.recurrence_count, e.is_recurring_parent,
		       e.created_by, e.created_at, e.updated_at
		FROM event_rsvps r
		JOIN events e ON r.event_id = e.id
		WHERE r.user_id = $1 AND r.status = 'attending' AND e.start_time > NOW()
		ORDER BY e.start_time ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.QueryContext(r.Context(), query, userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get RSVPs")
		return
	}
	defer rows.Close()

	results := []UserRSVPWithEventResponse{}
	for rows.Next() {
		var rsvp RSVPResponse
		var event EventResponse
		var venueID, youtubeURL, parentEventID, recurrenceRule, recurrenceEndDate sql.NullString
		var capacity, recurrenceCount sql.NullInt64

		err := rows.Scan(
			&rsvp.ID, &rsvp.UserID, &rsvp.EventID, &rsvp.Status, &rsvp.CreatedAt, &rsvp.UpdatedAt,
			&event.ID, &event.GroupID, &event.Title, &event.Slug, &event.Description,
			&event.Status, &event.StartTime, &event.EndTime, &event.Timezone,
			&venueID, &capacity, &youtubeURL, &parentEventID, &recurrenceRule,
			&recurrenceEndDate, &recurrenceCount, &event.IsRecurringParent,
			&event.CreatedBy, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan RSVP")
			return
		}

		// Convert nullable fields
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

		results = append(results, UserRSVPWithEventResponse{
			RSVP:  rsvp,
			Event: event,
		})
	}

	response := struct {
		RSVPs []UserRSVPWithEventResponse `json:"rsvps"`
		Total int                         `json:"total"`
		Page  int                         `json:"page"`
		Limit int                         `json:"limit"`
	}{
		RSVPs: results,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	RespondSuccess(w, http.StatusOK, response)
}

// GetMyHostedEvents handles GET /api/v1/me/events
// Returns events where user is a host, with RSVP counts
func (h *RSVPHandler) GetMyHostedEvents(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(DISTINCT e.id)
		FROM events e
		JOIN event_hosts eh ON e.id = eh.event_id
		WHERE eh.user_id = $1 AND e.start_time > NOW()
	`
	err := h.db.QueryRowContext(r.Context(), countQuery, userID).Scan(&total)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to count events")
		return
	}

	// Get hosted events with RSVP counts
	query := `
		SELECT DISTINCT e.id, e.group_id, e.title, e.slug, e.description, e.status,
		       e.start_time, e.end_time, e.timezone, e.venue_id, e.capacity,
		       e.youtube_url, e.parent_event_id, e.recurrence_rule,
		       e.recurrence_end_date, e.recurrence_count, e.is_recurring_parent,
		       e.created_by, e.created_at, e.updated_at,
		       (SELECT COUNT(*) FROM event_rsvps WHERE event_id = e.id AND status = 'attending') as rsvp_count
		FROM events e
		JOIN event_hosts eh ON e.id = eh.event_id
		WHERE eh.user_id = $1 AND e.start_time > NOW()
		ORDER BY e.start_time ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.QueryContext(r.Context(), query, userID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get events")
		return
	}
	defer rows.Close()

	results := []UserHostedEventResponse{}
	for rows.Next() {
		var event EventResponse
		var rsvpCount int
		var venueID, youtubeURL, parentEventID, recurrenceRule, recurrenceEndDate sql.NullString
		var capacity, recurrenceCount sql.NullInt64

		err := rows.Scan(
			&event.ID, &event.GroupID, &event.Title, &event.Slug, &event.Description,
			&event.Status, &event.StartTime, &event.EndTime, &event.Timezone,
			&venueID, &capacity, &youtubeURL, &parentEventID, &recurrenceRule,
			&recurrenceEndDate, &recurrenceCount, &event.IsRecurringParent,
			&event.CreatedBy, &event.CreatedAt, &event.UpdatedAt, &rsvpCount,
		)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan event")
			return
		}

		// Convert nullable fields
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

		results = append(results, UserHostedEventResponse{
			Event:     event,
			RSVPCount: rsvpCount,
		})
	}

	response := struct {
		Events []UserHostedEventResponse `json:"events"`
		Total  int                       `json:"total"`
		Page   int                       `json:"page"`
		Limit  int                       `json:"limit"`
	}{
		Events: results,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}

	RespondSuccess(w, http.StatusOK, response)
}

// checkCapacityNotifications checks if capacity notifications should be sent
// This is called asynchronously after an RSVP is created
func (h *RSVPHandler) checkCapacityNotifications(ctx context.Context, eventID string, capacity int) {
	// Get current RSVP count
	var count int
	err := h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM event_rsvps WHERE event_id = $1 AND status = 'attending'", eventID).Scan(&count)
	if err != nil {
		// Log error but don't fail
		return
	}

	percentFull := float64(count) / float64(capacity) * 100

	// Check if we should send notification
	if percentFull >= 80.0 && percentFull < 100.0 {
		// TODO: Send 80% capacity notification to hosts/admins
		// This will be implemented in Week 8 (Email Notification System)
	} else if percentFull >= 100.0 {
		// TODO: Send 100% capacity notification to hosts/admins
		// This will be implemented in Week 8 (Email Notification System)
	}
}
