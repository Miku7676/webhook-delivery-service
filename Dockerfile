# Stage 1: Build binaries
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build API binary
RUN go build -o /api ./cmd/api

# Build Worker binary
RUN go build -o /worker ./cmd/worker

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

COPY --from=builder /api /api
COPY --from=builder /worker /worker

EXPOSE 8080

CMD ["/api"]