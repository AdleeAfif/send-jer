# Send-Jer - Email Webhook Service

A production-ready email webhook service built with Go and MongoDB.

## Features

- 📧 Send emails via webhook API
- 🔐 JWT authentication for dashboard/API
- 🔑 API key authentication for webhooks
- 📊 Comprehensive email logging
- ⏰ Configurable log retention (TTL)
- 🔄 Async email sending with workers
- 🧹 Automatic cleanup of old logs
- 🔒 Encrypted SMTP credentials
- 🎯 Support for custom or default SMTP

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

- `POST /register` - Register new user
- `POST /login` - Login and get JWT token

### User Settings

- `PUT /user/settings/email-ttl` - Update email log retention days

### Email Operations

- `POST /webhook/send-email` - Send email via webhook (API key auth)
- `GET /user/email-logs` - View email logs (JWT auth)

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
