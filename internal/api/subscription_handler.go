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

// SubscriptionHandler handles subscription-related requests
type SubscriptionHandler struct {
	db      *sql.DB
	checker *auth.PermissionChecker
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(db *sql.DB, checker *auth.PermissionChecker) *SubscriptionHandler {
	return &SubscriptionHandler{
		db:      db,
		checker: checker,
	}
}

// ===============================
// Request/Response Types
// ===============================

// SubscriptionRequest for creating/updating subscriptions
type SubscriptionRequest struct {
	NotifyEmail   *bool `json:"notify_email"`
	NotifySMS     *bool `json:"notify_sms"`
	NotifyDiscord *bool `json:"notify_discord"`
}

// OrgSubscriptionResponse represents an organization subscription
type OrgSubscriptionResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	NotifyEmail    bool   `json:"notify_email"`
	NotifySMS      bool   `json:"notify_sms"`
	NotifyDiscord  bool   `json:"notify_discord"`
	CreatedAt      string `json:"created_at"`
}

// GroupSubscriptionResponse represents a group subscription
type GroupSubscriptionResponse struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	GroupID       string `json:"group_id"`
	NotifyEmail   bool   `json:"notify_email"`
	NotifySMS     bool   `json:"notify_sms"`
	NotifyDiscord bool   `json:"notify_discord"`
	CreatedAt     string `json:"created_at"`
}

