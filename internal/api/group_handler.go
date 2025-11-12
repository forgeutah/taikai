package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/forgeutah/taikai/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/gosimple/slug"
	"github.com/google/uuid"
)

// GroupHandler handles group-related requests
type GroupHandler struct {
	db      *sql.DB
	checker *auth.PermissionChecker
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(db *sql.DB, checker *auth.PermissionChecker) *GroupHandler {
	return &GroupHandler{
		db:      db,
		checker: checker,
	}
}

// GroupResponse represents a group in responses
type GroupResponse struct {
	ID                  string     `json:"id"`
	OrganizationID      string     `json:"organization_id"`
	Name                string     `json:"name"`
	Slug                string     `json:"slug"`
	Description         *string    `json:"description"`
	LogoURL             *string    `json:"logo_url"`
	PrimaryLocation     *string    `json:"primary_location"`
	CommunityChatURL    *string    `json:"community_chat_url"`
	CommunityChatName   *string    `json:"community_chat_name"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// CreateGroupRequest represents a group creation request
type CreateGroupRequest struct {
	OrganizationID    string  `json:"organization_id"`
	Name              string  `json:"name"`
	Description       *string `json:"description,omitempty"`
	LogoURL           *string `json:"logo_url,omitempty"`
	PrimaryLocation   *string `json:"primary_location,omitempty"`
	CommunityChatURL  *string `json:"community_chat_url,omitempty"`
	CommunityChatName *string `json:"community_chat_name,omitempty"`
}

// UpdateGroupRequest represents a group update request
type UpdateGroupRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	LogoURL           *string `json:"logo_url,omitempty"`
	PrimaryLocation   *string `json:"primary_location,omitempty"`
	CommunityChatURL  *string `json:"community_chat_url,omitempty"`
	CommunityChatName *string `json:"community_chat_name,omitempty"`
}

// ListGroupsResponse represents a paginated list of groups
type ListGroupsResponse struct {
	Groups []GroupResponse `json:"groups"`
	Total  int64           `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// GetGroup returns a group by ID
func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	var group GroupResponse
	query := `
		SELECT id, organization_id, name, slug, description, logo_url, primary_location,
		       community_chat_url, community_chat_name, created_at, updated_at
		FROM groups
		WHERE id = $1
	`
	err := h.db.QueryRowContext(r.Context(), query, groupID).Scan(
		&group.ID, &group.OrganizationID, &group.Name, &group.Slug, &group.Description,
		&group.LogoURL, &group.PrimaryLocation, &group.CommunityChatURL, &group.CommunityChatName,
		&group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Group not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get group")
		return
	}

	RespondSuccess(w, http.StatusOK, group)
}

// GetGroupBySlug returns a group by slug
func (h *GroupHandler) GetGroupBySlug(w http.ResponseWriter, r *http.Request) {
	slugParam := chi.URLParam(r, "slug")

	var group GroupResponse
	query := `
		SELECT id, organization_id, name, slug, description, logo_url, primary_location,
		       community_chat_url, community_chat_name, created_at, updated_at
		FROM groups
		WHERE slug = $1
	`
	err := h.db.QueryRowContext(r.Context(), query, slugParam).Scan(
		&group.ID, &group.OrganizationID, &group.Name, &group.Slug, &group.Description,
		&group.LogoURL, &group.PrimaryLocation, &group.CommunityChatURL, &group.CommunityChatName,
		&group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Group not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get group")
		return
	}

	RespondSuccess(w, http.StatusOK, group)
}

// ListGroups returns all groups with optional filtering and pagination
func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Parse filter params
	orgID := r.URL.Query().Get("organization_id")

	var query string
	var countQuery string
	var args []interface{}
	var total int64

	if orgID != "" {
		// Filter by organization
		query = `
			SELECT id, organization_id, name, slug, description, logo_url, primary_location,
			       community_chat_url, community_chat_name, created_at, updated_at
			FROM groups
			WHERE organization_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		countQuery = "SELECT COUNT(*) FROM groups WHERE organization_id = $1"
		args = []interface{}{orgID, limit, offset}
		h.db.QueryRowContext(r.Context(), countQuery, orgID).Scan(&total)
	} else {
		// All groups
		query = `
			SELECT id, organization_id, name, slug, description, logo_url, primary_location,
			       community_chat_url, community_chat_name, created_at, updated_at
			FROM groups
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		countQuery = "SELECT COUNT(*) FROM groups"
		args = []interface{}{limit, offset}
		h.db.QueryRowContext(r.Context(), countQuery).Scan(&total)
	}

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to list groups")
		return
	}
	defer rows.Close()

	groups := []GroupResponse{}
	for rows.Next() {
		var group GroupResponse
		err := rows.Scan(
			&group.ID, &group.OrganizationID, &group.Name, &group.Slug, &group.Description,
			&group.LogoURL, &group.PrimaryLocation, &group.CommunityChatURL, &group.CommunityChatName,
			&group.CreatedAt, &group.UpdatedAt,
		)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan group")
			return
		}
		groups = append(groups, group)
	}

	response := ListGroupsResponse{
		Groups: groups,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}

	RespondSuccess(w, http.StatusOK, response)
}

