# API Documentation

## Base URL

```text
http://localhost:8080
```

## Roles

- `admin`: Can manage users and also use the normal self-service/email operations.
- `user`: Can update their own settings and use email operations.

The first registered account becomes `admin`. All later registrations become `user`.

## Authentication

### Register

Create a new user account.

**Endpoint:** `POST /register`

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "smtpHost": "smtp.gmail.com",
  "smtpPort": 587,
  "smtpUser": "user@example.com",
  "smtpPassword": "app-password",
  "fromEmail": "user@example.com",
  "useTLS": true
}
```

**Response:**

```json
{
  "token": "jwt-token-here",
  "apiKey": "your-api-key-here",
  "user": {
    "id": "user-id",
    "email": "user@example.com",
    "role": "user",
    "emailLogTTLDays": 30,
    "createdAt": "2026-01-20T12:00:00Z"
  }
}
```

### Login

Authenticate and get JWT token.

**Endpoint:** `POST /login`

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**

```json
{
  "token": "jwt-token-here",
  "apiKey": "your-api-key-here",
  "user": {
    "id": "user-id",
    "email": "user@example.com",
    "role": "user"
  }
}
```

## User Endpoints

These endpoints are available to authenticated `user` and `admin` accounts for their own data.

### Update Email Log TTL

Update how long your email logs are retained.

**Endpoint:** `PUT /user/settings/email-ttl`

**Headers:**

```text
Authorization: Bearer <jwt-token>
```

**Request Body:**

```json
{
  "emailLogTTLDays": 60
}
```

Set `0` to never delete logs.

**Response:**

```json
{
  "success": true,
  "message": "Email log TTL updated successfully"
}
```

### Get Email Logs

Retrieve your own email sending history.

**Endpoint:** `GET /user/email-logs`

**Headers:**

```text
Authorization: Bearer <jwt-token>
```

**Response:**

```json
{
  "success": true,
  "count": 10,
  "logs": [
    {
      "id": "log-id",
      "userId": "user-id",
      "to": "recipient@example.com",
      "subject": "Test Email",
      "body": "Email content",
      "status": "success",
      "errorMessage": "",
      "sentAt": "2026-01-20T12:30:00Z"
    }
  ]
}
```

## Admin Endpoints

These endpoints require an authenticated `admin` account.

### List Users

**Endpoint:** `GET /admin/users`

**Headers:**

```text
Authorization: Bearer <admin-jwt-token>
```

### Get User

**Endpoint:** `GET /admin/users/{id}`

**Headers:**

```text
Authorization: Bearer <admin-jwt-token>
```

### Update User

Update a user's role and/or email log TTL.

**Endpoint:** `PUT /admin/users/{id}`

**Headers:**

```text
Authorization: Bearer <admin-jwt-token>
Content-Type: application/json
```

**Request Body:**

```json
{
  "role": "admin",
  "emailLogTTLDays": 90
}
```

### Delete User

**Endpoint:** `DELETE /admin/users/{id}`

**Headers:**

```text
Authorization: Bearer <admin-jwt-token>
```

## Webhook Endpoints

### Send Email

Send an email via webhook using the caller's API key.

**Endpoint:** `POST /webhook/send-email`

**Headers:**

```text
x-api-key: your-api-key-here
Content-Type: application/json
```

**Request Body:**

```json
{
  "to": "recipient@example.com",
  "subject": "Hello World",
  "body": "This is the email body",
  "isHTML": false
}
```

**Response:**

```json
{
  "success": true,
  "message": "Email queued for sending",
  "logId": "log-id-here"
}
```

## Error Responses

All endpoints may return plain-text error responses.

## Common Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `202 Accepted` - Request accepted (async processing)
- `400 Bad Request` - Invalid request body or parameters
- `401 Unauthorized` - Authentication failed
- `403 Forbidden` - Role does not have access
- `404 Not Found` - User not found
- `500 Internal Server Error` - Server error
