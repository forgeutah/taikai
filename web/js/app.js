// Main app JavaScript

// Check authentication on all pages
function checkAuth() {
    const token = localStorage.getItem('token');
    return !!token;
}

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

// Error handling
function handleError(error) {
    console.error('Error:', error);
    const message = error.message || 'An error occurred';

    // Show error to user
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-toast';
    errorDiv.textContent = message;
    document.body.appendChild(errorDiv);

    setTimeout(() => {
        errorDiv.remove();
    }, 5000);
}

// Success message
function showSuccess(message) {
    const successDiv = document.createElement('div');
    successDiv.className = 'success-toast';
    successDiv.textContent = message;
    document.body.appendChild(successDiv);

    setTimeout(() => {
        successDiv.remove();
    }, 3000);
}

// Export for use in other scripts
window.app = {
    checkAuth,
    formatDate,
    formatTime,
    handleError,
    showSuccess
};
