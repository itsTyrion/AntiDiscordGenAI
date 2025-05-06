FROM golang:1.24.2-alpine AS builder

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

VOLUME ["/app/config.json"]
ENV CONFIG_PATH=/app

ENTRYPOINT ["./AntiDcGenAI"]