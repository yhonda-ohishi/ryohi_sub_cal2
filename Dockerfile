# Multi-stage Dockerfile for Ryohi Router Module

# Build stage
FROM golang:1.23.0-alpine AS builder

RUN apk add --no-cache git make gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o router cmd/router/main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 router && adduser -D -u 1000 -G router router

WORKDIR /app

COPY --from=builder /build/router /app/router

USER router

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/router"]
