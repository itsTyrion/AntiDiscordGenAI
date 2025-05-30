FROM golang:1.24.3-alpine3.21 AS builder

WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

RUN go build -o AntiDcGenAI

FROM alpine:latest

# Install CA certificates for HTTPS support
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/AntiDcGenAI .

ENV CONFIG_PATH=/app/config

# Use unprivileged user and ensure it has the necessary permissions
RUN adduser -D bot && mkdir -p /app/config && chown -R bot:bot /app
USER bot

ENTRYPOINT ["./AntiDcGenAI"]