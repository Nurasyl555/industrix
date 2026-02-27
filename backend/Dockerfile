FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy dependency files first for cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
WORKDIR /app/cmd/server
RUN go build -o /server .

# Runtime
FROM alpine:latest
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/migrations ./migrations

ENV HTTP_PORT=8080
EXPOSE 8080
CMD ["./server"]
