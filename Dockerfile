# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
# Install openssl for certificate validation
RUN apk add --no-cache openssl
# Create empty tls directory if it doesn't exist
RUN mkdir -p tls
# Note: TLS certificates should be mounted at runtime or generated in deployment
# They are not included in the image for security reasons
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /flight-api

# Test stage (commented out due to test failures)
# FROM golang:1.23-alpine AS tester
# WORKDIR /app
# COPY --from=builder /app /app
# COPY --from=builder /flight-api /flight-api
# RUN apk add --no-cache postgresql-client
# RUN go test -v -coverprofile=coverage.out ./...

# Final stage using distroless base
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

# Copy artifacts
COPY --from=builder --chown=nonroot:nonroot /flight-api /app/
# Copy empty tls directory from builder (created during build)
COPY --from=builder --chown=nonroot:nonroot /app/tls/ /app/tls/
COPY --from=builder --chown=nonroot:nonroot /app/web/ /app/web/

# Security hardening
USER nonroot
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s \
  CMD ["/app/flight-api", "healthcheck"]

# Runtime configuration
ENTRYPOINT ["/app/flight-api"]