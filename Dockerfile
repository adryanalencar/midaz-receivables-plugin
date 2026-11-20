# ==============================================================================
# Multi-Stage Dockerfile for Midaz Acquirer Receivables Plugin
# ==============================================================================

# Stage 1: Build binary
FROM golang:1.24-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/plugin-receivables \
    cmd/app/main.go

# Stage 2: Minimal Production Image
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create non-root user
RUN addgroup -S midaz && adduser -S midaz -G midaz
USER midaz

COPY --from=builder /bin/plugin-receivables /app/plugin-receivables

EXPOSE 8082

ENTRYPOINT ["/app/plugin-receivables"]
