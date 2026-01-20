# Deployment Guide

## Prerequisites

- Go 1.21 or later
- MongoDB 4.4 or later
- SMTP server access (Gmail, SendGrid, etc.)

## Environment Setup

1. **Create `.env` file**

   ```bash
   cp .env.example .env
   ```

2. **Configure Environment Variables**

   Edit `.env` and set:
   - MongoDB connection string
   - JWT secret (use a strong random string)
   - Encryption key (must be exactly 32 bytes for AES-256)
   - Default SMTP credentials
   - Server port

## Local Development

1. **Install Dependencies**

   ```bash
   go mod download
   ```

2. **Run MongoDB**

   ```bash
   # Using Docker
   docker run -d -p 27017:27017 --name mongodb mongo:latest

   # Or use local MongoDB installation
   ```

3. **Start the Server**

   ```bash
   go run cmd/server/main.go
   ```

4. **Test the API**
   ```bash
   curl http://localhost:8080/health
   ```

## Production Deployment

### Option 1: Binary Deployment

1. **Build the Binary**

   ```bash
   # For Linux
   GOOS=linux GOARCH=amd64 go build -o send-jer cmd/server/main.go

   # For Windows
   GOOS=windows GOARCH=amd64 go build -o send-jer.exe cmd/server/main.go
   ```

2. **Transfer to Server**

   ```bash
   scp send-jer user@server:/path/to/app/
   scp .env user@server:/path/to/app/
   ```

3. **Run with systemd**

   Create `/etc/systemd/system/send-jer.service`:

   ```ini
   [Unit]
   Description=Send-Jer Email Webhook Service
   After=network.target

   [Service]
   Type=simple
   User=www-data
   WorkingDirectory=/path/to/app
   ExecStart=/path/to/app/send-jer
   Restart=always
   RestartSec=10

   [Install]
   WantedBy=multi-user.target
   ```

   Enable and start:

   ```bash
   sudo systemctl enable send-jer
   sudo systemctl start send-jer
   sudo systemctl status send-jer
   ```

### Option 2: Docker Deployment

1. **Create Dockerfile**

   ```dockerfile
   FROM golang:1.21-alpine AS builder
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN go build -o send-jer cmd/server/main.go

   FROM alpine:latest
   RUN apk --no-cache add ca-certificates
   WORKDIR /root/
   COPY --from=builder /app/send-jer .
   COPY .env .
   EXPOSE 8080
   CMD ["./send-jer"]
   ```

2. **Build and Run**
   ```bash
   docker build -t send-jer .
   docker run -d -p 8080:8080 --name send-jer send-jer
   ```

### Option 3: Docker Compose

Create `docker-compose.yml`:

```yaml
version: "3.8"

services:
  mongodb:
    image: mongo:latest
    restart: always
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db

  send-jer:
    build: .
    restart: always
    ports:
      - "8080:8080"
    depends_on:
      - mongodb
    environment:
      - MONGODB_URI=mongodb://mongodb:27017
      - PORT=8080
    env_file:
      - .env

volumes:
  mongodb_data:
```

Run with:

```bash
docker-compose up -d
```

## Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## SSL with Let's Encrypt

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d api.yourdomain.com
```

## Monitoring

### Logs

```bash
# View logs
journalctl -u send-jer -f

# Docker logs
docker logs -f send-jer
```

### Health Check

```bash
curl https://api.yourdomain.com/health
```

## Security Checklist

- [ ] Use strong JWT secret
- [ ] Use 32-byte encryption key
- [ ] Enable HTTPS in production
- [ ] Secure MongoDB with authentication
- [ ] Use environment variables for secrets
- [ ] Set up firewall rules
- [ ] Enable rate limiting (consider adding middleware)
- [ ] Regular backups of MongoDB
- [ ] Monitor logs for suspicious activity

## Scaling

For high-traffic scenarios:

1. **Horizontal Scaling**: Run multiple instances behind a load balancer
2. **MongoDB Replica Set**: For high availability
3. **Redis Queue**: Replace channel-based queue with Redis
4. **Separate Workers**: Run workers on separate servers
5. **CDN**: For static assets if added later

## Backup

```bash
# Backup MongoDB
mongodump --uri="mongodb://localhost:27017/email_webhook_service" --out=/backup/$(date +%Y%m%d)

# Restore
mongorestore --uri="mongodb://localhost:27017/email_webhook_service" /backup/20260120
```

## Troubleshooting

### MongoDB Connection Issues

- Check MongoDB is running: `systemctl status mongod`
- Verify connection string in `.env`
- Check firewall rules

### Email Sending Failures

- Verify SMTP credentials
- Check SMTP server allows connections
- For Gmail: Use App Passwords, not regular password
- Check worker logs for specific errors

### High Memory Usage

- Monitor goroutine count
- Check for memory leaks in workers
- Adjust worker concurrency in `.env`