// SubscriberResponse represents a subscriber with user details
type SubscriberResponse struct {
	ID            string  `json:"id"`
	UserID        string  `json:"user_id"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	NotifyEmail   bool    `json:"notify_email"`
	NotifySMS     bool    `json:"notify_sms"`
	NotifyDiscord bool    `json:"notify_discord"`
	CreatedAt     string  `json:"created_at"`
}

// UserSubscriptionsResponse represents all user's subscriptions
type UserSubscriptionsResponse struct {
	OrgSubscriptions   []OrgSubscriptionWithDetails   `json:"org_subscriptions"`
	GroupSubscriptions []GroupSubscriptionWithDetails `json:"group_subscriptions"`
}

// OrgSubscriptionWithDetails includes organization details
type OrgSubscriptionWithDetails struct {
	ID                 string  `json:"id"`
	UserID             string  `json:"user_id"`
	OrganizationID     string  `json:"organization_id"`
	OrganizationName   string  `json:"organization_name"`
	OrganizationSlug   string  `json:"organization_slug"`
	OrganizationLogoURL *string `json:"organization_logo_url,omitempty"`
	NotifyEmail        bool    `json:"notify_email"`
	NotifySMS          bool    `json:"notify_sms"`
	NotifyDiscord      bool    `json:"notify_discord"`
	CreatedAt          string  `json:"created_at"`
}

// GroupSubscriptionWithDetails includes group details
type GroupSubscriptionWithDetails struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	GroupID        string `json:"group_id"`
	GroupName      string `json:"group_name"`
	GroupSlug      string `json:"group_slug"`
	OrganizationID string `json:"organization_id"`
	NotifyEmail    bool   `json:"notify_email"`
	NotifySMS      bool   `json:"notify_sms"`
	NotifyDiscord  bool   `json:"notify_discord"`
	CreatedAt      string `json:"created_at"`
}

// ===============================
// Organization Subscription Endpoints
// ===============================

// SubscribeToOrganization handles POST /api/v1/organizations/:orgId/subscribe
func (h *SubscriptionHandler) SubscribeToOrganization(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "organization ID is required")
		return
	}

	if _, err := uuid.Parse(orgID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid organization ID format")
		return
	}

	// Verify organization exists
	var exists bool
	err := h.db.QueryRowContext(r.Context(), "SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)", orgID).Scan(&exists)
	if err != nil || !exists {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Organization not found")
		return
	}

	// Parse request body (optional notification preferences)
	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Set defaults if not provided
	notifyEmail := true
	notifySMS := false
	notifyDiscord := false

	if req.NotifyEmail != nil {
		notifyEmail = *req.NotifyEmail
	}
	if req.NotifySMS != nil {
		notifySMS = *req.NotifySMS
	}
	if req.NotifyDiscord != nil {
		notifyDiscord = *req.NotifyDiscord
	}

	// Create or update subscription
	query := `
		INSERT INTO org_subscriptions (user_id, organization_id, notify_email, notify_sms, notify_discord)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, organization_id)
		DO UPDATE SET
			notify_email = EXCLUDED.notify_email,
			notify_sms = EXCLUDED.notify_sms,
			notify_discord = EXCLUDED.notify_discord
		RETURNING id, user_id, organization_id, notify_email, notify_sms, notify_discord, created_at
	`

	var sub OrgSubscriptionResponse
	err = h.db.QueryRowContext(r.Context(), query,
		userID, orgID, notifyEmail, notifySMS, notifyDiscord,
	).Scan(&sub.ID, &sub.UserID, &sub.OrganizationID,
		&sub.NotifyEmail, &sub.NotifySMS, &sub.NotifyDiscord, &sub.CreatedAt)

	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create subscription")
		return
	}

	RespondSuccess(w, http.StatusOK, sub)
}

// UnsubscribeFromOrganization handles DELETE /api/v1/organizations/:orgId/subscribe
func (h *SubscriptionHandler) UnsubscribeFromOrganization(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "organization ID is required")
		return
	}

	if _, err := uuid.Parse(orgID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid organization ID format")
		return
	}

	_, err := h.db.ExecContext(r.Context(),
		"DELETE FROM org_subscriptions WHERE user_id = $1 AND organization_id = $2",
		userID, orgID,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateOrgSubscription handles PATCH /api/v1/organizations/:orgId/subscribe
func (h *SubscriptionHandler) UpdateOrgSubscription(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "organization ID is required")
		return
	}

	if _, err := uuid.Parse(orgID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid organization ID format")
		return
	}

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Update subscription
	query := `
		UPDATE org_subscriptions
		SET
			notify_email = COALESCE($3, notify_email),
			notify_sms = COALESCE($4, notify_sms),
			notify_discord = COALESCE($5, notify_discord)
		WHERE user_id = $1 AND organization_id = $2
		RETURNING id, user_id, organization_id, notify_email, notify_sms, notify_discord, created_at
	`

	var sub OrgSubscriptionResponse
	err := h.db.QueryRowContext(r.Context(), query,
		userID, orgID, req.NotifyEmail, req.NotifySMS, req.NotifyDiscord,
	).Scan(&sub.ID, &sub.UserID, &sub.OrganizationID,
		&sub.NotifyEmail, &sub.NotifySMS, &sub.NotifyDiscord, &sub.CreatedAt)

	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Subscription not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update subscription")
		return
	}

	RespondSuccess(w, http.StatusOK, sub)
}

// GetOrgSubscribers handles GET /api/v1/organizations/:orgId/subscribers
func (h *SubscriptionHandler) GetOrgSubscribers(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	orgID := chi.URLParam(r, "orgId")
	if orgID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "organization ID is required")
		return
	}

	if _, err := uuid.Parse(orgID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid organization ID format")
		return
	}

	// Check if user is org admin
	isAdmin, err := h.checker.IsOrgAdmin(r.Context(), userID, orgID)
	if err != nil || !isAdmin {
		RespondError(w, http.StatusForbidden, ErrCodeForbidden, "Only organization admins can view subscribers")
		return
	}

	// Parse pagination
	page, limit := parsePagination(r)
	offset := (page - 1) * limit

	// Get subscribers
	query := `
		SELECT
			os.id, os.user_id, os.organization_id,
			os.notify_email, os.notify_sms, os.notify_discord, os.created_at,
			u.email, u.name, u.avatar_url
		FROM org_subscriptions os
		JOIN users u ON os.user_id = u.id
		WHERE os.organization_id = $1
		ORDER BY os.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.QueryContext(r.Context(), query, orgID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to fetch subscribers")
		return
	}
	defer rows.Close()

	subscribers := []SubscriberResponse{}
	for rows.Next() {
		var sub SubscriberResponse
		var avatarURL sql.NullString
		var orgID string

		err := rows.Scan(&sub.ID, &sub.UserID, &orgID,
			&sub.NotifyEmail, &sub.NotifySMS, &sub.NotifyDiscord, &sub.CreatedAt,
			&sub.Email, &sub.Name, &avatarURL)
		if err != nil {
			continue
		}

		if avatarURL.Valid {
			sub.AvatarURL = &avatarURL.String
		}

		subscribers = append(subscribers, sub)
	}

	// Get total count
	var total int
	err = h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM org_subscriptions WHERE organization_id = $1",
		orgID,
	).Scan(&total)
	if err != nil {
		total = len(subscribers)
	}

	RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"subscribers": subscribers,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ===============================
