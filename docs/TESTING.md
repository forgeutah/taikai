# Testing Taikai API Endpoints

This guide provides curl commands to test all API endpoints for organizations, groups, and users.

## Prerequisites

1. **Start the development environment:**
   ```bash
   make setup  # First time only
   make dev    # Start server on port 8080
   ```

2. **Seed the database with test data:**
   ```bash
   make seed
   ```
   This creates:
   - 1 organization (Forge Utah Foundation)
   - 3 groups (Kubernetes Meetup, Go User Group, Data Engineering)
   - 1 admin user (admin@forgeutah.org / password123)
   - 3 test users

3. **Set environment variables for testing:**
   ```bash
   export API_URL="http://localhost:8080/api/v1"
   ```

## Authentication Flow

### 1. Register a New User

```bash
curl -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "name": "Test User",
    "timezone": "America/Denver"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "message": "Registration successful. Please check your email to verify your account.",
    "user_id": "uuid-here"
  }
}
```

### 2. Login

```bash
curl -X POST $API_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@forgeutah.org",
    "password": "password123"
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user": {
      "id": "uuid",
      "email": "admin@forgeutah.org",
      "name": "Admin User"
    }
  }
}
```

**Save the access token for subsequent requests:**
```bash
export TOKEN="your-access-token-here"
```

### 3. Get Current User Profile

```bash
curl -X GET $API_URL/users/me \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Update User Profile

```bash
curl -X PATCH $API_URL/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "bio": "Software developer and community organizer",
    "timezone": "America/Los_Angeles"
  }'
```

### 5. Refresh Access Token

```bash
curl -X POST $API_URL/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your-refresh-token-here"
  }'
```

### 6. Logout

```bash
curl -X POST $API_URL/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

---

## Organization Endpoints

### 1. List All Organizations (Public)

```bash
curl -X GET $API_URL/organizations
```

**Expected Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Forge Utah Foundation",
      "slug": "forge-utah-foundation",
      "description": "Utah's nonprofit tech community organization",
      "logo_url": "https://forgeutah.org/logo.png",
      "website_url": "https://forgeutah.org",
      "community_chat_url": "https://discord.gg/forgeutah",
      "community_chat_name": "Discord",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 2. Get Organization by ID (Public)

```bash
curl -X GET $API_URL/organizations/{orgId}
```

Example:
```bash
# Get the org ID from the list above, then:
export ORG_ID="your-org-id-here"
curl -X GET $API_URL/organizations/$ORG_ID
```

### 3. Get Organization by Slug (Public)

```bash
curl -X GET $API_URL/organizations/slug/forge-utah-foundation
```

### 4. Get Organization Admins (Public)

```bash
curl -X GET $API_URL/organizations/$ORG_ID/admins
```

**Expected Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "email": "admin@forgeutah.org",
      "name": "Admin User",
      "avatar_url": null,
      "assigned_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 5. Update Organization (Org Admin Only)

```bash
curl -X PATCH $API_URL/organizations/$ORG_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Org Name",
    "description": "Updated description",
    "website_url": "https://newdomain.com"
  }'
```

**Note:** You must be an org admin for this organization.

### 6. Add Organization Admin (Org Admin Only)

First, get a user ID to add as admin:
```bash
# Login as a different user or use an existing user ID
export USER_ID="user-id-to-add"

curl -X POST $API_URL/organizations/$ORG_ID/admins \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'$USER_ID'"
  }'
```

### 7. Remove Organization Admin (Org Admin Only)

```bash
curl -X DELETE $API_URL/organizations/$ORG_ID/admins/$USER_ID \
  -H "Authorization: Bearer $TOKEN"
```

---

## Group Endpoints

### 1. List All Groups (Public, with Pagination)

```bash
# List all groups
curl -X GET $API_URL/groups

# With pagination
curl -X GET "$API_URL/groups?page=1&limit=10"

# Filter by organization
curl -X GET "$API_URL/groups?organization_id=$ORG_ID"

# Combined
curl -X GET "$API_URL/groups?organization_id=$ORG_ID&page=1&limit=5"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "groups": [
      {
        "id": "uuid",
        "organization_id": "uuid",
        "name": "Kubernetes Meetup",
        "slug": "kubernetes-meetup",
        "description": "Monthly meetup for Kubernetes enthusiasts",
        "logo_url": null,
        "website_url": null,
        "community_chat_url": "https://discord.gg/forgeutah",
        "community_chat_name": "Discord",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 3,
      "total_pages": 1
    }
  }
}
```

