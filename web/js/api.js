// API Client for Taikai
const API_BASE = window.location.origin + '/api/v1';

const api = {
    // Helper to get auth headers
    getHeaders() {
        const token = localStorage.getItem('token');
        const headers = {
            'Content-Type': 'application/json'
        };
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }
        return headers;
    },

    // Helper for API requests
    async request(endpoint, options = {}) {
        const url = `${API_BASE}${endpoint}`;
        const config = {
            ...options,
            headers: {
                ...this.getHeaders(),
                ...options.headers
            }
        };

        const response = await fetch(url, config);

        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.error?.message || `HTTP ${response.status}`);
        }

        const data = await response.json();
        return data.data || data;
    },

    // Auth
    async login(email, password) {
        const data = await this.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });
        // Backend returns access_token and refresh_token
        if (data.access_token) {
            localStorage.setItem('token', data.access_token);
            localStorage.setItem('refresh_token', data.refresh_token);
            if (data.user) {
                localStorage.setItem('user', JSON.stringify(data.user));
            }
        }
        return data;
    },

    async register(userData) {
        const data = await this.request('/auth/register', {
            method: 'POST',
            body: JSON.stringify(userData)
        });
        // Backend returns access_token and refresh_token
        if (data.access_token) {
            localStorage.setItem('token', data.access_token);
            localStorage.setItem('refresh_token', data.refresh_token);
            if (data.user) {
                localStorage.setItem('user', JSON.stringify(data.user));
            }
        }
        return data;
    },

    async logout() {
        try {
            await this.request('/auth/logout', { method: 'POST' });
        } finally {
            localStorage.removeItem('token');
            localStorage.removeItem('refresh_token');
            localStorage.removeItem('user');
        }
    },

    // User
    async getMe() {
        return this.request('/users/me');
    },

    async updateMe(data) {
        return this.request('/users/me', {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    // Organizations
    async getOrganizations() {
        return this.request('/organizations');
    },

    async getOrganization(id) {
        return this.request(`/organizations/${id}`);
    },

    async createOrganization(data) {
        return this.request('/organizations', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateOrganization(id, data) {
        return this.request(`/organizations/${id}`, {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    // Groups
    async getGroups(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request(`/groups${query ? '?' + query : ''}`);
    },

    async getGroup(id) {
        return this.request(`/groups/${id}`);
    },

    async createGroup(data) {
        return this.request('/groups', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateGroup(id, data) {
        return this.request(`/groups/${id}`, {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    // Venues
    async getVenues() {
        return this.request('/venues');
    },

    async getVenue(id) {
        return this.request(`/venues/${id}`);
    },

    async createVenue(data) {
        return this.request('/venues', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateVenue(id, data) {
        return this.request(`/venues/${id}`, {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    // Events
    async getEvents(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request(`/events${query ? '?' + query : ''}`);
    },

    async getEvent(id) {
        return this.request(`/events/${id}`);
    },

    async createEvent(data) {
        return this.request('/events', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateEvent(id, data) {
        return this.request(`/events/${id}`, {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    async deleteEvent(id) {
        return this.request(`/events/${id}`, {
            method: 'DELETE'
        });
    },

    // Event Speakers
    async getEventSpeakers(eventId) {
        return this.request(`/events/${eventId}/speakers`);
    },

    async createEventSpeaker(eventId, data) {
        return this.request(`/events/${eventId}/speakers`, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateEventSpeaker(eventId, speakerId, data) {
        return this.request(`/events/${eventId}/speakers/${speakerId}`, {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    async deleteEventSpeaker(eventId, speakerId) {
        return this.request(`/events/${eventId}/speakers/${speakerId}`, {
            method: 'DELETE'
        });
    },

    // Event Sponsors
    async getEventSponsors(eventId) {
        return this.request(`/events/${eventId}/sponsors`);
    },

    async createEventSponsor(eventId, data) {
        return this.request(`/events/${eventId}/sponsors`, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateEventSponsor(eventId, sponsorId, data) {
        return this.request(`/events/${eventId}/sponsors/${sponsorId}`, {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    async deleteEventSponsor(eventId, sponsorId) {
        return this.request(`/events/${eventId}/sponsors/${sponsorId}`, {
            method: 'DELETE'
        });
    },

    // Event Schedule
    async getEventSchedule(eventId) {
        return this.request(`/events/${eventId}/schedule`);
    },

    async createScheduleItem(eventId, data) {
        return this.request(`/events/${eventId}/schedule`, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateScheduleItem(eventId, scheduleId, data) {
        return this.request(`/events/${eventId}/schedule/${scheduleId}`, {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    async deleteScheduleItem(eventId, scheduleId) {
        return this.request(`/events/${eventId}/schedule/${scheduleId}`, {
            method: 'DELETE'
        });
    },

    // RSVPs
    async getEventRSVPs(eventId) {
        return this.request(`/events/${eventId}/rsvps`);
    },

    async createOrUpdateRSVP(eventId, status) {
        return this.request(`/events/${eventId}/rsvp`, {
            method: 'POST',
            body: JSON.stringify({ status })
        });
    },

    async deleteRSVP(eventId) {
        return this.request(`/events/${eventId}/rsvp`, {
            method: 'DELETE'
        });
    },

    async getMyRSVPs() {
        return this.request('/me/rsvps');
    },

    // Subscriptions
    async subscribeToOrganization(orgId, preferences = {}) {
        return this.request(`/organizations/${orgId}/subscribe`, {
            method: 'POST',
            body: JSON.stringify(preferences)
        });
    },

    async unsubscribeFromOrganization(orgId) {
        return this.request(`/organizations/${orgId}/subscribe`, {
            method: 'DELETE'
        });
    },

    async subscribeToGroup(groupId, preferences = {}) {
        return this.request(`/groups/${groupId}/subscribe`, {
            method: 'POST',
            body: JSON.stringify(preferences)
        });
    },

    async unsubscribeFromGroup(groupId) {
        return this.request(`/groups/${groupId}/subscribe`, {
            method: 'DELETE'
        });
    },

    async getMySubscriptions() {
        return this.request('/me/subscriptions');
    },

    // File Upload
    async uploadImage(formData) {
        const token = localStorage.getItem('token');
        const response = await fetch(`${API_BASE}/upload/image`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData // Don't set Content-Type for FormData
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.error?.message || `HTTP ${response.status}`);
        }

        const data = await response.json();
        return data.data || data;
    }
};

// Make api globally available
window.api = api;