// CreateGroup creates a new group (org admin only)
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, "name is required")
		return
	}
	if req.OrganizationID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, "organization_id is required")
		return
	}

	// Check if user is org admin
	isOrgAdmin, err := h.checker.IsOrgAdmin(r.Context(), userID, req.OrganizationID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Permission check failed")
		return
	}
	if !isOrgAdmin {
		RespondError(w, http.StatusForbidden, ErrCodeForbidden, "Organization admin access required")
		return
	}

	// Get organization details for default community chat
	var orgChatURL, orgChatName sql.NullString
	err = h.db.QueryRowContext(r.Context(),
		"SELECT community_chat_url, community_chat_name FROM organizations WHERE id = $1",
		req.OrganizationID,
	).Scan(&orgChatURL, &orgChatName)
	if err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, "Organization not found")
		return
	}

	// Auto-inherit org community chat if not specified
	chatURL := req.CommunityChatURL
	chatName := req.CommunityChatName
	if chatURL == nil && orgChatURL.Valid {
		chatURL = &orgChatURL.String
	}
	if chatName == nil && orgChatName.Valid {
		chatName = &orgChatName.String
	}

	// Generate slug
	groupSlug := slug.Make(req.Name)
	groupID := uuid.New().String()

	// Create group
	query := `
		INSERT INTO groups (
			id, organization_id, name, slug, description, logo_url, primary_location,
			community_chat_url, community_chat_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, organization_id, name, slug, description, logo_url, primary_location,
		          community_chat_url, community_chat_name, created_at, updated_at
	`
	var group GroupResponse
	err = h.db.QueryRowContext(r.Context(), query,
		groupID, req.OrganizationID, req.Name, groupSlug, req.Description, req.LogoURL,
		req.PrimaryLocation, chatURL, chatName,
	).Scan(
		&group.ID, &group.OrganizationID, &group.Name, &group.Slug, &group.Description,
		&group.LogoURL, &group.PrimaryLocation, &group.CommunityChatURL, &group.CommunityChatName,
		&group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create group")
		return
	}

	// Auto-assign creator as group admin
	_, err = h.db.ExecContext(r.Context(),
		"INSERT INTO group_admins (user_id, group_id) VALUES ($1, $2)",
		userID, group.ID,
	)
	if err != nil {
		// Log but don't fail
		// In production, log this error properly
	}

	RespondSuccess(w, http.StatusCreated, group)
}

