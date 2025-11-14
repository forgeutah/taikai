// Main app JavaScript

// Authentication helper
const auth = {
    isAuthenticated() {
        const token = localStorage.getItem('token');
        return !!token;
    },

    getToken() {
        return localStorage.getItem('token');
    },

    async logout() {
        try {
            await api.logout();
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            localStorage.removeItem('token');
            localStorage.removeItem('refresh_token');
            localStorage.removeItem('user');
            window.location.href = '/';
        }
    },

    updateUI() {
        const isAuth = this.isAuthenticated();
        const loginBtn = document.getElementById('loginBtn');
        const logoutBtn = document.getElementById('logoutBtn');
        const profileBtn = document.getElementById('profileBtn');

        if (isAuth) {
            if (loginBtn) loginBtn.style.display = 'none';
            if (logoutBtn) logoutBtn.style.display = 'inline-block';
            if (profileBtn) profileBtn.style.display = 'inline-block';
        } else {
            if (loginBtn) loginBtn.style.display = 'inline-block';
            if (logoutBtn) logoutBtn.style.display = 'none';
            if (profileBtn) profileBtn.style.display = 'none';
        }
    }
};

// Format date helpers
function formatDate(dateString) {
    return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });
}

function formatTime(dateString) {
    return new Date(dateString).toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit'
    });
}

// Toast notifications
function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
        toast.remove();
    }, 3000);
}

// HTML escaping helper
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Error handling
function handleError(error) {
    console.error('Error:', error);
    const message = error.message || 'An error occurred';
    showToast(message, 'error');
}

// Success message
function showSuccess(message) {
    showToast(message, 'success');
}

// Make globally available
window.auth = auth;
window.showToast = showToast;
window.escapeHtml = escapeHtml;
window.handleError = handleError;
window.showSuccess = showSuccess;

// Export for use in other scripts
window.app = {
    auth,
    formatDate,
    formatTime,
    handleError,
    showSuccess,
    showToast,
    escapeHtml
};
