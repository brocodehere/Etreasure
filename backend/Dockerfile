# Simple backend Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o backend ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/backend /app/
# Copy .env file if it exists (for local development)
COPY --from=builder /app/.env* /app/
ENV BACKEND_PORT=8080
EXPOSE 8080
CMD ["/app/backend"]
