package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/forgeutah/taikai/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/gosimple/slug"
)

// OrganizationHandler handles organization-related requests
type OrganizationHandler struct {
	db      *sql.DB
	checker *auth.PermissionChecker
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(db *sql.DB, checker *auth.PermissionChecker) *OrganizationHandler {
	return &OrganizationHandler{
		db:      db,
		checker: checker,
	}
}

// OrganizationResponse represents an organization in responses
type OrganizationResponse struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Slug                string     `json:"slug"`
	Description         *string    `json:"description"`
	LogoURL             *string    `json:"logo_url"`
	WebsiteURL          *string    `json:"website_url"`
	CommunityChatURL    *string    `json:"community_chat_url"`
	CommunityChatName   *string    `json:"community_chat_name"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// UpdateOrganizationRequest represents an organization update request
type UpdateOrganizationRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	LogoURL           *string `json:"logo_url,omitempty"`
	WebsiteURL        *string `json:"website_url,omitempty"`
	CommunityChatURL  *string `json:"community_chat_url,omitempty"`
	CommunityChatName *string `json:"community_chat_name,omitempty"`
}

// AdminResponse represents an admin user in responses
type AdminResponse struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  *string   `json:"avatar_url"`
	AssignedAt time.Time `json:"assigned_at"`
}

// GetOrganization returns an organization by ID
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

	var org OrganizationResponse
	query := `
		SELECT id, name, slug, description, logo_url, website_url,
		       community_chat_url, community_chat_name, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	err := h.db.QueryRowContext(r.Context(), query, orgID).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description, &org.LogoURL, &org.WebsiteURL,
		&org.CommunityChatURL, &org.CommunityChatName, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Organization not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get organization")
		return
	}

	RespondSuccess(w, http.StatusOK, org)
}

// GetOrganizationBySlug returns an organization by slug
func (h *OrganizationHandler) GetOrganizationBySlug(w http.ResponseWriter, r *http.Request) {
	slugParam := chi.URLParam(r, "slug")

	var org OrganizationResponse
	query := `
		SELECT id, name, slug, description, logo_url, website_url,
		       community_chat_url, community_chat_name, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`
	err := h.db.QueryRowContext(r.Context(), query, slugParam).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description, &org.LogoURL, &org.WebsiteURL,
		&org.CommunityChatURL, &org.CommunityChatName, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Organization not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get organization")
		return
	}

	RespondSuccess(w, http.StatusOK, org)
}

// ListOrganizations returns all organizations
func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, name, slug, description, logo_url, website_url,
		       community_chat_url, community_chat_name, created_at, updated_at
		FROM organizations
		ORDER BY created_at DESC
	`
	rows, err := h.db.QueryContext(r.Context(), query)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to list organizations")
		return
	}
	defer rows.Close()

	orgs := []OrganizationResponse{}
	for rows.Next() {
		var org OrganizationResponse
		err := rows.Scan(
			&org.ID, &org.Name, &org.Slug, &org.Description, &org.LogoURL, &org.WebsiteURL,
			&org.CommunityChatURL, &org.CommunityChatName, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan organization")
			return
		}
		orgs = append(orgs, org)
	}

	RespondSuccess(w, http.StatusOK, orgs)
}

// UpdateOrganization updates an organization (org admin only)
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Get current organization
	var current OrganizationResponse
	query := `SELECT name, description, logo_url, website_url, community_chat_url, community_chat_name FROM organizations WHERE id = $1`
	err := h.db.QueryRowContext(r.Context(), query, orgID).Scan(
		&current.Name, &current.Description, &current.LogoURL, &current.WebsiteURL,
		&current.CommunityChatURL, &current.CommunityChatName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Organization not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get organization")
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
	websiteURL := current.WebsiteURL
	if req.WebsiteURL != nil {
		websiteURL = req.WebsiteURL
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

	// Update organization
	updateQuery := `
		UPDATE organizations
		SET name = $2, slug = $3, description = $4, logo_url = $5, website_url = $6,
		    community_chat_url = $7, community_chat_name = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, slug, description, logo_url, website_url,
		          community_chat_url, community_chat_name, created_at, updated_at
	`
	var org OrganizationResponse
	err = h.db.QueryRowContext(r.Context(), updateQuery,
		orgID, name, newSlug, description, logoURL, websiteURL, communityChatURL, communityChatName,
	).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Description, &org.LogoURL, &org.WebsiteURL,
		&org.CommunityChatURL, &org.CommunityChatName, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update organization")
		return
	}

	RespondSuccess(w, http.StatusOK, org)
}

// GetOrgAdmins returns all admins for an organization
func (h *OrganizationHandler) GetOrgAdmins(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

	query := `
		SELECT u.id, u.email, u.name, u.avatar_url, oa.created_at
		FROM org_admins oa
		JOIN users u ON u.id = oa.user_id
		WHERE oa.organization_id = $1 AND u.is_active = true
		ORDER BY oa.created_at ASC
	`
	rows, err := h.db.QueryContext(r.Context(), query, orgID)
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

// AddOrgAdmin adds a user as an organization admin
func (h *OrganizationHandler) AddOrgAdmin(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

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
		"INSERT INTO org_admins (user_id, organization_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		req.UserID, orgID,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to add admin")
		return
	}

	RespondSuccess(w, http.StatusCreated, map[string]string{"message": "Admin added successfully"})
}

// RemoveOrgAdmin removes a user as an organization admin
func (h *OrganizationHandler) RemoveOrgAdmin(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")
	userID := chi.URLParam(r, "userId")

	result, err := h.db.ExecContext(r.Context(),
		"DELETE FROM org_admins WHERE user_id = $1 AND organization_id = $2",
		userID, orgID,
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
