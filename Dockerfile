# ---- build stage ----
FROM golang:1.26.6-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

# ---- run stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app
USER app
WORKDIR /app

COPY --from=builder /bin/api /app/api

EXPOSE 8080
CMD ["/app/api"]