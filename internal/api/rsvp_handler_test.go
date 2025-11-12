package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/forgeutah/taikai/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to set up permission checker mocks
func setupPermissionMocks(mock sqlmock.Sqlmock, eventID string, canManage bool) {
	if canManage {
		// Mock IsEventHost check - returns true
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM event_hosts").
			WithArgs(sqlmock.AnyArg(), eventID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	} else {
		// Mock IsEventHost check - returns false
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM event_hosts").
			WithArgs(sqlmock.AnyArg(), eventID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock group_id lookup (for subsequent group/org admin checks)
		mock.ExpectQuery("SELECT group_id FROM events WHERE id = \\$1").
			WithArgs(eventID).
			WillReturnRows(sqlmock.NewRows([]string{"group_id"}).AddRow("group-1"))

		// Mock IsGroupAdmin check - returns false
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM group_admins").
			WithArgs(sqlmock.AnyArg(), "group-1").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Mock org_id lookup
		mock.ExpectQuery("SELECT organization_id FROM groups WHERE id = \\$1").
			WithArgs("group-1").
			WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).AddRow("org-1"))

		// Mock IsOrgAdmin check - returns false
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM org_admins").
			WithArgs(sqlmock.AnyArg(), "org-1").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	}
}

// Helper to create chi context with URL params
func createChiContext(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, val := range params {
		rctx.URLParams.Add(key, val)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateOrUpdateRSVP_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "user-123"

	// Mock event capacity check
	mock.ExpectQuery("SELECT capacity FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(sqlmock.NewRows([]string{"capacity"}).AddRow(10))

	// Mock RSVP count check (0 existing RSVPs)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps").
		WithArgs(eventID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock RSVP insert
	now := time.Now()
	mock.ExpectQuery("INSERT INTO event_rsvps").
		WithArgs(userID, eventID, "attending").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "event_id", "status", "created_at", "updated_at"}).
			AddRow("rsvp-123", userID, eventID, "attending", now, now))

	// Mock capacity notification check
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps WHERE event_id = \\$1 AND status = 'attending'").
		WithArgs(eventID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	reqBody := bytes.NewBufferString(`{"status":"attending"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/rsvp", reqBody)
	w := httptest.NewRecorder()

	// Add user to context
	ctx := auth.SetUserContext(req.Context(), userID, "test@example.com", "Test User")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": eventID})

	handler.CreateOrUpdateRSVP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, "rsvp-123", data["id"])
	assert.Equal(t, "attending", data["status"])

	// Give goroutine time to finish
	time.Sleep(100 * time.Millisecond)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCreateOrUpdateRSVP_CapacityEnforcement(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "user-123"

	// Mock event capacity check (capacity = 2)
	mock.ExpectQuery("SELECT capacity FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(sqlmock.NewRows([]string{"capacity"}).AddRow(2))

	// Mock RSVP count check (2 existing RSVPs, at capacity)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps").
		WithArgs(eventID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	reqBody := bytes.NewBufferString(`{"status":"attending"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/rsvp", reqBody)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "test@example.com", "Test User")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": eventID})

	handler.CreateOrUpdateRSVP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	errObj, hasError := response["error"].(map[string]interface{})
	assert.True(t, hasError, "Response should have error field")
	assert.Contains(t, errObj["message"], "capacity")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCreateOrUpdateRSVP_Unauthorized(t *testing.T) {
	handler := &RSVPHandler{
		db:      nil,
		checker: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/test-event/rsvp", nil)
	w := httptest.NewRecorder()

	req = createChiContext(req, map[string]string{"eventId": "123e4567-e89b-12d3-a456-426614174000"})

	// No user in context - should return 401
	handler.CreateOrUpdateRSVP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	_, hasError := response["error"]
	assert.True(t, hasError, "Response should have error field")
}

func TestCreateOrUpdateRSVP_InvalidEventID(t *testing.T) {
	handler := &RSVPHandler{
		db:      nil,
		checker: nil,
	}

	reqBody := bytes.NewBufferString(`{"status":"attending"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/invalid-id/rsvp", reqBody)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), "user-123", "test@example.com", "Test User")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": "invalid-uuid"})

	handler.CreateOrUpdateRSVP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	_, hasError := response["error"]
	assert.True(t, hasError, "Response should have error field")
}

func TestCreateOrUpdateRSVP_InvalidStatus(t *testing.T) {
	handler := &RSVPHandler{
		db:      nil,
		checker: nil,
	}

	reqBody := bytes.NewBufferString(`{"status":"maybe"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/test-event/rsvp", reqBody)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), "user-123", "test@example.com", "Test User")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": "123e4567-e89b-12d3-a456-426614174000"})

	handler.CreateOrUpdateRSVP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	_, hasError := response["error"]
	assert.True(t, hasError, "Response should have error field")
	errObj := response["error"].(map[string]interface{})
	message := errObj["message"].(string)
	assert.Contains(t, message, "attending")
	assert.Contains(t, message, "not_attending")
}

func TestCreateOrUpdateRSVP_EventNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "user-123"

	// Mock event not found
	mock.ExpectQuery("SELECT capacity FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnError(sql.ErrNoRows)

	reqBody := bytes.NewBufferString(`{"status":"attending"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/rsvp", reqBody)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "test@example.com", "Test User")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": eventID})

	handler.CreateOrUpdateRSVP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDeleteRSVP_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "user-123"

	mock.ExpectExec("DELETE FROM event_rsvps WHERE user_id = \\$1 AND event_id = \\$2").
		WithArgs(userID, eventID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/"+eventID+"/rsvp", nil)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "test@example.com", "Test User")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": eventID})

	handler.DeleteRSVP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDeleteRSVP_Unauthorized(t *testing.T) {
	handler := &RSVPHandler{
		db:      nil,
		checker: nil,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/test-event/rsvp", nil)
	w := httptest.NewRecorder()

	req = createChiContext(req, map[string]string{"eventId": "123e4567-e89b-12d3-a456-426614174000"})

	handler.DeleteRSVP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetEventRSVPs_PublicCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "123e4567-e89b-12d3-a456-426614174000"

	// Mock RSVP count query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps WHERE event_id = \\$1 AND status = 'attending'").
		WithArgs(eventID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// No permission mocks needed - unauthenticated user

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/"+eventID+"/rsvps", nil)
	w := httptest.NewRecorder()

	req = createChiContext(req, map[string]string{"eventId": eventID})

	handler.GetEventRSVPs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, float64(5), data["count"])

	// Public user should NOT see rsvps array
	_, hasRSVPs := data["rsvps"]
	assert.False(t, hasRSVPs, "Public user should not see RSVP details")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetEventRSVPs_HostSeesDetails(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "host-123"

	// Mock RSVP count query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps WHERE event_id = \\$1 AND status = 'attending'").
		WithArgs(eventID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock permission check - IsEventHost returns true (user is host)
	mock.ExpectQuery("SELECT EXISTS\\(.*FROM event_hosts.*WHERE user_id = \\$1 AND event_id = \\$2").
		WithArgs(userID, eventID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock full RSVP list with user details
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "event_id", "status", "created_at", "updated_at", "email", "name", "avatar_url"}).
		AddRow("rsvp-1", "user-1", eventID, "attending", now, now, "user1@example.com", "User One", "https://example.com/avatar1.jpg").
		AddRow("rsvp-2", "user-2", eventID, "attending", now, now, "user2@example.com", "User Two", nil).
		AddRow("rsvp-3", "user-3", eventID, "not_attending", now, now, "user3@example.com", "User Three", "https://example.com/avatar3.jpg")

	mock.ExpectQuery("SELECT r.id, r.user_id, r.event_id, r.status, r.created_at, r.updated_at").
		WithArgs(eventID).
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/"+eventID+"/rsvps", nil)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "host@example.com", "Host User")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": eventID})

	handler.GetEventRSVPs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, float64(3), data["count"])

	// Host should see full RSVP list with user details
	rsvps, hasRSVPs := data["rsvps"].([]interface{})
	assert.True(t, hasRSVPs, "Host should see RSVP details")
	assert.Len(t, rsvps, 3)

	// Verify first RSVP has user details
	firstRSVP := rsvps[0].(map[string]interface{})
	assert.Equal(t, "user1@example.com", firstRSVP["email"])
	assert.Equal(t, "User One", firstRSVP["name"])
	assert.Equal(t, "https://example.com/avatar1.jpg", firstRSVP["avatar_url"])

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetMyRSVPs_Pagination(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	userID := "user-123"

	// Mock total count query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps r JOIN events e").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))

	// Mock paginated results (page 1, limit 10)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "event_id", "status", "created_at", "updated_at",
		"id", "group_id", "title", "slug", "description", "status",
		"start_time", "end_time", "timezone", "venue_id", "capacity",
		"youtube_url", "parent_event_id", "recurrence_rule", "recurrence_end_date",
		"recurrence_count", "is_recurring_parent", "created_by", "created_at", "updated_at",
	})

	for i := 1; i <= 10; i++ {
		rows.AddRow(
			"rsvp-"+string(rune(i)), userID, "event-"+string(rune(i)), "attending", now, now,
			"event-"+string(rune(i)), "group-1", "Event "+string(rune(i)), "event-"+string(rune(i)), "Description", "published",
			now.Add(24*time.Hour), now.Add(26*time.Hour), "America/Denver", nil, 50,
			nil, nil, nil, nil, nil, false, userID, now, now,
		)
	}

	mock.ExpectQuery("SELECT r.id, r.user_id, r.event_id, r.status").
		WithArgs(userID, 10, 0).
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/rsvps?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "test@example.com", "Test User")
	req = req.WithContext(ctx)

	handler.GetMyRSVPs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, float64(25), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(10), data["limit"])

	rsvps := data["rsvps"].([]interface{})
	assert.Len(t, rsvps, 10)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetMyRSVPs_OnlyUpcoming(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	userID := "user-123"

	// Mock total count query - only 3 upcoming events
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps r JOIN events e").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock results - all future events
	now := time.Now()
	futureTime := now.Add(48 * time.Hour)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "event_id", "status", "created_at", "updated_at",
		"id", "group_id", "title", "slug", "description", "status",
		"start_time", "end_time", "timezone", "venue_id", "capacity",
		"youtube_url", "parent_event_id", "recurrence_rule", "recurrence_end_date",
		"recurrence_count", "is_recurring_parent", "created_by", "created_at", "updated_at",
	}).
		AddRow(
			"rsvp-1", userID, "event-1", "attending", now, now,
			"event-1", "group-1", "Future Event 1", "future-1", "Description", "published",
			futureTime, futureTime.Add(2*time.Hour), "America/Denver", nil, 50,
			nil, nil, nil, nil, nil, false, userID, now, now,
		).
		AddRow(
			"rsvp-2", userID, "event-2", "attending", now, now,
			"event-2", "group-1", "Future Event 2", "future-2", "Description", "published",
			futureTime.Add(24*time.Hour), futureTime.Add(26*time.Hour), "America/Denver", nil, 50,
			nil, nil, nil, nil, nil, false, userID, now, now,
		).
		AddRow(
			"rsvp-3", userID, "event-3", "attending", now, now,
			"event-3", "group-1", "Future Event 3", "future-3", "Description", "published",
			futureTime.Add(48*time.Hour), futureTime.Add(50*time.Hour), "America/Denver", nil, 50,
			nil, nil, nil, nil, nil, false, userID, now, now,
		)

	mock.ExpectQuery("SELECT r.id, r.user_id, r.event_id, r.status").
		WithArgs(userID, 20, 0).
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/rsvps", nil)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "test@example.com", "Test User")
	req = req.WithContext(ctx)

	handler.GetMyRSVPs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, float64(3), data["total"])

	rsvps := data["rsvps"].([]interface{})
	assert.Len(t, rsvps, 3, "Should only return upcoming events")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetMyHostedEvents_WithRSVPCounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	userID := "host-123"

	// Mock total count
	mock.ExpectQuery("SELECT COUNT\\(DISTINCT e.id\\) FROM events e JOIN event_hosts eh").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Mock hosted events with RSVP counts
	now := time.Now()
	futureTime := now.Add(48 * time.Hour)
	rows := sqlmock.NewRows([]string{
		"id", "group_id", "title", "slug", "description", "status",
		"start_time", "end_time", "timezone", "venue_id", "capacity",
		"youtube_url", "parent_event_id", "recurrence_rule", "recurrence_end_date",
		"recurrence_count", "is_recurring_parent", "created_by", "created_at", "updated_at", "rsvp_count",
	}).
		AddRow(
			"event-1", "group-1", "Event 1", "event-1", "Description", "published",
			futureTime, futureTime.Add(2*time.Hour), "America/Denver", nil, 50,
			nil, nil, nil, nil, nil, false, userID, now, now, 10,
		).
		AddRow(
			"event-2", "group-1", "Event 2", "event-2", "Description", "published",
			futureTime.Add(24*time.Hour), futureTime.Add(26*time.Hour), "America/Denver", nil, 50,
			nil, nil, nil, nil, nil, false, userID, now, now, 5,
		).
		AddRow(
			"event-3", "group-1", "Event 3", "event-3", "Description", "published",
			futureTime.Add(48*time.Hour), futureTime.Add(50*time.Hour), "America/Denver", nil, 50,
			nil, nil, nil, nil, nil, false, userID, now, now, 0,
		)

	mock.ExpectQuery("SELECT DISTINCT e.id, e.group_id, e.title").
		WithArgs(userID, 20, 0).
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/events", nil)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "host@example.com", "Host User")
	req = req.WithContext(ctx)

	handler.GetMyHostedEvents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, float64(3), data["total"])

	events := data["events"].([]interface{})
	assert.Len(t, events, 3)

	// Verify RSVP counts
	event1 := events[0].(map[string]interface{})
	assert.Equal(t, float64(10), event1["rsvp_count"])

	event2 := events[1].(map[string]interface{})
	assert.Equal(t, float64(5), event2["rsvp_count"])

	event3 := events[2].(map[string]interface{})
	assert.Equal(t, float64(0), event3["rsvp_count"])

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRSVP_UpdateFromAttendingToNotAttending(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "user-123"

	// Mock event capacity check
	mock.ExpectQuery("SELECT capacity FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(sqlmock.NewRows([]string{"capacity"}).AddRow(10))

	// Mock RSVP update (upsert pattern)
	now := time.Now()
	mock.ExpectQuery("INSERT INTO event_rsvps").
		WithArgs(userID, eventID, "not_attending").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "event_id", "status", "created_at", "updated_at"}).
			AddRow("rsvp-123", userID, eventID, "not_attending", now, now))

	reqBody := bytes.NewBufferString(`{"status":"not_attending"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/rsvp", reqBody)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "test@example.com", "Test User")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": eventID})

	handler.CreateOrUpdateRSVP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, "not_attending", data["status"])

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRSVP_CapacityFreedWhenChangedToNotAttending(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &RSVPHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "123e4567-e89b-12d3-a456-426614174000"
	userID := "user-c"

	// User C tries to RSVP after User A changed to not_attending
	// Capacity check should now show 1 RSVP (only User B)
	mock.ExpectQuery("SELECT capacity FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(sqlmock.NewRows([]string{"capacity"}).AddRow(2))

	// Mock count check - only 1 RSVP now (User B), so User C can RSVP
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps").
		WithArgs(eventID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock successful RSVP insert
	now := time.Now()
	mock.ExpectQuery("INSERT INTO event_rsvps").
		WithArgs(userID, eventID, "attending").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "event_id", "status", "created_at", "updated_at"}).
			AddRow("rsvp-c", userID, eventID, "attending", now, now))

	// Mock capacity notification check
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM event_rsvps WHERE event_id = \\$1 AND status = 'attending'").
		WithArgs(eventID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	reqBody := bytes.NewBufferString(`{"status":"attending"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/rsvp", reqBody)
	w := httptest.NewRecorder()

	ctx := auth.SetUserContext(req.Context(), userID, "userc@example.com", "User C")
	req = req.WithContext(ctx)
	req = createChiContext(req, map[string]string{"eventId": eventID})

	handler.CreateOrUpdateRSVP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	_, hasData := response["data"]
	assert.True(t, hasData, "Response should have data field")

	// Give goroutine time to finish
	time.Sleep(100 * time.Millisecond)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