// UpdateGroup updates a group (group admin or org admin)
func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Get current group
	var current GroupResponse
	query := `
		SELECT name, description, logo_url, primary_location, community_chat_url, community_chat_name
		FROM groups WHERE id = $1
	`
	err := h.db.QueryRowContext(r.Context(), query, groupID).Scan(
		&current.Name, &current.Description, &current.LogoURL, &current.PrimaryLocation,
		&current.CommunityChatURL, &current.CommunityChatName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Group not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get group")
		return
	}

	// Apply updates
	name := current.Name
	if req.Name != nil {
		name = *req.Name
	}
	description := current.Description
	if req.Description != nil {
		description = req.Description
	}
	logoURL := current.LogoURL
	if req.LogoURL != nil {
		logoURL = req.LogoURL
	}
	primaryLocation := current.PrimaryLocation
	if req.PrimaryLocation != nil {
		primaryLocation = req.PrimaryLocation
	}
	communityChatURL := current.CommunityChatURL
	if req.CommunityChatURL != nil {
		communityChatURL = req.CommunityChatURL
	}
	communityChatName := current.CommunityChatName
	if req.CommunityChatName != nil {
		communityChatName = req.CommunityChatName
	}

	// Generate new slug if name changed
	newSlug := slug.Make(name)

	// Update group
	updateQuery := `
		UPDATE groups
		SET name = $2, slug = $3, description = $4, logo_url = $5, primary_location = $6,
		    community_chat_url = $7, community_chat_name = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING id, organization_id, name, slug, description, logo_url, primary_location,
		          community_chat_url, community_chat_name, created_at, updated_at
	`
	var group GroupResponse
	err = h.db.QueryRowContext(r.Context(), updateQuery,
		groupID, name, newSlug, description, logoURL, primaryLocation, communityChatURL, communityChatName,
	).Scan(
		&group.ID, &group.OrganizationID, &group.Name, &group.Slug, &group.Description,
		&group.LogoURL, &group.PrimaryLocation, &group.CommunityChatURL, &group.CommunityChatName,
		&group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update group")
		return
	}

	RespondSuccess(w, http.StatusOK, group)
}

// DeleteGroup deletes a group (org admin only)
func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")
	userID := auth.GetUserIDFromContext(r.Context())

	// Get org ID for permission check
	orgID, err := h.checker.GetOrgIDForGroup(r.Context(), groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Group not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get group")
		return
	}

	// Only org admins can delete groups
	isOrgAdmin, err := h.checker.IsOrgAdmin(r.Context(), userID, orgID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Permission check failed")
		return
	}
	if !isOrgAdmin {
		RespondError(w, http.StatusForbidden, ErrCodeForbidden, "Organization admin access required")
		return
	}

	// Delete group (cascades to group_admins, events, etc.)
	_, err = h.db.ExecContext(r.Context(), "DELETE FROM groups WHERE id = $1", groupID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to delete group")
		return
	}

	RespondSuccess(w, http.StatusOK, map[string]string{"message": "Group deleted successfully"})
}

// GetGroupAdmins returns all admins for a group
func (h *GroupHandler) GetGroupAdmins(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	query := `
		SELECT u.id, u.email, u.name, u.avatar_url, ga.created_at
		FROM group_admins ga
		JOIN users u ON u.id = ga.user_id
		WHERE ga.group_id = $1 AND u.is_active = true
		ORDER BY ga.created_at ASC
	`
	rows, err := h.db.QueryContext(r.Context(), query, groupID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get admins")
		return
	}
	defer rows.Close()

	admins := []AdminResponse{}
	for rows.Next() {
		var admin AdminResponse
		err := rows.Scan(&admin.ID, &admin.Email, &admin.Name, &admin.AvatarURL, &admin.AssignedAt)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan admin")
			return
		}
		admins = append(admins, admin)
	}

	RespondSuccess(w, http.StatusOK, admins)
}

// AddGroupAdmin adds a user as a group admin
func (h *GroupHandler) AddGroupAdmin(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, "user_id is required")
		return
	}

	// Verify user exists and is active
	var exists bool
	err := h.db.QueryRowContext(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND is_active = true)",
		req.UserID,
	).Scan(&exists)
	if err != nil || !exists {
		RespondError(w, http.StatusBadRequest, ErrCodeValidation, "User not found or inactive")
		return
	}

	// Add as admin
	_, err = h.db.ExecContext(r.Context(),
		"INSERT INTO group_admins (user_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		req.UserID, groupID,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to add admin")
		return
	}

	RespondSuccess(w, http.StatusCreated, map[string]string{"message": "Admin added successfully"})
}

// RemoveGroupAdmin removes a user as a group admin
func (h *GroupHandler) RemoveGroupAdmin(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")
	userID := chi.URLParam(r, "userId")

	result, err := h.db.ExecContext(r.Context(),
		"DELETE FROM group_admins WHERE user_id = $1 AND group_id = $2",
		userID, groupID,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to remove admin")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Admin assignment not found")
		return
	}

	RespondSuccess(w, http.StatusOK, map[string]string{"message": "Admin removed successfully"})
}
