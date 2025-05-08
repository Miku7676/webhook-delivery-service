# Stage 1: Build binaries
FROM golang:1.24-alpine AS builder

# Install required packages
RUN apk add --no-cache git gcc musl-dev

# Install swag for Swagger docs generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Swagger docs
RUN swag init --dir ./cmd/api,./handlers,./models --output ./docs

# Build API binary
RUN go build -o /api ./cmd/api

# Build Worker binary
RUN go build -o /worker ./cmd/worker

# Stage 2: Run
FROM alpine:latest

RUN apk add --no-cache ca-certificates postgresql-client bash

WORKDIR /app

COPY --from=builder /api /api
COPY --from=builder /worker /worker

# Copy Swagger docs
COPY --from=builder /app/docs /docs

# Copy wait-for-postgres.sh
COPY wait-for-db.sh /app/wait-for-db.sh
RUN chmod +x /app/wait-for-db.sh

EXPOSE 8080

CMD ["/api"]
