package auth

import (
	"context"
	"database/sql"
)

// PermissionChecker handles permission checks
type PermissionChecker struct {
	db *sql.DB
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(db *sql.DB) *PermissionChecker {
	return &PermissionChecker{db: db}
}

// IsOrgAdmin checks if a user is an organization admin
func (p *PermissionChecker) IsOrgAdmin(ctx context.Context, userID, orgID string) (bool, error) {
	var isAdmin bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM org_admins
			WHERE user_id = $1 AND organization_id = $2
		)
	`
	err := p.db.QueryRowContext(ctx, query, userID, orgID).Scan(&isAdmin)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}

// IsGroupAdmin checks if a user is a group admin
func (p *PermissionChecker) IsGroupAdmin(ctx context.Context, userID, groupID string) (bool, error) {
	var isAdmin bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM group_admins
			WHERE user_id = $1 AND group_id = $2
		)
	`
	err := p.db.QueryRowContext(ctx, query, userID, groupID).Scan(&isAdmin)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}

// IsEventHost checks if a user is an event host
func (p *PermissionChecker) IsEventHost(ctx context.Context, userID, eventID string) (bool, error) {
	var isHost bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM event_hosts
			WHERE user_id = $1 AND event_id = $2
		)
	`
	err := p.db.QueryRowContext(ctx, query, userID, eventID).Scan(&isHost)
	if err != nil {
		return false, err
	}
	return isHost, nil
}

// CanManageGroup checks if a user can manage a group (is group admin OR org admin)
func (p *PermissionChecker) CanManageGroup(ctx context.Context, userID, groupID string) (bool, error) {
	// First check if group admin
	isGroupAdmin, err := p.IsGroupAdmin(ctx, userID, groupID)
	if err != nil {
		return false, err
	}
	if isGroupAdmin {
		return true, nil
	}

	// Check if org admin of the group's organization
	var orgID string
	err = p.db.QueryRowContext(ctx, "SELECT organization_id FROM groups WHERE id = $1", groupID).Scan(&orgID)
	if err != nil {
		return false, err
	}

	return p.IsOrgAdmin(ctx, userID, orgID)
}

// CanManageEvent checks if a user can manage an event (is event host OR group admin OR org admin)
func (p *PermissionChecker) CanManageEvent(ctx context.Context, userID, eventID string) (bool, error) {
	// First check if event host
	isHost, err := p.IsEventHost(ctx, userID, eventID)
	if err != nil {
		return false, err
	}
	if isHost {
		return true, nil
	}

	// Get event's group ID
	var groupID string
	err = p.db.QueryRowContext(ctx, "SELECT group_id FROM events WHERE id = $1", eventID).Scan(&groupID)
	if err != nil {
		return false, err
	}

	// Check if can manage the group (which checks both group admin and org admin)
	return p.CanManageGroup(ctx, userID, groupID)
}

// GetOrgIDForGroup returns the organization ID for a group
func (p *PermissionChecker) GetOrgIDForGroup(ctx context.Context, groupID string) (string, error) {
	var orgID string
	err := p.db.QueryRowContext(ctx, "SELECT organization_id FROM groups WHERE id = $1", groupID).Scan(&orgID)
	return orgID, err
}

// GetGroupIDForEvent returns the group ID for an event
func (p *PermissionChecker) GetGroupIDForEvent(ctx context.Context, eventID string) (string, error) {
	var groupID string
	err := p.db.QueryRowContext(ctx, "SELECT group_id FROM events WHERE id = $1", eventID).Scan(&groupID)
	return groupID, err
}
