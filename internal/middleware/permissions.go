package middleware

import (
	"net/http"

	"github.com/forgeutah/taikai/internal/auth"
	"github.com/go-chi/chi/v5"
)

// PermissionMiddleware provides permission checking middleware
type PermissionMiddleware struct {
	checker *auth.PermissionChecker
}

// NewPermissionMiddleware creates a new permission middleware
func NewPermissionMiddleware(checker *auth.PermissionChecker) *PermissionMiddleware {
	return &PermissionMiddleware{
		checker: checker,
	}
}

// RequireOrgAdmin middleware ensures user is an org admin
func (m *PermissionMiddleware) RequireOrgAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIDFromContext(r.Context())
		if userID == "" {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		orgID := chi.URLParam(r, "orgId")
		if orgID == "" {
			respondError(w, http.StatusBadRequest, "organization ID required")
			return
		}

		isAdmin, err := m.checker.IsOrgAdmin(r.Context(), userID, orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "permission check failed")
			return
		}

		if !isAdmin {
			respondError(w, http.StatusForbidden, "organization admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireGroupAdmin middleware ensures user is a group admin or org admin
func (m *PermissionMiddleware) RequireGroupAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIDFromContext(r.Context())
		if userID == "" {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		groupID := chi.URLParam(r, "groupId")
		if groupID == "" {
			respondError(w, http.StatusBadRequest, "group ID required")
			return
		}

		canManage, err := m.checker.CanManageGroup(r.Context(), userID, groupID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "permission check failed")
			return
		}

		if !canManage {
			respondError(w, http.StatusForbidden, "group admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireEventManagement middleware ensures user can manage an event
func (m *PermissionMiddleware) RequireEventManagement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIDFromContext(r.Context())
		if userID == "" {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		eventID := chi.URLParam(r, "eventId")
		if eventID == "" {
			respondError(w, http.StatusBadRequest, "event ID required")
			return
		}

		canManage, err := m.checker.CanManageEvent(r.Context(), userID, eventID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "permission check failed")
			return
		}

		if !canManage {
			respondError(w, http.StatusForbidden, "event management access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}