### 2. Get Group by ID (Public)

```bash
export GROUP_ID="your-group-id-here"
curl -X GET $API_URL/groups/$GROUP_ID
```

### 3. Get Group by Slug (Public)

```bash
curl -X GET $API_URL/groups/slug/kubernetes-meetup
```

### 4. Get Group Admins (Public)

```bash
curl -X GET $API_URL/groups/$GROUP_ID/admins
```

### 5. Create Group (Org Admin Only)

```bash
curl -X POST $API_URL/groups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "'$ORG_ID'",
    "name": "New Tech Group",
    "description": "A brand new tech community group",
    "website_url": "https://example.com"
  }'
```

**Notes:**
- You must be an org admin for the organization
- Slug is auto-generated from the name
- Community chat settings are inherited from organization (unless overridden)
- Creator is automatically added as group admin

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "id": "new-uuid",
    "organization_id": "org-uuid",
    "name": "New Tech Group",
    "slug": "new-tech-group",
    "description": "A brand new tech community group",
    "logo_url": null,
    "website_url": "https://example.com",
    "community_chat_url": "https://discord.gg/forgeutah",
    "community_chat_name": "Discord",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### 6. Update Group (Group Admin or Org Admin)

```bash
curl -X PATCH $API_URL/groups/$GROUP_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Group Name",
    "description": "Updated description",
    "community_chat_url": "https://slack.com/customgroup",
    "community_chat_name": "Slack"
  }'
```

**Note:** You must be either:
- A group admin for this group, OR
- An org admin for the parent organization

### 7. Delete Group (Org Admin Only)

```bash
curl -X DELETE $API_URL/groups/$GROUP_ID \
  -H "Authorization: Bearer $TOKEN"
```

**Note:** Only org admins can delete groups (not group admins).

### 8. Add Group Admin (Group Admin or Org Admin)

```bash
curl -X POST $API_URL/groups/$GROUP_ID/admins \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "'$USER_ID'"
  }'
```

### 9. Remove Group Admin (Group Admin or Org Admin)

```bash
curl -X DELETE $API_URL/groups/$GROUP_ID/admins/$USER_ID \
  -H "Authorization: Bearer $TOKEN"
```

---

## Complete Testing Workflow

Here's a complete workflow to test the organization and group management features:

### Step 1: Setup and Login

```bash
# Set API URL
export API_URL="http://localhost:8080/api/v1"

# Login as admin
curl -X POST $API_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@forgeutah.org",
    "password": "password123"
  }' | jq -r '.data.access_token' > token.txt

export TOKEN=$(cat token.txt)
```

### Step 2: Explore Organizations

```bash
# List all organizations
curl -X GET $API_URL/organizations | jq

# Get first organization ID
export ORG_ID=$(curl -s $API_URL/organizations | jq -r '.data[0].id')
echo "Organization ID: $ORG_ID"

# Get organization details
curl -X GET $API_URL/organizations/$ORG_ID | jq

# Get by slug
curl -X GET $API_URL/organizations/slug/forge-utah-foundation | jq

# View org admins
curl -X GET $API_URL/organizations/$ORG_ID/admins | jq
```

### Step 3: Manage Groups

```bash
# List all groups
curl -X GET $API_URL/groups | jq

# List groups for organization
curl -X GET "$API_URL/groups?organization_id=$ORG_ID" | jq

# Get first group ID
export GROUP_ID=$(curl -s $API_URL/groups | jq -r '.data.groups[0].id')
echo "Group ID: $GROUP_ID"

# Get group details
curl -X GET $API_URL/groups/$GROUP_ID | jq

# View group admins
curl -X GET $API_URL/groups/$GROUP_ID/admins | jq
```

### Step 4: Create a New Group

```bash
# Create a new group
curl -X POST $API_URL/groups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": "'$ORG_ID'",
    "name": "Rust Programming Group",
    "description": "Learn and discuss Rust programming language",
    "website_url": "https://rust-slc.com"
  }' | jq

# Get the new group ID
export NEW_GROUP_ID=$(curl -s "$API_URL/groups?organization_id=$ORG_ID" | jq -r '.data.groups[] | select(.name == "Rust Programming Group") | .id')
echo "New Group ID: $NEW_GROUP_ID"
```

