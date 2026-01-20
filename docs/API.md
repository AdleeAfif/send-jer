# API Documentation

## Base URL

```
http://localhost:8080
```

## Authentication

### Register

Create a new user account.

**Endpoint:** `POST /register`

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "smtpHost": "smtp.gmail.com", // Optional
  "smtpPort": 587, // Optional
  "smtpUser": "user@example.com", // Optional
  "smtpPassword": "app-password", // Optional
  "fromEmail": "user@example.com", // Optional
  "useTLS": true // Optional
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
  "user": { ... }
}
```

## User Endpoints

### Update Email Log TTL

Update how long email logs are retained.

**Endpoint:** `PUT /user/settings/email-ttl`

**Headers:**

```
Authorization: Bearer <jwt-token>
```

**Request Body:**

```json
{
  "emailLogTTLDays": 60
}
```

_Note: Set to 0 to never delete logs_

**Response:**

```json
{
  "success": true,
  "message": "Email log TTL updated successfully"
}
```

### Get Email Logs

Retrieve your email sending history.

**Endpoint:** `GET /user/email-logs`

**Headers:**

```
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

## Webhook Endpoints

### Send Email

Send an email via webhook.

**Endpoint:** `POST /webhook/send-email`

**Headers:**

```
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

All endpoints may return error responses in this format:

**Status:** 400, 401, 500, etc.

**Body:**

```
Error message as plain text
```

## Common Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `202 Accepted` - Request accepted (async processing)
- `400 Bad Request` - Invalid request body or parameters
- `401 Unauthorized` - Authentication failed
- `500 Internal Server Error` - Server error

## Rate Limiting

Currently no rate limiting is implemented. Consider adding rate limiting in production.

## SMTP Configuration

### Using Default SMTP

If you don't provide SMTP credentials during registration, the service will use the default SMTP configured in `.env`.

### Using Custom SMTP

Provide your SMTP credentials during registration. The password will be encrypted before storage.

### Gmail Example

For Gmail, you need to:

1. Enable 2FA
2. Generate an App Password
3. Use `smtp.gmail.com` with port `587`

## Security Notes

1. **HTTPS**: Always use HTTPS in production
2. **API Keys**: Keep your API key secure - it's like a password
3. **JWT Tokens**: Tokens expire after 24 hours
4. **SMTP Passwords**: Encrypted using AES-256-GCM
5. **Environment Variables**: Never commit `.env` to version control
