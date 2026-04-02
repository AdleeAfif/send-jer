# Send-Jer - Email Webhook Service

A production-ready email webhook service built with Go and MongoDB.

## Features

- Send emails via webhook API
- JWT authentication for dashboard/API
- API key authentication for webhooks
- Role-based access with `admin` and `user`
- Comprehensive email logging
- Configurable log retention (TTL)
- Async email sending with workers
- Automatic cleanup of old logs
- Encrypted SMTP credentials
- Support for custom or default SMTP

## Quick Start

1. **Setup Environment**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Install Dependencies**

   ```bash
   go mod download
   ```

3. **Run the Server**

   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints

### Authentication

- `POST /register` - Register a new account. The first account becomes `admin`; later accounts become `user`.
- `POST /login` - Login and get JWT token

### User Settings

- `PUT /user/settings/email-ttl` - Update your own email log retention days

### Email Operations

- `POST /webhook/send-email` - Send email via webhook (API key auth)
- `GET /user/email-logs` - View your own email logs (JWT auth)

### Admin Management

- `GET /admin/users` - List all users
- `GET /admin/users/{id}` - View a specific user
- `PUT /admin/users/{id}` - Update a user's role or email log TTL
- `DELETE /admin/users/{id}` - Delete a user account

## Architecture

- **API Layer**: HTTP handlers for all endpoints
- **Service Layer**: Business logic for users and emails
- **Worker**: Async email sending with queue
- **Cron Jobs**: Automatic cleanup of old logs
- **Encryption**: AES-256 for SMTP credentials

## Configuration

See `.env.example` for all available configuration options.

## License

MIT
