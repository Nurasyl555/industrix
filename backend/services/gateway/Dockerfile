FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY backend/pkg backend/pkg
COPY backend/gen/go backend/gen/go
COPY backend/services/gateway backend/services/gateway
RUN go work sync
WORKDIR /app/backend/services/gateway
RUN go build -o main main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/backend/services/gateway/main .
COPY --from=builder /app/backend/services/gateway/docs ./docs
ENV PORT=8080
EXPOSE 8080
CMD ["./main"]
