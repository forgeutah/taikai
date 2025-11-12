### Taikai API Documentation

**Base URL:** `http://localhost:8080/api/v1`

---

## Authentication Endpoints

### 1. Register

Create a new user account.

**Endpoint:** `POST /auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe",
  "timezone": "America/Denver"
}
```

**Success Response (201 Created):**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "avatar_url": null,
      "timezone": "America/Denver",
      "email_verified": false,
      "created_at": "2025-01-01T00:00:00Z"
    },
    "message": "Registration successful. Please check your email to verify your account."
  }
}
```

**Validation:**
- Email must be valid format
- Password must be at least 8 characters
- Name is required
- Timezone defaults to "America/Denver" if not provided

---

### 2. Login

Authenticate and receive JWT tokens.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "avatar_url": null,
      "timezone": "America/Denver",
      "email_verified": true,
      "created_at": "2025-01-01T00:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 900
  }
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid email or password"
  }
}
```

---

### 3. Logout

Invalidate refresh token (blacklist it).

**Endpoint:** `POST /auth/logout`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "Logged out successfully"
  }
}
```

---

### 4. Refresh Token

Get a new access token using refresh token.

**Endpoint:** `POST /auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 900
  }
}
```

---

### 5. Verify Email

Verify email address using token from email.

**Endpoint:** `GET /auth/verify-email?token={token}`

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "Email verified successfully"
  }
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid or expired verification token"
  }
}
```

---

### 6. Forgot Password

Request password reset email.

**Endpoint:** `POST /auth/forgot-password`

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "If the email exists, a password reset link has been sent"
  }
}
```

**Note:** Returns success even if email doesn't exist (security best practice).

---

### 7. Reset Password

Reset password using token from email.

**Endpoint:** `POST /auth/reset-password`

**Request Body:**
```json
{
  "token": "reset-token-from-email",
  "password": "NewSecurePass123!"
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "Password reset successfully"
  }
}
```

---

## User Profile Endpoints

All user endpoints require authentication via Bearer token.

### 8. Get Current User

Get the authenticated user's profile.

**Endpoint:** `GET /users/me`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "avatar_url": null,
    "timezone": "America/Denver",
    "email_verified": true,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### 9. Update Current User

Update the authenticated user's profile.

**Endpoint:** `PATCH /users/me`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Request Body:** (all fields optional)
```json
{
  "name": "Jane Doe",
  "bio": "Software engineer and community organizer",
  "phone": "+1234567890",
  "timezone": "America/New_York",
  "discord_id": "user#1234"
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "Jane Doe",
    "avatar_url": null,
    "timezone": "America/New_York",
    "email_verified": true,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### 10. Delete Current User

Soft delete the authenticated user's account.

**Endpoint:** `DELETE /users/me`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "message": "Account deleted successfully"
  }
}
```

---

## Error Responses

All endpoints may return the following error codes:

### 400 Bad Request
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request body"
  }
}
```

### 401 Unauthorized
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing authorization header"
  }
}
```

### 404 Not Found
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

### 409 Conflict
```json
{
  "error": {
    "code": "CONFLICT",
    "message": "Email already registered"
  }
}
```

### 500 Internal Server Error
```json
{
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "An unexpected error occurred"
  }
}
```

---

## Authentication Flow

### Registration Flow
1. User submits registration form → `POST /auth/register`
2. System creates user account (email_verified = false)
3. System sends verification email with token
4. User clicks link in email → `GET /auth/verify-email?token=...`
5. System sets email_verified = true
6. User can now login

### Login Flow
1. User submits login credentials → `POST /auth/login`
2. System validates credentials
3. System returns access_token (15 min) and refresh_token (7 days)
4. Client stores both tokens securely
5. Client uses access_token for API requests

### Token Refresh Flow
1. Access token expires
2. Client sends refresh_token → `POST /auth/refresh`
3. System validates refresh_token
4. System returns new access_token
5. Client uses new access_token

### Password Reset Flow
1. User requests reset → `POST /auth/forgot-password`
2. System sends reset email with token
3. User clicks link and enters new password → `POST /auth/reset-password`
4. System updates password
5. User can login with new password

---

## Rate Limiting

*(To be implemented in future)*

- Registration: 5 requests per hour per IP
- Login: 10 requests per 5 minutes per IP
- Password reset: 3 requests per hour per email

---

## Notes for SSO/OAuth Integration (Phase 2)

The authentication system is designed to support OAuth providers:

- OAuth accounts table already exists in the database
- Password hash is nullable (supports OAuth-only users)
- Future endpoints will include:
  - `GET /auth/oauth/google` - Initiate Google OAuth
  - `GET /auth/oauth/google/callback` - Handle Google callback
  - `GET /auth/oauth/github` - Initiate GitHub OAuth
  - `GET /auth/oauth/github/callback` - Handle GitHub callback

OAuth users will receive the same JWT tokens as password-based users, ensuring a unified authentication experience.
