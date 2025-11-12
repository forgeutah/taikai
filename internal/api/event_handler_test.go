package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/forgeutah/taikai/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create chi context with URL params
func createChiContextForEvents(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, val := range params {
		rctx.URLParams.Add(key, val)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// TestUpdateSingleRecurringEvent tests editing a single event instance
func TestUpdateSingleRecurringEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Request body
	body := `{"title": "Updated Title", "recurrence_scope": "this_event"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+eventID, strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()

	// Mock checking if event is part of recurring series
	parentID := "550e8400-e29b-41d4-a716-446655440002"
	recurringCheckRow := sqlmock.NewRows([]string{"parent_event_id", "is_recurring_parent"}).
		AddRow(parentID, false) // This is a child event
	mock.ExpectQuery("SELECT parent_event_id, is_recurring_parent FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(recurringCheckRow)

	// Mock the update query that disconnects from series
	updatedEventRow := sqlmock.NewRows([]string{
		"id", "group_id", "title", "slug", "description", "status",
		"start_time", "end_time", "timezone", "venue_id", "capacity",
		"youtube_url", "parent_event_id", "recurrence_rule",
		"recurrence_end_date", "recurrence_count", "is_recurring_parent",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		eventID, "550e8400-e29b-41d4-a716-446655440003", "Updated Title", "event-slug", "Description", "published",
		time.Now(), time.Now().Add(1*time.Hour), "UTC", nil, nil,
		nil, nil, nil, nil, nil, false,
		"550e8400-e29b-41d4-a716-446655440004", time.Now(), time.Now(),
	)
	mock.ExpectQuery("UPDATE events SET (.+) WHERE id = \\$1 RETURNING").
		WithArgs(eventID, "Updated Title", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(updatedEventRow)

	// Execute the handler
	handler.UpdateEvent(w, req)

	// Assertions
	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Print response for debugging if failed
	if w.Code != http.StatusOK {
		t.Logf("Response code: %d, Body: %+v", w.Code, response)
	}

	assert.Equal(t, http.StatusOK, w.Code)

	// Should have data field with updated event
	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	if hasData {
		assert.Equal(t, "Updated Title", data["title"])
		if data["parent_event_id"] != nil {
			assert.Nil(t, data["parent_event_id"], "Event should be disconnected from series")
		}
		if data["is_recurring_parent"] != nil {
			assert.False(t, data["is_recurring_parent"].(bool), "Event should not be recurring parent")
		}
	}

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateFutureRecurringEvents tests editing this and future events
func TestUpdateFutureRecurringEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"
	parentID := "550e8400-e29b-41d4-a716-446655440002"
	userID := "550e8400-e29b-41d4-a716-446655440001"
	currentStartTime := time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC)

	// Request body
	body := `{"description": "Updated Description", "recurrence_scope": "future_events"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+eventID, strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()


	// Mock checking if event is part of recurring series
	recurringCheckRow := sqlmock.NewRows([]string{"parent_event_id", "is_recurring_parent"}).
		AddRow(parentID, false) // This is a child event with a parent
	mock.ExpectQuery("SELECT parent_event_id, is_recurring_parent FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(recurringCheckRow)

	// Mock getting current event's start time
	startTimeRow := sqlmock.NewRows([]string{"start_time"}).AddRow(currentStartTime)
	mock.ExpectQuery("SELECT start_time FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(startTimeRow)

	// Mock updating future events
	mock.ExpectExec("UPDATE events SET (.+) WHERE \\(id = \\$1 OR parent_event_id = \\$1\\) AND start_time >= \\$9").
		WithArgs(parentID, sqlmock.AnyArg(), "Updated Description", sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), currentStartTime).
		WillReturnResult(sqlmock.NewResult(0, 3)) // 3 events updated

	// Mock fetching updated event
	updatedEventRow := sqlmock.NewRows([]string{
		"id", "group_id", "title", "slug", "description", "status",
		"start_time", "end_time", "timezone", "venue_id", "capacity",
		"youtube_url", "parent_event_id", "recurrence_rule",
		"recurrence_end_date", "recurrence_count", "is_recurring_parent",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		eventID, "550e8400-e29b-41d4-a716-446655440003", "Event Title", "event-slug", "Updated Description", "published",
		currentStartTime, currentStartTime.Add(1*time.Hour), "UTC", nil, nil,
		nil, parentID, nil, nil, nil, false,
		"550e8400-e29b-41d4-a716-446655440004", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT (.+) FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(updatedEventRow)

	// Execute the handler
	handler.UpdateEvent(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should have data field with updated event
	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, "Updated Description", data["description"])
	assert.Equal(t, parentID, data["parent_event_id"], "Event should still be linked to parent")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateAllRecurringEvents tests editing entire series
func TestUpdateAllRecurringEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"
	parentID := "550e8400-e29b-41d4-a716-446655440002"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Request body
	body := `{"title": "Updated Series Title", "recurrence_scope": "all_events"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+eventID, strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()


	// Mock checking if event is part of recurring series
	recurringCheckRow := sqlmock.NewRows([]string{"parent_event_id", "is_recurring_parent"}).
		AddRow(parentID, false) // This is a child event
	mock.ExpectQuery("SELECT parent_event_id, is_recurring_parent FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(recurringCheckRow)

	// Mock updating parent event
	mock.ExpectExec("UPDATE events SET (.+) WHERE id = \\$1").
		WithArgs(parentID, "Updated Series Title", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock updating all child events
	mock.ExpectExec("UPDATE events SET (.+) WHERE parent_event_id = \\$1").
		WithArgs(parentID, "Updated Series Title", sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 5)) // 5 child events updated

	// Mock fetching updated event
	updatedEventRow := sqlmock.NewRows([]string{
		"id", "group_id", "title", "slug", "description", "status",
		"start_time", "end_time", "timezone", "venue_id", "capacity",
		"youtube_url", "parent_event_id", "recurrence_rule",
		"recurrence_end_date", "recurrence_count", "is_recurring_parent",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		eventID, "550e8400-e29b-41d4-a716-446655440003", "Updated Series Title", "event-slug", "Description", "published",
		time.Now(), time.Now().Add(1*time.Hour), "UTC", nil, nil,
		nil, parentID, nil, nil, nil, false,
		"550e8400-e29b-41d4-a716-446655440004", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT (.+) FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(updatedEventRow)

	// Execute the handler
	handler.UpdateEvent(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should have data field with updated event
	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, "Updated Series Title", data["title"])

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateRecurringEvent_ParentEvent tests editing when the event itself is the parent
func TestUpdateRecurringEvent_ParentEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	parentID := "550e8400-e29b-41d4-a716-446655440002"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Request body
	body := `{"capacity": 50, "recurrence_scope": "all_events"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+parentID, strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"eventId": parentID})

	w := httptest.NewRecorder()


	// Mock checking if event is part of recurring series
	recurringCheckRow := sqlmock.NewRows([]string{"parent_event_id", "is_recurring_parent"}).
		AddRow(nil, true) // This IS the parent event
	mock.ExpectQuery("SELECT parent_event_id, is_recurring_parent FROM events WHERE id = \\$1").
		WithArgs(parentID).
		WillReturnRows(recurringCheckRow)

	// Mock updating parent event
	mock.ExpectExec("UPDATE events SET (.+) WHERE id = \\$1").
		WithArgs(parentID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 50, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock updating all child events
	mock.ExpectExec("UPDATE events SET (.+) WHERE parent_event_id = \\$1").
		WithArgs(parentID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), 50, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 10)) // 10 child events updated

	// Mock fetching updated event
	updatedEventRow := sqlmock.NewRows([]string{
		"id", "group_id", "title", "slug", "description", "status",
		"start_time", "end_time", "timezone", "venue_id", "capacity",
		"youtube_url", "parent_event_id", "recurrence_rule",
		"recurrence_end_date", "recurrence_count", "is_recurring_parent",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		parentID, "550e8400-e29b-41d4-a716-446655440003", "Event Title", "event-slug", "Description", "published",
		time.Now(), time.Now().Add(1*time.Hour), "UTC", nil, 50,
		nil, nil, "FREQ=WEEKLY;COUNT=10", nil, 10, true,
		"550e8400-e29b-41d4-a716-446655440004", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT (.+) FROM events WHERE id = \\$1").
		WithArgs(parentID).
		WillReturnRows(updatedEventRow)

	// Execute the handler
	handler.UpdateEvent(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should have data field with updated event
	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, float64(50), data["capacity"], "Capacity should be updated")
	assert.True(t, data["is_recurring_parent"].(bool), "Should still be recurring parent")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateRecurringEvent_InvalidScope tests error handling for invalid scope
func TestUpdateRecurringEvent_InvalidScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Request body with invalid scope
	body := `{"title": "Updated", "recurrence_scope": "invalid_scope"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+eventID, strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()


	// Mock checking if event is part of recurring series
	parentID := "550e8400-e29b-41d4-a716-446655440002"
	recurringCheckRow := sqlmock.NewRows([]string{"parent_event_id", "is_recurring_parent"}).
		AddRow(parentID, false)
	mock.ExpectQuery("SELECT parent_event_id, is_recurring_parent FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(recurringCheckRow)

	// Execute the handler
	handler.UpdateEvent(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should have error field
	errObj, hasError := response["error"].(map[string]interface{})
	assert.True(t, hasError, "Response should have error field")
	assert.Contains(t, errObj["message"], "recurrence_scope")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateNonRecurringEvent tests that normal events still work
func TestUpdateNonRecurringEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "550e8400-e29b-41d4-a716-446655440001"

	// Request body without recurrence_scope
	body := `{"title": "Updated Title", "description": "Updated Description"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/"+eventID, strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()


	// Mock checking if event is part of recurring series
	recurringCheckRow := sqlmock.NewRows([]string{"parent_event_id", "is_recurring_parent"}).
		AddRow(nil, false) // Not a recurring event
	mock.ExpectQuery("SELECT parent_event_id, is_recurring_parent FROM events WHERE id = \\$1").
		WithArgs(eventID).
		WillReturnRows(recurringCheckRow)

	// Mock normal update query
	updatedEventRow := sqlmock.NewRows([]string{
		"id", "group_id", "title", "slug", "description", "status",
		"start_time", "end_time", "timezone", "venue_id", "capacity",
		"youtube_url", "parent_event_id", "recurrence_rule",
		"recurrence_end_date", "recurrence_count", "is_recurring_parent",
		"created_by", "created_at", "updated_at",
	}).AddRow(
		eventID, "550e8400-e29b-41d4-a716-446655440003", "Updated Title", "event-slug", "Updated Description", "published",
		time.Now(), time.Now().Add(1*time.Hour), "UTC", nil, nil,
		nil, nil, nil, nil, nil, false,
		"550e8400-e29b-41d4-a716-446655440004", time.Now(), time.Now(),
	)
	mock.ExpectQuery("UPDATE events SET (.+) WHERE id = \\$1 RETURNING").
		WithArgs(eventID, "Updated Title", "Updated Description", sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(updatedEventRow)

	// Execute the handler
	handler.UpdateEvent(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should have data field with updated event
	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData, "Response should have data field")
	assert.Equal(t, "Updated Title", data["title"])
	assert.Equal(t, "Updated Description", data["description"])
	assert.False(t, data["is_recurring_parent"].(bool))

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