// Group Subscription Endpoints
// ===============================

// SubscribeToGroup handles POST /api/v1/groups/:groupId/subscribe
func (h *SubscriptionHandler) SubscribeToGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	groupID := chi.URLParam(r, "groupId")
	if groupID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "group ID is required")
		return
	}

	if _, err := uuid.Parse(groupID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid group ID format")
		return
	}

	// Get group's organization ID
	var orgID string
	err := h.db.QueryRowContext(r.Context(),
		"SELECT organization_id FROM groups WHERE id = $1",
		groupID,
	).Scan(&orgID)
	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Group not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to fetch group")
		return
	}

	// Check if user already has org subscription
	var hasOrgSub bool
	err = h.db.QueryRowContext(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM org_subscriptions WHERE user_id = $1 AND organization_id = $2)",
		userID, orgID,
	).Scan(&hasOrgSub)
	if err != nil {
		hasOrgSub = false
	}

	if hasOrgSub {
		RespondError(w, http.StatusConflict, ErrCodeConflict, "You are already subscribed to this organization, which includes all groups")
		return
	}

	// Parse request body
	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Set defaults
	notifyEmail := true
	notifySMS := false
	notifyDiscord := false

	if req.NotifyEmail != nil {
		notifyEmail = *req.NotifyEmail
	}
	if req.NotifySMS != nil {
		notifySMS = *req.NotifySMS
	}
	if req.NotifyDiscord != nil {
		notifyDiscord = *req.NotifyDiscord
	}

	// Create or update subscription
	query := `
		INSERT INTO group_subscriptions (user_id, group_id, notify_email, notify_sms, notify_discord)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, group_id)
		DO UPDATE SET
			notify_email = EXCLUDED.notify_email,
			notify_sms = EXCLUDED.notify_sms,
			notify_discord = EXCLUDED.notify_discord
		RETURNING id, user_id, group_id, notify_email, notify_sms, notify_discord, created_at
	`

	var sub GroupSubscriptionResponse
	err = h.db.QueryRowContext(r.Context(), query,
		userID, groupID, notifyEmail, notifySMS, notifyDiscord,
	).Scan(&sub.ID, &sub.UserID, &sub.GroupID,
		&sub.NotifyEmail, &sub.NotifySMS, &sub.NotifyDiscord, &sub.CreatedAt)

	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create subscription")
		return
	}

	RespondSuccess(w, http.StatusOK, sub)
}

