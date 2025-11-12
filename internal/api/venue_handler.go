package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// VenueHandler handles venue-related requests
type VenueHandler struct {
	db *sql.DB
}

// NewVenueHandler creates a new venue handler
func NewVenueHandler(db *sql.DB) *VenueHandler {
	return &VenueHandler{
		db: db,
	}
}

// Venue response structure
type VenueResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Address    *string `json:"address"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postal_code"`
	Country    *string `json:"country"`
	Directions *string `json:"directions"`
	Capacity   *int    `json:"capacity"`
	Timezone   string  `json:"timezone"`
	IsVirtual  bool    `json:"is_virtual"`
	CreatedAt  string  `json:"created_at"`
}

// ListVenuesResponse for paginated venue list
type ListVenuesResponse struct {
	Venues []VenueResponse `json:"venues"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// CreateVenueRequest for creating a new venue
type CreateVenueRequest struct {
	Name       string  `json:"name"`
	Address    *string `json:"address"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postal_code"`
	Country    *string `json:"country"`
	Directions *string `json:"directions"`
	Capacity   *int    `json:"capacity"`
	Timezone   string  `json:"timezone"`
	IsVirtual  bool    `json:"is_virtual"`
}

// UpdateVenueRequest for updating venue
type UpdateVenueRequest struct {
	Name       *string `json:"name"`
	Address    *string `json:"address"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postal_code"`
	Country    *string `json:"country"`
	Directions *string `json:"directions"`
	Capacity   *int    `json:"capacity"`
	Timezone   *string `json:"timezone"`
	IsVirtual  *bool   `json:"is_virtual"`
}

// GetVenue handles GET /api/v1/venues/:venueId
func (h *VenueHandler) GetVenue(w http.ResponseWriter, r *http.Request) {
	venueID := chi.URLParam(r, "venueId")
	if venueID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "venue ID is required")
		return
	}

	if _, err := uuid.Parse(venueID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid venue ID format")
		return
	}

	query := `
		SELECT id, name, address, city, state, postal_code, country,
		       directions, capacity, timezone, is_virtual, created_at
		FROM venues
		WHERE id = $1
	`

	var venue VenueResponse
	var address, city, state, postalCode, country, directions sql.NullString
	var capacity sql.NullInt64
	err := h.db.QueryRowContext(r.Context(), query, venueID).Scan(
		&venue.ID, &venue.Name, &address, &city, &state, &postalCode, &country,
		&directions, &capacity, &venue.Timezone, &venue.IsVirtual, &venue.CreatedAt,
	)

	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Venue not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to get venue")
		return
	}

	// Convert nullable fields
	if address.Valid {
		venue.Address = &address.String
	}
	if city.Valid {
		venue.City = &city.String
	}
	if state.Valid {
		venue.State = &state.String
	}
	if postalCode.Valid {
		venue.PostalCode = &postalCode.String
	}
	if country.Valid {
		venue.Country = &country.String
	}
	if directions.Valid {
		venue.Directions = &directions.String
	}
	if capacity.Valid {
		cap := int(capacity.Int64)
		venue.Capacity = &cap
	}

	RespondSuccess(w, http.StatusOK, venue)
}