### Step 5: Update the Group

```bash
curl -X PATCH $API_URL/groups/$NEW_GROUP_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Monthly meetup for Rust developers in Salt Lake City",
    "logo_url": "https://example.com/rust-logo.png"
  }' | jq
```

### Step 6: Register Another User and Add as Group Admin

```bash
# Register new user
curl -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rust-admin@example.com",
    "password": "SecurePass123!",
    "name": "Rust Admin",
    "timezone": "America/Denver"
  }' | jq

# Note: In production, user would need to verify email first
# For testing, manually get the user ID from database or response

# Add as group admin (using org admin token)
curl -X POST $API_URL/groups/$NEW_GROUP_ID/admins \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "new-user-id-here"
  }' | jq

# View updated group admins
curl -X GET $API_URL/groups/$NEW_GROUP_ID/admins | jq
```

### Step 7: Test Permissions

```bash
# Try to update org as non-admin (should fail)
# First, login as regular user
curl -X POST $API_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@example.com",
    "password": "password123"
  }' | jq -r '.data.access_token' > user_token.txt

export USER_TOKEN=$(cat user_token.txt)

# Try to update organization (should return 403 Forbidden)
curl -X PATCH $API_URL/organizations/$ORG_ID \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Trying to hack"
  }'

# Expected: {"success":false,"error":{"code":"forbidden","message":"organization admin access required"}}
```

### Step 8: Delete the Test Group (Optional)

```bash
# Delete group (org admin only)
curl -X DELETE $API_URL/groups/$NEW_GROUP_ID \
  -H "Authorization: Bearer $TOKEN"

# Verify deletion
curl -X GET $API_URL/groups/$NEW_GROUP_ID
# Expected: 404 Not Found
```

---

## Testing with Postman

If you prefer using Postman:

1. **Import this collection:**
   - Create a new Postman collection
   - Set a collection variable: `baseUrl` = `http://localhost:8080/api/v1`
   - Set a collection variable: `token` = (will be set after login)

2. **Add requests:**
   - Copy the curl commands above
   - Import each curl command using "Import" → "Raw text"
   - Or manually create requests following the patterns above

3. **Setup authentication:**
   - After login, save the access token to `{{token}}`
   - Use Bearer Token authentication type
   - Token: `{{token}}`

4. **Create test scripts:**
   ```javascript
   // For login request:
   pm.test("Status is 200", function () {
       pm.response.to.have.status(200);
   });

   pm.test("Save token", function () {
       var data = pm.response.json();
       pm.collectionVariables.set("token", data.data.access_token);
   });
   ```

---

## Common HTTP Status Codes

- **200 OK**: Request succeeded
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid request body or parameters
- **401 Unauthorized**: Missing or invalid authentication token
- **403 Forbidden**: Authenticated but not authorized (permission denied)
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource already exists (e.g., duplicate email)
- **500 Internal Server Error**: Server error (check logs)

---

## Troubleshooting

### "Connection refused"
```bash
# Check if server is running
curl http://localhost:8080/health

# If not, start it:
make dev
```

### "Unauthorized" errors
```bash
# Check if token is valid
echo $TOKEN

# If empty or expired, login again
curl -X POST $API_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@forgeutah.org","password":"password123"}' \
  | jq -r '.data.access_token'

export TOKEN="new-token-here"
```

### "Organization/Group not found"
```bash
# Make sure database is seeded
make seed

# List available organizations
curl -X GET $API_URL/organizations | jq

# List available groups
curl -X GET $API_URL/groups | jq
```

### "Permission denied" (403)
```bash
# Check which user you're logged in as
curl -X GET $API_URL/users/me \
  -H "Authorization: Bearer $TOKEN" | jq

# Check if user is org admin
curl -X GET $API_URL/organizations/$ORG_ID/admins | jq

# Login as admin if needed
export TOKEN=$(curl -s -X POST $API_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@forgeutah.org","password":"password123"}' \
  | jq -r '.data.access_token')
```

---

## Next Steps

- See [API.md](./API.md) for complete API documentation
- See [DEPLOYMENT.md](./DEPLOYMENT.md) for production deployment
- See [SELF-HOSTING.md](./SELF-HOSTING.md) for self-hosting guide

## Questions?

- Check server logs: `docker-compose logs -f server`
- Check database: `make db-console`
- Review [docs/API.md](./API.md) for detailed API specs