// UnsubscribeFromGroup handles DELETE /api/v1/groups/:groupId/subscribe
func (h *SubscriptionHandler) UnsubscribeFromGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	groupID := chi.URLParam(r, "groupId")
	if groupID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "group ID is required")
		return
	}

	if _, err := uuid.Parse(groupID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid group ID format")
		return
	}

	_, err := h.db.ExecContext(r.Context(),
		"DELETE FROM group_subscriptions WHERE user_id = $1 AND group_id = $2",
		userID, groupID,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateGroupSubscription handles PATCH /api/v1/groups/:groupId/subscribe
func (h *SubscriptionHandler) UpdateGroupSubscription(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	groupID := chi.URLParam(r, "groupId")
	if groupID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "group ID is required")
		return
	}

	if _, err := uuid.Parse(groupID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid group ID format")
		return
	}

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Update subscription
	query := `
		UPDATE group_subscriptions
		SET
			notify_email = COALESCE($3, notify_email),
			notify_sms = COALESCE($4, notify_sms),
			notify_discord = COALESCE($5, notify_discord)
		WHERE user_id = $1 AND group_id = $2
		RETURNING id, user_id, group_id, notify_email, notify_sms, notify_discord, created_at
	`

	var sub GroupSubscriptionResponse
	err := h.db.QueryRowContext(r.Context(), query,
		userID, groupID, req.NotifyEmail, req.NotifySMS, req.NotifyDiscord,
	).Scan(&sub.ID, &sub.UserID, &sub.GroupID,
		&sub.NotifyEmail, &sub.NotifySMS, &sub.NotifyDiscord, &sub.CreatedAt)

	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Subscription not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update subscription")
		return
	}

	RespondSuccess(w, http.StatusOK, sub)
}

