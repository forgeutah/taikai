package api

import (
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

// TestGetEventSponsors tests fetching sponsors for an event
func TestGetEventSponsors(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/"+eventID+"/sponsors", nil)
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()

	// Mock sponsors query
	sponsorRows := sqlmock.NewRows([]string{
		"id", "event_id", "company_name", "logo_url", "description", "email", "website", "display_order", "created_at",
	}).AddRow(
		"sponsor-1", eventID, "Acme Corp", "https://example.com/logo.png", "Amazing sponsor", "sponsor@acme.com", "https://acme.com", 0, "2025-01-01T00:00:00Z",
	).AddRow(
		"sponsor-2", eventID, "Tech Inc", nil, nil, "contact@tech.com", nil, 1, "2025-01-01T00:00:00Z",
	)

	mock.ExpectQuery("SELECT (.+) FROM event_sponsors WHERE event_id = \\$1").
		WithArgs(eventID).
		WillReturnRows(sponsorRows)

	handler.GetEventSponsors(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	firstSponsor := data[0].(map[string]interface{})
	assert.Equal(t, "Acme Corp", firstSponsor["company_name"])
	assert.Equal(t, "sponsor@acme.com", firstSponsor["email"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCreateSponsor tests creating a sponsor
func TestCreateSponsor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"

	body := `{
		"company_name": "Acme Corp",
		"logo_url": "https://example.com/logo.png",
		"description": "Amazing sponsor",
		"email": "sponsor@acme.com",
		"website": "https://acme.com",
		"display_order": 0
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/sponsors", strings.NewReader(body))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()

	// Mock sponsor creation
	sponsorRow := sqlmock.NewRows([]string{
		"id", "event_id", "company_name", "logo_url", "description", "email", "website", "display_order", "created_at",
	}).AddRow(
		"sponsor-1", eventID, "Acme Corp", "https://example.com/logo.png", "Amazing sponsor", "sponsor@acme.com", "https://acme.com", 0, "2025-01-01T00:00:00Z",
	)

	mock.ExpectQuery("INSERT INTO event_sponsors").
		WithArgs(eventID, "Acme Corp", "https://example.com/logo.png", "Amazing sponsor", "sponsor@acme.com", "https://acme.com", 0).
		WillReturnRows(sponsorRow)

	handler.CreateSponsor(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData)
	assert.Equal(t, "Acme Corp", data["company_name"])
	assert.Equal(t, "sponsor@acme.com", data["email"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCreateSponsor_MissingCompanyName tests validation
func TestCreateSponsor_MissingCompanyName(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"

	body := `{"email": "sponsor@acme.com"}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/sponsors", strings.NewReader(body))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()

	handler.CreateSponsor(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	errObj, hasError := response["error"].(map[string]interface{})
	assert.True(t, hasError)
	assert.Contains(t, errObj["message"], "company_name")
}

// TestCreateSponsor_MissingEmail tests email validation
func TestCreateSponsor_MissingEmail(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"

	body := `{"company_name": "Acme Corp"}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/sponsors", strings.NewReader(body))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()

	handler.CreateSponsor(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	errObj, hasError := response["error"].(map[string]interface{})
	assert.True(t, hasError)
	assert.Contains(t, errObj["message"], "email")
}

// TestUpdateSponsor tests updating a sponsor
func TestUpdateSponsor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	sponsorID := "sponsor-1"
	eventID := "550e8400-e29b-41d4-a716-446655440000"

	body := `{"company_name": "Acme Corporation", "website": "https://acme.corp"}`

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/events/"+eventID+"/sponsors/"+sponsorID, strings.NewReader(body))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID, "sponsorId": sponsorID})

	w := httptest.NewRecorder()

	// Mock sponsor update
	sponsorRow := sqlmock.NewRows([]string{
		"id", "event_id", "company_name", "logo_url", "description", "email", "website", "display_order", "created_at",
	}).AddRow(
		sponsorID, eventID, "Acme Corporation", "https://example.com/logo.png", "Amazing sponsor", "sponsor@acme.com", "https://acme.corp", 0, "2025-01-01T00:00:00Z",
	)

	mock.ExpectQuery("UPDATE event_sponsors SET").
		WithArgs(sponsorID, "Acme Corporation", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "https://acme.corp", sqlmock.AnyArg()).
		WillReturnRows(sponsorRow)

	handler.UpdateSponsor(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData)
	assert.Equal(t, "Acme Corporation", data["company_name"])
	assert.Equal(t, "https://acme.corp", data["website"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateSponsor_NotFound tests updating non-existent sponsor
func TestUpdateSponsor_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	sponsorID := "sponsor-nonexistent"
	eventID := "550e8400-e29b-41d4-a716-446655440000"

	body := `{"company_name": "Acme Corporation"}`

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/events/"+eventID+"/sponsors/"+sponsorID, strings.NewReader(body))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID, "sponsorId": sponsorID})

	w := httptest.NewRecorder()

	// Mock sponsor not found
	mock.ExpectQuery("UPDATE event_sponsors SET").
		WithArgs(sponsorID, "Acme Corporation", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sqlmock.ErrCancelled) // Simulate no rows

	handler.UpdateSponsor(w, req)

	// Should return error
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// TestDeleteSponsor tests deleting a sponsor
func TestDeleteSponsor(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	sponsorID := "sponsor-1"
	eventID := "550e8400-e29b-41d4-a716-446655440000"

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/"+eventID+"/sponsors/"+sponsorID, nil)
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID, "sponsorId": sponsorID})

	w := httptest.NewRecorder()

	// Mock sponsor deletion
	mock.ExpectExec("DELETE FROM event_sponsors WHERE id = \\$1").
		WithArgs(sponsorID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	handler.DeleteSponsor(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetEventSponsors_EmptyList tests fetching sponsors when none exist
func TestGetEventSponsors_EmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/"+eventID+"/sponsors", nil)
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()

	// Mock empty result
	sponsorRows := sqlmock.NewRows([]string{
		"id", "event_id", "company_name", "logo_url", "description", "email", "website", "display_order", "created_at",
	})

	mock.ExpectQuery("SELECT (.+) FROM event_sponsors WHERE event_id = \\$1").
		WithArgs(eventID).
		WillReturnRows(sponsorRows)

	handler.GetEventSponsors(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Len(t, data, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCreateSponsor_OptionalFields tests creating sponsor with minimal fields
func TestCreateSponsor_OptionalFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	handler := &EventHandler{
		db:      db,
		checker: auth.NewPermissionChecker(db),
	}

	eventID := "550e8400-e29b-41d4-a716-446655440000"

	// Only required fields
	body := `{
		"company_name": "Minimal Corp",
		"email": "contact@minimal.com",
		"display_order": 0
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID+"/sponsors", strings.NewReader(body))
	req = createChiContextForEvents(req, map[string]string{"eventId": eventID})

	w := httptest.NewRecorder()

	// Mock sponsor creation with nulls
	sponsorRow := sqlmock.NewRows([]string{
		"id", "event_id", "company_name", "logo_url", "description", "email", "website", "display_order", "created_at",
	}).AddRow(
		"sponsor-1", eventID, "Minimal Corp", nil, nil, "contact@minimal.com", nil, 0, "2025-01-01T00:00:00Z",
	)

	mock.ExpectQuery("INSERT INTO event_sponsors").
		WithArgs(eventID, "Minimal Corp", nil, nil, "contact@minimal.com", nil, 0).
		WillReturnRows(sponsorRow)

	handler.CreateSponsor(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, hasData := response["data"].(map[string]interface{})
	assert.True(t, hasData)
	assert.Equal(t, "Minimal Corp", data["company_name"])
	assert.Nil(t, data["logo_url"])
	assert.Nil(t, data["description"])
	assert.Nil(t, data["website"])

	assert.NoError(t, mock.ExpectationsWereMet())
}
