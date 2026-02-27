FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY backend/pkg backend/pkg
COPY backend/gen/go backend/gen/go
COPY backend/services/gateway backend/services/gateway
COPY backend/services/identity backend/services/identity
RUN go work sync
WORKDIR /app/backend/services/identity
RUN go mod download
RUN go build -o main main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/backend/services/identity/main .
COPY --from=builder /app/backend/services/identity/docs ./docs
ENV HTTP_PORT=8081
ENV GRPC_PORT=50051
EXPOSE 8081 50051
CMD ["./main"]
