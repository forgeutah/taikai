package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/forgeutah/taikai/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSubscribeToOrganization tests organization subscription creation
func TestSubscribeToOrganization(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &SubscriptionHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	orgID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/"+orgID+"/subscribe", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"orgId": orgID})

	w := httptest.NewRecorder()

	// Mock organization exists check
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM organizations WHERE id = \\$1\\)").
		WithArgs(orgID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock subscription creation
	subRow := sqlmock.NewRows([]string{
		"id", "user_id", "organization_id", "notify_email", "notify_sms", "notify_discord", "created_at",
	}).AddRow(
		"sub-id", userID, orgID, true, false, false, "2025-01-01T00:00:00Z",
	)
	mock.ExpectQuery("INSERT INTO org_subscriptions").
		WithArgs(userID, orgID, true, false, false).
		WillReturnRows(subRow)

	handler.SubscribeToOrganization(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData)
	assert.Equal(t, userID, data["user_id"])
	assert.Equal(t, orgID, data["organization_id"])
	assert.True(t, data["notify_email"].(bool))

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestSubscribeToGroup tests group subscription creation
func TestSubscribeToGroup(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &SubscriptionHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	groupID := "550e8400-e29b-41d4-a716-446655440002"
	orgID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+groupID+"/subscribe", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"groupId": groupID})

	w := httptest.NewRecorder()

	// Mock getting group's organization
	mock.ExpectQuery("SELECT organization_id FROM groups WHERE id = \\$1").
		WithArgs(groupID).
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).AddRow(orgID))

	// Mock checking for existing org subscription
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM org_subscriptions").
		WithArgs(userID, orgID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock subscription creation
	subRow := sqlmock.NewRows([]string{
		"id", "user_id", "group_id", "notify_email", "notify_sms", "notify_discord", "created_at",
	}).AddRow(
		"sub-id", userID, groupID, true, false, false, "2025-01-01T00:00:00Z",
	)
	mock.ExpectQuery("INSERT INTO group_subscriptions").
		WithArgs(userID, groupID, true, false, false).
		WillReturnRows(subRow)

	handler.SubscribeToGroup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData)
	assert.Equal(t, userID, data["user_id"])
	assert.Equal(t, groupID, data["group_id"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestSubscribeToGroup_WithOrgSubscription tests that users with org subscription get error
func TestSubscribeToGroup_WithOrgSubscription(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &SubscriptionHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	groupID := "550e8400-e29b-41d4-a716-446655440002"
	orgID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+groupID+"/subscribe", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"groupId": groupID})

	w := httptest.NewRecorder()

	// Mock getting group's organization
	mock.ExpectQuery("SELECT organization_id FROM groups WHERE id = \\$1").
		WithArgs(groupID).
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).AddRow(orgID))

	// Mock checking for existing org subscription (user HAS org subscription)
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM org_subscriptions").
		WithArgs(userID, orgID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	handler.SubscribeToGroup(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	errObj, hasError := response["error"].(map[string]interface{})
	assert.True(t, hasError)
	assert.Contains(t, errObj["message"], "already subscribed to this organization")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUnsubscribeFromOrganization tests organization unsubscription
func TestUnsubscribeFromOrganization(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &SubscriptionHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	orgID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+orgID+"/subscribe", nil)
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"orgId": orgID})

	w := httptest.NewRecorder()

	// Mock deletion
	mock.ExpectExec("DELETE FROM org_subscriptions WHERE user_id = \\$1 AND organization_id = \\$2").
		WithArgs(userID, orgID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	handler.UnsubscribeFromOrganization(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateOrgSubscription tests updating notification preferences
func TestUpdateOrgSubscription(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &SubscriptionHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	orgID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	body := `{"notify_email": false, "notify_sms": true}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/organizations/"+orgID+"/subscribe", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"orgId": orgID})

	w := httptest.NewRecorder()

	notifyEmail := false
	notifySMS := true

	// Mock update
	subRow := sqlmock.NewRows([]string{
		"id", "user_id", "organization_id", "notify_email", "notify_sms", "notify_discord", "created_at",
	}).AddRow(
		"sub-id", userID, orgID, false, true, false, "2025-01-01T00:00:00Z",
	)
	mock.ExpectQuery("UPDATE org_subscriptions SET").
		WithArgs(userID, orgID, &notifyEmail, &notifySMS, sqlmock.AnyArg()).
		WillReturnRows(subRow)

	handler.UpdateOrgSubscription(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData)
	assert.False(t, data["notify_email"].(bool))
	assert.True(t, data["notify_sms"].(bool))

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetMySubscriptions tests fetching user's subscriptions
func TestGetMySubscriptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &SubscriptionHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	userID := "550e8400-e29b-41d4-a716-446655440001"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/subscriptions", nil)
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))

	w := httptest.NewRecorder()

	// Mock org subscriptions query
	orgRows := sqlmock.NewRows([]string{
		"id", "user_id", "organization_id", "notify_email", "notify_sms", "notify_discord",
		"created_at", "name", "slug", "logo_url",
	}).AddRow(
		"sub-1", userID, "org-1", true, false, false,
		"2025-01-01T00:00:00Z", "Org Name", "org-slug", nil,
	)
	mock.ExpectQuery("SELECT (.+) FROM org_subscriptions os JOIN organizations o").
		WithArgs(userID).
		WillReturnRows(orgRows)

	// Mock group subscriptions query
	groupRows := sqlmock.NewRows([]string{
		"id", "user_id", "group_id", "notify_email", "notify_sms", "notify_discord",
		"created_at", "name", "slug", "organization_id",
	}).AddRow(
		"sub-2", userID, "group-1", true, false, false,
		"2025-01-01T00:00:00Z", "Group Name", "group-slug", "org-1",
	)
	mock.ExpectQuery("SELECT (.+) FROM group_subscriptions gs JOIN groups g").
		WithArgs(userID).
		WillReturnRows(groupRows)

	handler.GetMySubscriptions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData)

	orgSubs := data["org_subscriptions"].([]interface{})
	groupSubs := data["group_subscriptions"].([]interface{})

	assert.Len(t, orgSubs, 1)
	assert.Len(t, groupSubs, 1)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSubscribeHost tests the auto-subscription helper
func TestAutoSubscribeHost(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &SubscriptionHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	userID := "550e8400-e29b-41d4-a716-446655440001"
	groupID := "550e8400-e29b-41d4-a716-446655440002"
	orgID := "550e8400-e29b-41d4-a716-446655440000"

	ctx := context.Background()

	// Mock getting group's organization
	mock.ExpectQuery("SELECT organization_id FROM groups WHERE id = \\$1").
		WithArgs(groupID).
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).AddRow(orgID))

	// Mock checking for org subscription (not exists)
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM org_subscriptions").
		WithArgs(userID, orgID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock checking for group subscription (not exists)
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM group_subscriptions").
		WithArgs(userID, groupID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock creating group subscription
	mock.ExpectExec("INSERT INTO group_subscriptions").
		WithArgs(userID, groupID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = handler.AutoSubscribeHost(ctx, userID, groupID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestAutoSubscribeHost_WithOrgSubscription tests no duplicate subscription created
func TestAutoSubscribeHost_WithOrgSubscription(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &SubscriptionHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	userID := "550e8400-e29b-41d4-a716-446655440001"
	groupID := "550e8400-e29b-41d4-a716-446655440002"
	orgID := "550e8400-e29b-41d4-a716-446655440000"

	ctx := context.Background()

	// Mock getting group's organization
	mock.ExpectQuery("SELECT organization_id FROM groups WHERE id = \\$1").
		WithArgs(groupID).
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).AddRow(orgID))

	// Mock checking for org subscription (EXISTS)
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM org_subscriptions").
		WithArgs(userID, orgID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Should NOT create group subscription

	err = handler.AutoSubscribeHost(ctx, userID, groupID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
