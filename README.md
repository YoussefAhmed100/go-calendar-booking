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

## Status

Work in progress.