// ListVenues handles GET /api/v1/venues with search and pagination
func (h *VenueHandler) ListVenues(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Parse search parameter
	search := r.URL.Query().Get("search")

	// Count total venues matching search
	countQuery := `
		SELECT COUNT(*)
		FROM venues
		WHERE
			CASE
				WHEN $1::text != '' THEN
					name ILIKE '%' || $1 || '%' OR
					city ILIKE '%' || $1 || '%' OR
					state ILIKE '%' || $1 || '%'
				ELSE TRUE
			END
	`

	var total int
	err := h.db.QueryRowContext(r.Context(), countQuery, search).Scan(&total)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to count venues")
		return
	}

	// Query venues
	query := `
		SELECT id, name, address, city, state, postal_code, country,
		       directions, capacity, timezone, is_virtual, created_at
		FROM venues
		WHERE
			CASE
				WHEN $1::text != '' THEN
					name ILIKE '%' || $1 || '%' OR
					city ILIKE '%' || $1 || '%' OR
					state ILIKE '%' || $1 || '%'
				ELSE TRUE
			END
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.QueryContext(r.Context(), query, search, limit, offset)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to list venues")
		return
	}
	defer rows.Close()

	venues := []VenueResponse{}
	for rows.Next() {
		var venue VenueResponse
		var address, city, state, postalCode, country, directions sql.NullString
		var capacity sql.NullInt64

		err := rows.Scan(
			&venue.ID, &venue.Name, &address, &city, &state, &postalCode, &country,
			&directions, &capacity, &venue.Timezone, &venue.IsVirtual, &venue.CreatedAt,
		)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to scan venue")
			return
		}

		// Convert nullable fields
		if address.Valid {
			venue.Address = &address.String
		}
		if city.Valid {
			venue.City = &city.String
		}
		if state.Valid {
			venue.State = &state.String
		}
		if postalCode.Valid {
			venue.PostalCode = &postalCode.String
		}
		if country.Valid {
			venue.Country = &country.String
		}
		if directions.Valid {
			venue.Directions = &directions.String
		}
		if capacity.Valid {
			cap := int(capacity.Int64)
			venue.Capacity = &cap
		}

		venues = append(venues, venue)
	}

	response := ListVenuesResponse{
		Venues: venues,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}

	RespondSuccess(w, http.StatusOK, response)
}

// CreateVenue handles POST /api/v1/venues
func (h *VenueHandler) CreateVenue(w http.ResponseWriter, r *http.Request) {
	var req CreateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "name is required")
		return
	}
	if req.Timezone == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "timezone is required")
		return
	}

	query := `
		INSERT INTO venues (
			name, address, city, state, postal_code, country,
			directions, capacity, timezone, is_virtual
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, name, address, city, state, postal_code, country,
		          directions, capacity, timezone, is_virtual, created_at
	`

	var venue VenueResponse
	var address, city, state, postalCode, country, directions sql.NullString
	var capacity sql.NullInt64

	err := h.db.QueryRowContext(r.Context(), query,
		req.Name, req.Address, req.City, req.State, req.PostalCode, req.Country,
		req.Directions, req.Capacity, req.Timezone, req.IsVirtual,
	).Scan(
		&venue.ID, &venue.Name, &address, &city, &state, &postalCode, &country,
		&directions, &capacity, &venue.Timezone, &venue.IsVirtual, &venue.CreatedAt,
	)

	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to create venue")
		return
	}

	// Convert nullable fields
	if address.Valid {
		venue.Address = &address.String
	}
	if city.Valid {
		venue.City = &city.String
	}
	if state.Valid {
		venue.State = &state.String
	}
	if postalCode.Valid {
		venue.PostalCode = &postalCode.String
	}
	if country.Valid {
		venue.Country = &country.String
	}
	if directions.Valid {
		venue.Directions = &directions.String
	}
	if capacity.Valid {
		cap := int(capacity.Int64)
		venue.Capacity = &cap
	}

	RespondSuccess(w, http.StatusOK, venue)
}

// UpdateVenue handles PATCH /api/v1/venues/:venueId
func (h *VenueHandler) UpdateVenue(w http.ResponseWriter, r *http.Request) {
	venueID := chi.URLParam(r, "venueId")
	if venueID == "" {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "venue ID is required")
		return
	}

	if _, err := uuid.Parse(venueID); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "invalid venue ID format")
		return
	}

	var req UpdateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, ErrCodeBadRequest, "Invalid request body")
		return
	}

	query := `
		UPDATE venues
		SET
			name = COALESCE($2, name),
			address = COALESCE($3, address),
			city = COALESCE($4, city),
			state = COALESCE($5, state),
			postal_code = COALESCE($6, postal_code),
			country = COALESCE($7, country),
			directions = COALESCE($8, directions),
			capacity = COALESCE($9, capacity),
			timezone = COALESCE($10, timezone),
			is_virtual = COALESCE($11, is_virtual)
		WHERE id = $1
		RETURNING id, name, address, city, state, postal_code, country,
		          directions, capacity, timezone, is_virtual, created_at
	`

	var venue VenueResponse
	var address, city, state, postalCode, country, directions sql.NullString
	var capacity sql.NullInt64

	err := h.db.QueryRowContext(r.Context(), query,
		venueID, req.Name, req.Address, req.City, req.State, req.PostalCode, req.Country,
		req.Directions, req.Capacity, req.Timezone, req.IsVirtual,
	).Scan(
		&venue.ID, &venue.Name, &address, &city, &state, &postalCode, &country,
		&directions, &capacity, &venue.Timezone, &venue.IsVirtual, &venue.CreatedAt,
	)

	if err == sql.ErrNoRows {
		RespondError(w, http.StatusNotFound, ErrCodeNotFound, "Venue not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, ErrCodeInternalServer, "Failed to update venue")
		return
	}

	// Convert nullable fields
	if address.Valid {
		venue.Address = &address.String
	}
	if city.Valid {
		venue.City = &city.String
	}
	if state.Valid {
		venue.State = &state.String
	}
	if postalCode.Valid {
		venue.PostalCode = &postalCode.String
	}
	if country.Valid {
		venue.Country = &country.String
	}
	if directions.Valid {
		venue.Directions = &directions.String
	}
	if capacity.Valid {
		cap := int(capacity.Int64)
		venue.Capacity = &cap
	}

	RespondSuccess(w, http.StatusOK, venue)
}
