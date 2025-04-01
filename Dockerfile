# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
# Validate TLS certificates
RUN if [ ! -f tls/tls.crt ] || [ ! -f tls/tls.key ] || [ ! -f tls/ca.crt ]; then \
        echo "Missing required TLS certificate files"; \
        exit 1; \
    fi
RUN openssl x509 -in tls/tls.crt -noout -dates && \
    openssl x509 -in tls/ca.crt -noout -dates && \
    openssl rsa -in tls/tls.key -check -noout
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /flight-api

# Test stage
FROM golang:1.21-alpine AS tester
WORKDIR /app
COPY --from=builder /app /app
COPY --from=builder /flight-api /flight-api
RUN apk add --no-cache postgresql-client
RUN go test -v -coverprofile=coverage.out ./...

# Final stage using distroless base
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

# Copy artifacts
COPY --from=builder --chown=nonroot:nonroot /flight-api /app/
COPY --from=builder --chown=nonroot:nonroot /app/config/prod.yaml /app/config/
# TLS certificates with secure permissions
COPY --from=builder --chown=nonroot:nonroot /app/tls/ /app/tls/
RUN chmod 600 /app/tls/*.key && \
    chmod 644 /app/tls/*.crt
COPY --from=builder --chown=nonroot:nonroot /app/web/ /app/web/

# Security hardening
USER nonroot
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s \
  CMD ["/app/flight-api", "healthcheck"]

# Runtime configuration
ENTRYPOINT ["/app/flight-api"]
