# Example Server

A Go-based API server with JWT authentication and Swagger documentation.

## Features

- JWT-based authentication
- Protected and public endpoints
- Swagger UI documentation
- Environment variable configuration
- Static file serving

## Setup

1. Make sure you have Go 1.21 or later installed
2. Clone this repository
3. Copy `.env.example` to `.env` and adjust the values as needed
4. Install dependencies:
   ```bash
   go mod download
   ```
5. Run the server:
   ```bash
   go run main.go
   ```

## API Documentation

Once the server is running, you can access the Swagger UI documentation at:
```
http://localhost:8080/public/
```

## Available Endpoints

- `POST /api/login` - Get JWT token (public)
- `GET /api/customers` - Get customers list (protected)

## Authentication

To access protected endpoints, include the JWT token in the Authorization header:
```
Authorization: Bearer <your-token>
```

## Environment Variables

- `PORT` - Server port (default: 8080)
- `JWT_SECRET` - Secret key for JWT signing
- `SWAGGER_HOST` - Host for Swagger documentation

## Datadog Setup

This project includes Datadog integration for logging and monitoring. To set up Datadog:

1. Install the Datadog Agent for Windows
2. Run the setup script as Administrator:
   ```powershell
   # Open PowerShell as Administrator
   cd <project-root>
   .\scripts\setup-datadog.ps1
   ```
3. Restart the Datadog Agent:
   ```powershell
   Restart-Service datadogagent
   ```

The setup script will:
- Create symbolic links from the Datadog Agent config directory to your project's config
- Set required environment variables
- Configure log collection

### Development

For local development, logs will be written to:
- Application logs: `./logs/app.log`
- Datadog logs: `./logs/datadog.log`

The Datadog Agent will automatically collect logs from these files when properly configured. 