// GetGroupSubscribers handles GET /api/v1/groups/:groupId/subscribers
func (h *SubscriptionHandler) GetGroupSubscribers(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	groupID := chi.URLParam(r, "groupId")
	if groupID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "group ID is required")
		return
	}

	if _, err := uuid.Parse(groupID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid group ID format")
		return
	}

	// Check if user is group admin or org admin
	canView, err := h.checker.CanManageGroup(r.Context(), userID, groupID)
	if err != nil || !canView {
		RespondError(w, http.StatusForbidden, ErrCodeForbidden, "Only group or organization admins can view subscribers")
		return
	}

	// Parse pagination
	page, limit := parsePagination(r)
	offset := (page - 1) * limit

	// Get subscribers
	query := `
		SELECT
			gs.id, gs.user_id, gs.group_id,
			gs.notify_email, gs.notify_sms, gs.notify_discord, gs.created_at,
			u.email, u.name, u.avatar_url
		FROM group_subscriptions gs
		JOIN users u ON gs.user_id = u.id
		WHERE gs.group_id = $1
		ORDER BY gs.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.QueryContext(r.Context(), query, groupID, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to fetch subscribers")
		return
	}
	defer rows.Close()

	subscribers := []SubscriberResponse{}
	for rows.Next() {
		var sub SubscriberResponse
		var avatarURL sql.NullString
		var gID string

		err := rows.Scan(&sub.ID, &sub.UserID, &gID,
			&sub.NotifyEmail, &sub.NotifySMS, &sub.NotifyDiscord, &sub.CreatedAt,
			&sub.Email, &sub.Name, &avatarURL)
		if err != nil {
			continue
		}

		if avatarURL.Valid {
			sub.AvatarURL = &avatarURL.String
		}

		subscribers = append(subscribers, sub)
	}

	// Get total count
	var total int
	err = h.db.QueryRowContext(r.Context(),
		"SELECT COUNT(*) FROM group_subscriptions WHERE group_id = $1",
		groupID,
	).Scan(&total)
	if err != nil {
		total = len(subscribers)
	}

	RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"subscribers": subscribers,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ===============================
// User Subscription Management
// ===============================

// GetMySubscriptions handles GET /api/v1/me/subscriptions
func (h *SubscriptionHandler) GetMySubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		RespondError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	// Get organization subscriptions
	orgQuery := `
		SELECT
			os.id, os.user_id, os.organization_id,
			os.notify_email, os.notify_sms, os.notify_discord, os.created_at,
			o.name, o.slug, o.logo_url
		FROM org_subscriptions os
		JOIN organizations o ON os.organization_id = o.id
		WHERE os.user_id = $1
		ORDER BY o.name ASC
	`

	orgRows, err := h.db.QueryContext(r.Context(), orgQuery, userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to fetch subscriptions")
		return
	}
	defer orgRows.Close()

	orgSubs := []OrgSubscriptionWithDetails{}
	for orgRows.Next() {
		var sub OrgSubscriptionWithDetails
		var logoURL sql.NullString

		err := orgRows.Scan(&sub.ID, &sub.UserID, &sub.OrganizationID,
			&sub.NotifyEmail, &sub.NotifySMS, &sub.NotifyDiscord, &sub.CreatedAt,
			&sub.OrganizationName, &sub.OrganizationSlug, &logoURL)
		if err != nil {
			continue
		}

		if logoURL.Valid {
			sub.OrganizationLogoURL = &logoURL.String
		}

		orgSubs = append(orgSubs, sub)
	}

	// Get group subscriptions
	groupQuery := `
		SELECT
			gs.id, gs.user_id, gs.group_id,
			gs.notify_email, gs.notify_sms, gs.notify_discord, gs.created_at,
			g.name, g.slug, g.organization_id
		FROM group_subscriptions gs
		JOIN groups g ON gs.group_id = g.id
		WHERE gs.user_id = $1
		ORDER BY g.name ASC
	`

	groupRows, err := h.db.QueryContext(r.Context(), groupQuery, userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to fetch subscriptions")
		return
	}
	defer groupRows.Close()

	groupSubs := []GroupSubscriptionWithDetails{}
	for groupRows.Next() {
		var sub GroupSubscriptionWithDetails

		err := groupRows.Scan(&sub.ID, &sub.UserID, &sub.GroupID,
			&sub.NotifyEmail, &sub.NotifySMS, &sub.NotifyDiscord, &sub.CreatedAt,
			&sub.GroupName, &sub.GroupSlug, &sub.OrganizationID)
		if err != nil {
			continue
		}

		groupSubs = append(groupSubs, sub)
	}

	response := UserSubscriptionsResponse{
		OrgSubscriptions:   orgSubs,
		GroupSubscriptions: groupSubs,
	}

	RespondSuccess(w, http.StatusOK, response)
}

// ===============================
// Helper Functions
// ===============================

// parsePagination extracts page and limit from request
func parsePagination(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return
}

// AutoSubscribeHost automatically subscribes an event host to the group (if not already subscribed)
func (h *SubscriptionHandler) AutoSubscribeHost(ctx context.Context, userID, groupID string) error {
	// Get group's organization ID
	var orgID string
	err := h.db.QueryRowContext(ctx,
		"SELECT organization_id FROM groups WHERE id = $1",
		groupID,
	).Scan(&orgID)
	if err != nil {
		return err
	}

	// Check if user has org subscription
	var hasOrgSub bool
	err = h.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM org_subscriptions WHERE user_id = $1 AND organization_id = $2)",
		userID, orgID,
	).Scan(&hasOrgSub)
	if err != nil {
		hasOrgSub = false
	}

	// If user has org subscription, don't create group subscription
	if hasOrgSub {
		return nil
	}

	// Check if user has group subscription
	var hasGroupSub bool
	err = h.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM group_subscriptions WHERE user_id = $1 AND group_id = $2)",
		userID, groupID,
	).Scan(&hasGroupSub)
	if err != nil {
		hasGroupSub = false
	}

	// If user already has group subscription, don't create
	if hasGroupSub {
		return nil
	}

	// Create group subscription with default preferences
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO group_subscriptions (user_id, group_id, notify_email, notify_sms, notify_discord)
		VALUES ($1, $2, true, false, false)
		ON CONFLICT (user_id, group_id) DO NOTHING
	`, userID, groupID)

	return err
}
