# Go Calendar Booking

A small calendar booking API built with Go. Users register, connect a Google account,
choose a calendar, and create/cancel appointments that stay in sync with Google Calendar.

## Stack

- Go, Gin
- PostgreSQL, GORM
- JWT authentication
- Google Calendar API (OAuth2)

## Architecture

Modular monolith. Each module follows `model -> service -> controller` with DTOs.

```
internal/
├── common/      # config, database, event bus, middleware
└── modules/
    ├── auth/
    ├── google/
    └── appointment/
```

## Getting Started

```bash
cp .env.example .env      # then fill in the values
docker compose --env-file .env up -d
go run ./cmd/api
```

Health check: `GET http://localhost:8080/health`

## Authentication API

The API supports user registration, login, and logout using JWT-based authentication.

### Endpoints

| Method | Endpoint | Authentication | Description |
|---|---|---|---|
| POST | `/api/auth/register` | Not required | Register a new user |
| POST | `/api/auth/login` | Not required | Authenticate a user and issue a JWT |
| POST | `/api/auth/logout` | Bearer token required | Revoke the current session |

### Logout

Send the JWT in the `Authorization` header:

```http
Authorization: Bearer <your-jwt-token>
```

A successful logout revokes the associated session. Subsequent requests using the revoked token are rejected by the authentication middleware.

### Testing

Run the test suite from the project root:

```bash
go test ./...
```

Use Postman to verify registration, login, successful logout, and rejection of a revoked token.


## Status

Work in progress.